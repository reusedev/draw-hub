package image

import (
	"github.com/reusedev/draw-hub/internal/modules/ai_model"
	"io"
	"net/http"
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
	return nil
}

type GPT4oImageResponse struct {
	Supplier   string        `json:"supplier"`
	Model      string        `json:"ai_model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	URLs       []string      `json:"URLs"`
}

func (r *GPT4oImageResponse) GetSupplier() string {
	return r.Supplier
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
	Supplier string `json:"supplier"`
	Model    string `json:"ai_model"`
}

func (r *GPTImage1Response) GetSupplier() string {
	return r.Supplier
}
func (r *GPTImage1Response) GetModel() string {
	return r.Model
}
