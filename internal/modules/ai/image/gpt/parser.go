package gpt

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type Image4oParser struct{}

func (g *Image4oParser) Parse(resp *http.Response, response image.Response) error {
	realResp := response.(*Image4oResponse)
	realResp.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	realResp.RespBody = string(body)
	realResp.RespAt = time.Now()

	// 首先尝试解析JSON格式的聊天完成响应
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := jsoniter.Unmarshal(body, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		// 从Markdown格式中提取图片URL
		content := chatResp.Choices[0].Message.Content
		markdownReg := `!\[.*?\]\((https?://[^)]+)\)`
		pattern, _ := regexp.Compile(markdownReg)
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				url := match[1]
				url = strings.ReplaceAll(url, "\\u0026", "&")
				realResp.URLs = append(realResp.URLs, url)
			}
		}
	} else {
		// 如果JSON解析失败，尝试原来的正则表达式
		reg := `(https?[^)]+)\)\\n\\n\[点击下载\]`
		pattern, _ := regexp.Compile(reg)
		matches := pattern.FindAllStringSubmatch(string(body), -1)
		for _, match := range matches {
			if len(match) >= 2 {
				url := match[1]
				url = strings.ReplaceAll(url, "\\u0026", "&")
				realResp.URLs = append(realResp.URLs, url)
			}
		}
	}
	if !realResp.Succeed() {
		logs.Logger.Warn().Str("supplier", realResp.Supplier).
			Str("token_desc", realResp.TokenDesc).
			Str("model", realResp.Model).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Dur("duration", realResp.Duration).
			Str("body", string(body)).
			Msg("image resp error")
		err := detectError(realResp.Supplier, realResp.Model, realResp.RespBody)
		if err != nil {
			realResp.Error = err
		}
	}
	return nil
}

type Image1Parser struct{}

func (g *Image1Parser) Parse(resp *http.Response, response image.Response) error {
	realResp := response.(*Image1Response)
	realResp.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	realResp.RespBody = string(body)
	realResp.RespAt = time.Now()
	var s struct {
		Data []struct {
			URL           string `json:"url,omitempty"`
			B64JSON       string `json:"b64_json,omitempty"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		} `json:"data"`
	}
	err = jsoniter.Unmarshal(body, &s)
	if err != nil {
		return err
	}
	for _, v := range s.Data {
		realResp.URLs = append(realResp.URLs, v.URL)
		realResp.Base64 = append(realResp.Base64, v.B64JSON)
	}
	if !realResp.Succeed() {
		logs.Logger.Warn().Str("supplier", realResp.Supplier).
			Str("token_desc", realResp.TokenDesc).
			Str("model", realResp.Model).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Dur("duration", realResp.Duration).
			Str("body", string(body)).
			Msg("image resp error")
		err := detectError(realResp.Supplier, realResp.Model, realResp.RespBody)
		if err != nil {
			realResp.Error = err
		}
	}
	return nil
}

type Image4oResponse struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	URLs       []string      `json:"URLs"`
	Error      error         `json:"error,omitempty"`
}

func (r *Image4oResponse) GetSupplier() string {
	return r.Supplier
}
func (r *Image4oResponse) GetTokenDesc() string {
	return r.TokenDesc
}
func (r *Image4oResponse) GetModel() string {
	return r.Model
}
func (r *Image4oResponse) GetStatusCode() int {
	return r.StatusCode
}
func (r *Image4oResponse) GetRespAt() time.Time {
	return r.RespAt
}
func (r *Image4oResponse) FailedRespBody() string {
	if r.StatusCode != http.StatusOK {
		return r.RespBody
	}
	return ""
}
func (r *Image4oResponse) DurationMs() int64 {
	return r.Duration.Milliseconds()
}
func (r *Image4oResponse) Succeed() bool {
	return len(r.URLs) != 0
}
func (r *Image4oResponse) GetURLs() []string {
	return r.URLs
}
func (r *Image4oResponse) GetError() error {
	return r.Error
}

type Image1Response struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	Base64     []string      `json:"base64"`
	URLs       []string      `json:"URLs"`
	Error      error         `json:"error,omitempty"`
}

func (r *Image1Response) GetSupplier() string {
	return r.Supplier
}
func (r *Image1Response) GetTokenDesc() string {
	return r.TokenDesc
}
func (r *Image1Response) GetModel() string {
	return r.Model
}
func (r *Image1Response) GetStatusCode() int {
	return r.StatusCode
}
func (r *Image1Response) GetRespAt() time.Time {
	return r.RespAt
}
func (r *Image1Response) FailedRespBody() string {
	if r.StatusCode != http.StatusOK {
		return r.RespBody
	}
	return ""
}
func (r *Image1Response) DurationMs() int64 {
	return r.Duration.Milliseconds()
}
func (r *Image1Response) Succeed() bool {
	return len(r.Base64) != 0 || len(r.URLs) != 0
}
func (r *Image1Response) GetURLs() []string {
	return r.URLs
}
func (r *Image1Response) GetError() error {
	return r.Error
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

func detectError(supplier, model, body string) error {
	if errs, ok := errorMap[supplier+model]; ok {
		for key, err := range errs {
			if strings.Contains(body, key) {
				return err
			}
		}
	}
	return nil
}
