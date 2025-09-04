package gemini

import (
	"net/http"
	"time"

	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type FlashImageParser struct {
	*image.StreamParser
}

func NewFlashImageParser() *FlashImageParser {
	return &FlashImageParser{
		StreamParser: image.NewStreamParser(&image.MarkdownImageStrategy{}),
	}
}

type FlashImageResponse struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Model      string        `json:"model"`
	StatusCode int           `json:"status_code"`
	RespBody   string        `json:"resp_body"`
	RespAt     time.Time     `json:"resp_at"`
	Duration   time.Duration `json:"duration"`
	URLs       []string      `json:"URLs"`
	Error      error         `json:"error,omitempty"`
	TaskID     int           `json:"task_id"` // 添加TaskID字段
}

func (f *FlashImageResponse) GetSupplier() string {
	return f.Supplier
}
func (f *FlashImageResponse) GetTokenDesc() string {
	return f.TokenDesc
}
func (f *FlashImageResponse) GetModel() string {
	return f.Model
}
func (f *FlashImageResponse) GetStatusCode() int {
	return f.StatusCode
}
func (f *FlashImageResponse) GetRespAt() time.Time {
	return f.RespAt
}
func (f *FlashImageResponse) FailedRespBody() string {
	if f.StatusCode != http.StatusOK {
		return f.RespBody
	}
	return ""
}
func (f *FlashImageResponse) DurationMs() int64 {
	return f.Duration.Milliseconds()
}
func (f *FlashImageResponse) Succeed() bool {
	return len(f.URLs) != 0
}
func (f *FlashImageResponse) GetURLs() []string {
	return f.URLs
}
func (f *FlashImageResponse) GetError() error {
	return f.Error
}

func (f *FlashImageResponse) SetBasicResponse(statusCode int, respBody string, respAt time.Time) {
	f.StatusCode = statusCode
	f.RespBody = respBody
	f.RespAt = respAt
}

func (f *FlashImageResponse) SetURLs(urls []string) {
	f.URLs = urls
}

func (f *FlashImageResponse) SetError(err error) {
	f.Error = err
}

func (f *FlashImageResponse) GetTaskID() int {
	return f.TaskID
}

func (f *FlashImageResponse) SetTaskID(taskID int) {
	f.TaskID = taskID
}
