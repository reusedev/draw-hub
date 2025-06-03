package gpt

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Image4oParser struct{}

func (g *Image4oParser) Parse(resp *http.Response, response image.Response) error {
	realResp := response.(*Image4oResponse)
	realResp.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logs.Logger.Warn().Str("supplier", realResp.Supplier).
			Str("token_desc", realResp.TokenDesc).
			Str("model", realResp.Model).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Dur("duration", realResp.Duration).
			Str("body", string(body)).
			Msg("image request failed")
	}
	realResp.RespBody = string(body)
	realResp.RespAt = time.Now()
	reg := `]\((https?[^)]+)\)`
	pattern, _ := regexp.Compile(reg)
	matches := pattern.FindAllStringSubmatch(string(body), -1)
	if len(matches) > 0 && len(matches[len(matches)-1]) >= 2 {
		url := matches[len(matches)-1][1]
		url = strings.ReplaceAll(url, "\\u0026", "&")
		realResp.URLs = append(realResp.URLs, url)
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
	if resp.StatusCode != http.StatusOK {
		logs.Logger.Warn().Str("supplier", realResp.Supplier).
			Str("token_desc", realResp.TokenDesc).
			Str("model", realResp.Model).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Dur("duration", realResp.Duration).
			Str("body", string(body)).
			Msg("image request failed")
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
func (r *Image4oResponse) GetBase64() []string {
	return []string{}
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
func (r *Image1Response) GetBase64() []string {
	return r.Base64
}
