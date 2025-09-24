package image

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net/http"
	"regexp"
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
	GetError() error // is nil if Succeed() return true

	SetBasicResponse(statusCode int, respBody string, respAt time.Time)
	SetURLs(urls []string)
	SetError(err error)
	SetTaskID(taskID int)
}

type SysExitResponse interface {
	GetTaskID() int
}

type Parser[T any] interface {
	Parse(resp *http.Response, response T) error
}

type ParseStrategy interface {
	ExtractURLs(body []byte) ([]string, error)
}

type MarkdownImageStrategy struct{}

func (m *MarkdownImageStrategy) ExtractURLs(body []byte) ([]string, error) {
	var urls []string
	var content string

	// 首先尝试解析JSON格式的聊天完成响应
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := jsoniter.Unmarshal(body, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		content = chatResp.Choices[0].Message.Content
	} else {
		// 如果不是JSON格式，直接使用原始body作为内容
		content = string(body)
	}

	// 尝试提取markdown格式的图片链接
	markdownReg := `!\[.*?\]\((https?://[^)]+)\)`
	pattern, _ := regexp.Compile(markdownReg)
	matches := pattern.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			url := match[1]
			url = strings.ReplaceAll(url, "\\u0026", "&")
			urls = append(urls, url)
		}
	}

	// 尝试解析JSON代码块中的图片链接（无论是否找到markdown图片）
	jsonBlockReg := "```json\\s*\\n([\\s\\S]*?)\\n```"
	jsonPattern, _ := regexp.Compile(jsonBlockReg)
	jsonMatches := jsonPattern.FindAllStringSubmatch(content, -1)
	for _, jsonMatch := range jsonMatches {
		if len(jsonMatch) >= 2 {
			var jsonData struct {
				Image []string `json:"image"`
			}
			if err := jsoniter.Unmarshal([]byte(jsonMatch[1]), &jsonData); err == nil {
				for _, imageURL := range jsonData.Image {
					if imageURL != "" {
						imageURL = strings.ReplaceAll(imageURL, "\\u0026", "&")
						urls = append(urls, imageURL)
					}
				}
			}
		}
	}

	// 如果仍然没有找到URL，尝试原来的正则表达式作为后备
	if len(urls) == 0 {
		reg := `(https?[^)]+)\)`
		pattern, _ := regexp.Compile(reg)
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				url := match[1]
				url = strings.ReplaceAll(url, "\\u0026", "&")
				urls = append(urls, url)
			}
		}
	}

	return urls, nil
}

type OpenAIImageStrategy struct{}

func (o *OpenAIImageStrategy) ExtractURLs(body []byte) ([]string, error) {
	var urls []string
	var s struct {
		Data []struct {
			URL           string `json:"url,omitempty"`
			B64JSON       string `json:"b64_json,omitempty"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		} `json:"data"`
	}
	err := jsoniter.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	for _, v := range s.Data {
		if v.URL != "" {
			urls = append(urls, v.URL)
		}
	}
	return urls, nil
}

type GenericParser struct {
	strategy ParseStrategy
}

func NewGenericParser(strategy ParseStrategy) *GenericParser {
	return &GenericParser{strategy: strategy}
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
	urls, err := g.strategy.ExtractURLs(body)
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
	strategy ParseStrategy
}

func NewStreamParser(strategy ParseStrategy) *StreamParser {
	return &StreamParser{strategy: strategy}
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

	if extractedURLs, err := s.strategy.ExtractURLs([]byte(finalContent)); err == nil {
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
