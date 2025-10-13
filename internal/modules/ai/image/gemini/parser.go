package gemini

import (
	"time"

	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type FlashImageParser struct {
	*image.StreamParser
}

func NewFlashImageParser() *FlashImageParser {
	return &FlashImageParser{
		StreamParser: image.NewStreamParser(&image.MarkdownURLStrategy{}, &image.GenericB64Strategy{}),
	}
}

type FlashImageResponse struct {
	Supplier   string    `json:"supplier"`
	TokenDesc  string    `json:"token_desc"`
	Model      string    `json:"model"`
	StatusCode int       `json:"status_code"`
	RespBody   string    `json:"resp_body"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	ReqAt      time.Time `json:"req_at"`
	RespAt     time.Time `json:"resp_at"`
	URLs       []string  `json:"URLs"`
	B64s       []string  `json:"b64s"`
	Error      error     `json:"error,omitempty"`
	TaskID     int       `json:"task_id"` // 添加TaskID字段
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
func (f *FlashImageResponse) GetRespBody() string {
	return f.RespBody
}
func (f *FlashImageResponse) TaskConsumeMs() int64 {
	return f.EndAt.Sub(f.StartAt).Milliseconds()
}
func (f *FlashImageResponse) ReqConsumeMs() int64 {
	return f.RespAt.Sub(f.ReqAt).Milliseconds()
}
func (f *FlashImageResponse) Succeed() bool {
	return len(f.URLs) != 0 || len(f.B64s) != 0
}
func (f *FlashImageResponse) GetURLs() []string {
	return f.URLs
}
func (f *FlashImageResponse) GetB64s() []string {
	return f.B64s
}
func (f *FlashImageResponse) GetError() error {
	return f.Error
}

func (f *FlashImageResponse) SetBasicResponse(statusCode int, respBody string) {
	f.StatusCode = statusCode
	f.RespBody = respBody
}

func (f *FlashImageResponse) SetStartAt(startAt time.Time) {
	f.StartAt = startAt
}

func (f *FlashImageResponse) SetEndAt(endAt time.Time) {
	f.EndAt = endAt
}

func (f *FlashImageResponse) SetReqAt(reqAt time.Time) {
	f.ReqAt = reqAt
}

func (f *FlashImageResponse) SetRespAt(respAt time.Time) {
	f.RespAt = respAt
}

func (f *FlashImageResponse) SetURLs(urls []string) {
	f.URLs = urls
}

func (f *FlashImageResponse) SetB64s(b64 []string) {
	f.B64s = b64
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
