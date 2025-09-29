package image

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type Response interface {
	GetModel() string
	GetSupplier() string
	GetTokenDesc() string
	GetStatusCode() int
	GetRespAt() time.Time
	GetRespBody() string
	DurationMs() int64
	GetTaskID() int

	Succeed() bool
	GetURLs() []string
	GetB64s() []string
	GetError() error // is nil if Succeed() return true

	SetBasicResponse(statusCode int, respBody string, respAt time.Time)
	SetURLs(urls []string)
	SetB64s(b64 []string)
	SetError(err error)
	SetTaskID(taskID int)
}

type SysExitResponse interface {
	GetTaskID() int
}

type Parser[T any] interface {
	Parse(resp *http.Response, response T) error
}

type GenericParser struct {
	urlStrategy URLParseStrategy
}

func NewGenericParser(urlStrategy URLParseStrategy) *GenericParser {
	return &GenericParser{urlStrategy: urlStrategy}
}

func (g *GenericParser) Parse(resp *http.Response, response Response) error {
	if resp.StatusCode != http.StatusOK {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*90)
		defer cancel()
		type result struct {
			data []byte
			err  error
		}
		resultCh := make(chan result, 1)
		go func() {
			data, err := io.ReadAll(resp.Body)
			resultCh <- result{data: data, err: err}
		}()
		var respBody []byte
		select {
		case res := <-resultCh:
			if res.err != nil {
				return res.err
			}
			respBody = res.data
		case <-ctx.Done():
		}
		// Read body with timeout, because sometime it will block about 900s.
		response.SetBasicResponse(resp.StatusCode, string(respBody), time.Now())
		if detectedErr := DetectError(response, string(respBody)); detectedErr != nil {
			response.SetError(detectedErr)
		}
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response.SetBasicResponse(resp.StatusCode, string(body), time.Now())
	urls, err := g.urlStrategy.ExtractURLs(body)
	if err != nil {
		return err
	}
	response.SetURLs(urls)
	if !response.Succeed() {
		logs.Logger.Warn().
			Int("task_id", response.GetTaskID()).
			Str("supplier", response.GetSupplier()).
			Str("token_desc", response.GetTokenDesc()).
			Str("model", response.GetModel()).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Int64("duration", response.DurationMs()).
			Str("body", string(body)).
			Msg("image resp error")
		if detectedErr := DetectError(response, string(body)); detectedErr != nil {
			response.SetError(detectedErr)
		}
	}
	return nil
}

type StreamParser struct {
	urlStrategy URLParseStrategy
	b64Strategy B64ParseStrategy
}

func NewStreamParser(urlStrategy URLParseStrategy, b64Strategy B64ParseStrategy) *StreamParser {
	return &StreamParser{urlStrategy: urlStrategy, b64Strategy: b64Strategy}
}

func (s *StreamParser) extractContent(chunk []byte) []byte {
	var chatResp struct {
		Choices []struct {
			Delta struct {
				Content string `json:"content"`
			} `json:"delta"`
		} `json:"choices"`
	}
	if err := jsoniter.Unmarshal(chunk, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		return []byte(chatResp.Choices[0].Delta.Content)
	} else if err != nil {
		// Log parsing errors for debugging (only in development)
		logs.Logger.Info().
			Err(err).
			Str("chunk", string(chunk)).
			Msg("Failed to parse SSE chunk")
	}
	return nil
}

func (s *StreamParser) Parse(resp *http.Response, response Response) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*90)
		defer cancel()
		type result struct {
			data []byte
			err  error
		}
		resultCh := make(chan result, 1)
		go func() {
			data, err := io.ReadAll(resp.Body)
			resultCh <- result{data: data, err: err}
		}()
		var respBody []byte
		select {
		case res := <-resultCh:
			if res.err != nil {
				return res.err
			}
			respBody = res.data
		case <-ctx.Done():
		}
		// Read body with timeout, because sometime it will block about 900s.
		response.SetBasicResponse(resp.StatusCode, string(respBody), time.Now())
		if detectedErr := DetectError(response, string(respBody)); detectedErr != nil {
			response.SetError(detectedErr)
		}
		return nil
	}
	var content strings.Builder
	var totalChunks int

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 50*1024*1024) // Increase buffer size for large chunks

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Process SSE data lines
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")
			dataStr = strings.TrimSpace(dataStr)

			// Skip [DONE] marker
			if dataStr == "[DONE]" {
				logs.Logger.Info().Msg("StreamParser: Received [DONE] marker")
				break
			}

			// Try to parse the JSON chunk
			chunk := s.extractContent([]byte(dataStr))
			if chunk != nil {
				content.Write(chunk)
				totalChunks++
				// 不在流式过程中提取URL，等流结束后统一提取
			}
		}
		// Ignore other SSE fields like event:, id:, retry:, etc.
	}

	// 流结束后，从完整内容中提取URL
	var urls []string
	finalContent := content.String()

	// 记录最终的完整内容
	logs.Logger.Info().
		Str("final_content", finalContent).
		Msg("StreamParser: Final accumulated content")

	if extractedURLs, err := s.urlStrategy.ExtractURLs([]byte(finalContent)); err == nil {
		// 去重
		urlSet := make(map[string]struct{})
		for _, url := range extractedURLs {
			if _, exists := urlSet[url]; !exists {
				urlSet[url] = struct{}{}
				urls = append(urls, url)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		logs.Logger.Error().Err(err).Msg("Error reading SSE stream")
		return err
	}
	bodyString := content.String()
	response.SetBasicResponse(resp.StatusCode, bodyString, time.Now())
	response.SetURLs(urls)

	b64s, err := s.b64Strategy.ExtractB64s([]byte(finalContent))
	if err != nil {
		logs.Logger.Error().Err(err).Msg("Extract b64s error")
	}
	response.SetB64s(b64s)

	if !response.Succeed() {
		logs.Logger.Warn().
			Int("task_id", response.GetTaskID()).
			Str("supplier", response.GetSupplier()).
			Str("token_desc", response.GetTokenDesc()).
			Str("model", response.GetModel()).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Int64("duration", response.DurationMs()).
			Str("body", bodyString).
			Msg("stream image resp error")
		if detectedErr := DetectError(response, bodyString); detectedErr != nil {
			response.SetError(detectedErr)
		}
	}
	return nil
}

var (
	PromptError     = errors.New("图片检测系统认为内容可能违反相关政策")
	NoImageError    = errors.New("未提取到图片")
	StatusCodeError = errors.New("http状态码非200")
)

var (
	errorMap = map[string]map[string]error{
		consts.Tuzi.String() + consts.GPT4oImage.String(): {
			"图片检测系统认为内容可能违反相关政策": PromptError,
		},
		consts.Tuzi.String() + consts.GPT4oImageVip.String(): {
			"图片检测系统认为内容可能违反相关政策": PromptError,
		},
		consts.Geek.String() + consts.GPT4oImage.String(): {
			"图片检测系统认为内容可能违反相关政策": PromptError,
		},
		consts.V3.String() + consts.GPT4oImageVip.String(): {
			"该任务的输入或者输出可能违反了OpenAI的相关服务政策，请重新发起请求或调整提示词进行重试": PromptError,
		},
		consts.Geek.String() + consts.GPTImage1.String(): {
			"Your request may contain content that is not allowed by our safety system. Please try change the prompt and image.": PromptError,
		},
	}
)

func DetectError(response Response, body string) error {
	if response.Succeed() {
		return nil
	}
	if errs, ok := errorMap[response.GetSupplier()+response.GetModel()]; ok {
		for key, err := range errs {
			if strings.Contains(body, key) {
				return err
			}
		}
	}
	if response.GetStatusCode() != http.StatusOK {
		return StatusCodeError
	}
	if len(response.GetURLs()) == 0 {
		return NoImageError
	}
	return nil
}

type GenericSysExitResponse struct {
	TaskID int
}

func (g *GenericSysExitResponse) GetTaskID() int {
	return g.TaskID
}

func ShouldBanToken(response Response) bool {
	c := response.GetStatusCode()
	if c >= 500 && c < 600 {
		return true
	}
	return false
}
