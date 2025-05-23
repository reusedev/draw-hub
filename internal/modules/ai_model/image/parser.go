package image

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/modules/ai_model"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type GPT4oImageParser struct{}

func (g *GPT4oImageParser) Parse(resp *http.Response, response ai_model.Response) error {
	response.(*GPT4oImageResponse).StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response.(*GPT4oImageResponse).RespBody = string(body)
	response.(*GPT4oImageResponse).RespAt = time.Now()
	reg := `]\((https?[^)]+)\)`
	pattern, _ := regexp.Compile(reg)
	matches := pattern.FindAllStringSubmatch(string(body), -1)
	if len(matches) > 0 && len(matches[len(matches)-1]) >= 2 {
		url := matches[len(matches)-1][1]
		url = strings.ReplaceAll(url, "\\u0026", "&")
		response.(*GPT4oImageResponse).URLs = append(response.(*GPT4oImageResponse).URLs, url)
	}
	return nil
}

type GPTImage1Parser struct{}

func (g *GPTImage1Parser) Parse(resp *http.Response, response ai_model.Response) error {
	response.(*GPTImage1Response).StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response.(*GPTImage1Response).RespBody = string(body)
	response.(*GPTImage1Response).RespAt = time.Now()
	var s struct {
		Data []struct {
			B64Json string `json:"b64_json"`
		} `json:"data"`
	}
	err = jsoniter.Unmarshal(body, &s)
	if err != nil {
		return err
	}
	if len(s.Data) > 0 && len(s.Data[0].B64Json) != 0 {
		response.(*GPTImage1Response).Base64 = s.Data[0].B64Json
	}
	return nil
}

type GPT4oImageResponse struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	URLs       []string      `json:"URLs"`
}

func (r *GPT4oImageResponse) GetSupplier() string {
	return r.Supplier
}
func (r *GPT4oImageResponse) GetTokenDesc() string {
	return r.TokenDesc
}
func (r *GPT4oImageResponse) GetModel() string {
	return r.Model
}
func (r *GPT4oImageResponse) GetStatusCode() int {
	return r.StatusCode
}
func (r *GPT4oImageResponse) GetRespAt() time.Time {
	return r.RespAt
}
func (r *GPT4oImageResponse) FailedRespBody() string {
	if r.StatusCode != http.StatusOK {
		return r.RespBody
	}
	return ""
}
func (r *GPT4oImageResponse) DurationMs() int64 {
	return r.Duration.Milliseconds()
}
func (r *GPT4oImageResponse) Succeed() bool {
	return len(r.URLs) != 0
}
func (r *GPT4oImageResponse) GetURLs() []string {
	return r.URLs
}
func (r *GPT4oImageResponse) GetBase64() string {
	return ""
}

type GPTImage1Response struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	Base64     string        `json:"base64"`
}

func (r *GPTImage1Response) GetSupplier() string {
	return r.Supplier
}
func (r *GPTImage1Response) GetTokenDesc() string {
	return r.TokenDesc
}
func (r *GPTImage1Response) GetModel() string {
	return r.Model
}
func (r *GPTImage1Response) GetStatusCode() int {
	return r.StatusCode
}
func (r *GPTImage1Response) GetRespAt() time.Time {
	return r.RespAt
}
func (r *GPTImage1Response) FailedRespBody() string {
	if r.StatusCode != http.StatusOK {
		return r.RespBody
	}
	return ""
}
func (r *GPTImage1Response) DurationMs() int64 {
	return r.Duration.Milliseconds()
}
func (r *GPTImage1Response) Succeed() bool {
	return len(r.Base64) != 0
}
func (r *GPTImage1Response) GetURLs() []string {
	return []string{}
}
func (r *GPTImage1Response) GetBase64() string {
	return r.Base64
}
