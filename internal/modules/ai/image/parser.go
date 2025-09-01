package image

import (
	"bytes"
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
	FailedRespBody() string // != 200
	DurationMs() int64

	Succeed() bool
	GetURLs() []string
	GetError() error

	SetBasicResponse(statusCode int, respBody string, respAt time.Time)
	SetURLs(urls []string)
	SetError(err error)
}

type Parser interface {
	Parse(resp *http.Response, response Response) error
}

type ParseStrategy interface {
	ExtractURLs(body []byte) ([]string, error)
	ValidateResponse(response Response) bool
}

type MarkdownImageStrategy struct{}

func (m *MarkdownImageStrategy) ExtractURLs(body []byte) ([]string, error) {
	var urls []string
	// 首先尝试解析JSON格式的聊天完成响应
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := jsoniter.Unmarshal(body, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		content := chatResp.Choices[0].Message.Content
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
	} else {
		// 如果JSON解析失败，尝试原来的正则表达式
		reg := `(https?[^)]+)\)`
		pattern, _ := regexp.Compile(reg)
		matches := pattern.FindAllStringSubmatch(string(body), -1)
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

func (m *MarkdownImageStrategy) ValidateResponse(response Response) bool {
	return len(response.GetURLs()) > 0
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

func (o *OpenAIImageStrategy) ValidateResponse(response Response) bool {
	return len(response.GetURLs()) > 0
}

type GenericParser struct {
	strategy ParseStrategy
}

func NewGenericParser(strategy ParseStrategy) *GenericParser {
	return &GenericParser{strategy: strategy}
}

func (g *GenericParser) Parse(resp *http.Response, response Response) error {
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
	if !g.strategy.ValidateResponse(response) {
		logs.Logger.Warn().Str("supplier", response.GetSupplier()).
			Str("token_desc", response.GetTokenDesc()).
			Str("model", response.GetModel()).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Int64("duration", response.DurationMs()).
			Str("body", string(body)).
			Msg("image resp error")
		if detectedErr := DetectError(response.GetSupplier(), response.GetModel(), string(body)); detectedErr != nil {
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
	}
	return nil
}

func (s *StreamParser) Parse(resp *http.Response, response Response) error {
	defer resp.Body.Close()
	var content strings.Builder
	buffer := make([]byte, 4096)
	var urls []string
	urlSet := make(map[string]struct{})

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			chunk := s.extractContent(bytes.TrimPrefix(buffer[:n], []byte("data: ")))
			content.Write(chunk)
			us, err := s.strategy.ExtractURLs([]byte(content.String()))
			if err != nil {
				return err
			}
			for _, url := range us {
				if _, exists := urlSet[url]; !exists {
					urlSet[url] = struct{}{}
					urls = append(urls, url)
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	bodyString := content.String()
	response.SetBasicResponse(resp.StatusCode, bodyString, time.Now())
	response.SetURLs(urls)
	if !s.strategy.ValidateResponse(response) {
		logs.Logger.Warn().Str("supplier", response.GetSupplier()).
			Str("token_desc", response.GetTokenDesc()).
			Str("model", response.GetModel()).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Int64("duration", response.DurationMs()).
			Str("body", bodyString).
			Msg("stream image resp error")
		if detectedErr := DetectError(response.GetSupplier(), response.GetModel(), bodyString); detectedErr != nil {
			response.SetError(detectedErr)
		}
	}
	return nil
}

var (
	PromptError = errors.New("system judge that the prompt is not suitable for image generation, please try again with a different prompt")
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
	}
)

func DetectError(supplier, model, body string) error {
	if errs, ok := errorMap[supplier+model]; ok {
		for key, err := range errs {
			if strings.Contains(body, key) {
				return err
			}
		}
	}
	return nil
}
