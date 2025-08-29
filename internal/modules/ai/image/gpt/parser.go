package gpt

import (
	"net/http"
	"time"

	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type Image4oParser struct {
	*image.GenericParser
}

func NewImage4oParser() *Image4oParser {
	return &Image4oParser{
		GenericParser: image.NewGenericParser(&image.MarkdownImageStrategy{}),
	}
}

type Image1Parser struct {
	*image.GenericParser
}

func NewImage1Parser() *Image1Parser {
	return &Image1Parser{
		GenericParser: image.NewGenericParser(&image.OpenAIImageStrategy{}),
	}
}

type OpenAIWithBase64Strategy struct {
	*image.OpenAIImageStrategy
}

func (o *OpenAIWithBase64Strategy) ExtractURLs(body []byte) ([]string, error) {
	urls, err := o.OpenAIImageStrategy.ExtractURLs(body)
	if err != nil {
		return nil, err
	}
	return urls, nil
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

func (r *Image4oResponse) SetBasicResponse(statusCode int, respBody string, respAt time.Time) {
	r.StatusCode = statusCode
	r.RespBody = respBody
	r.RespAt = respAt
}

func (r *Image4oResponse) SetURLs(urls []string) {
	r.URLs = urls
}

func (r *Image4oResponse) SetError(err error) {
	r.Error = err
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

func (r *Image1Response) SetBasicResponse(statusCode int, respBody string, respAt time.Time) {
	r.StatusCode = statusCode
	r.RespBody = respBody
	r.RespAt = respAt
}

func (r *Image1Response) SetURLs(urls []string) {
	r.URLs = urls
}

func (r *Image1Response) SetError(err error) {
	r.Error = err
}
