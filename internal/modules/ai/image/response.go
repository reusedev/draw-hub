package image

import "time"

type SubmitResponse interface {
	GetReqAt() time.Time
	GetRespAt() time.Time
	GetProviderTaskID() int64
	ReqConsumeMs() int64
	GetTaskID() int
	Succeed() bool

	SetBasicResponse(statusCode int, respBody string)
	SetReqAt(reqAt time.Time)
	SetRespAt(respAt time.Time)
	SetProviderTaskID(id int64)
	SetTaskID(id int)
}

type Response interface {
	GetModel() string
	GetSupplier() string
	GetTokenDesc() string
	GetStatusCode() int
	GetRespAt() time.Time
	GetRespBody() string
	TaskConsumeMs() int64
	ReqConsumeMs() int64
	GetTaskID() int
	Succeed() bool
	GetURLs() []string
	GetB64s() []string
	GetError() error // is nil if Succeed() return true

	SetBasicResponse(statusCode int, respBody string)
	SetStartAt(startAt time.Time)
	SetEndAt(endAt time.Time)
	SetReqAt(reqAt time.Time)
	SetRespAt(respAt time.Time)
	SetURLs(urls []string)
	SetB64s(b64 []string)
	SetError(err error)
	SetTaskID(taskID int)
}

type SysExitResponse interface {
	GetTaskID() int
}

type BaseResponse struct {
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
	TaskID     int       `json:"task_id"`
}

func (r *BaseResponse) GetSupplier() string  { return r.Supplier }
func (r *BaseResponse) GetTokenDesc() string { return r.TokenDesc }
func (r *BaseResponse) GetModel() string     { return r.Model }
func (r *BaseResponse) GetStatusCode() int   { return r.StatusCode }
func (r *BaseResponse) GetRespAt() time.Time { return r.RespAt }
func (r *BaseResponse) GetRespBody() string  { return r.RespBody }
func (r *BaseResponse) GetTaskID() int       { return r.TaskID }
func (r *BaseResponse) GetURLs() []string    { return r.URLs }
func (r *BaseResponse) GetB64s() []string    { return r.B64s }
func (r *BaseResponse) GetError() error      { return r.Error }
func (r *BaseResponse) Succeed() bool        { return len(r.URLs) != 0 || len(r.B64s) != 0 }
func (r *BaseResponse) TaskConsumeMs() int64 { return r.EndAt.Sub(r.StartAt).Milliseconds() }
func (r *BaseResponse) ReqConsumeMs() int64  { return r.RespAt.Sub(r.ReqAt).Milliseconds() }

func (r *BaseResponse) SetBasicResponse(statusCode int, respBody string) {
	r.StatusCode = statusCode
	r.RespBody = respBody
}
func (r *BaseResponse) SetStartAt(startAt time.Time) { r.StartAt = startAt }
func (r *BaseResponse) SetEndAt(endAt time.Time)     { r.EndAt = endAt }
func (r *BaseResponse) SetReqAt(reqAt time.Time)     { r.ReqAt = reqAt }
func (r *BaseResponse) SetRespAt(respAt time.Time)   { r.RespAt = respAt }
func (r *BaseResponse) SetURLs(urls []string)        { r.URLs = urls }
func (r *BaseResponse) SetB64s(b64 []string)         { r.B64s = b64 }
func (r *BaseResponse) SetError(err error)           { r.Error = err }
func (r *BaseResponse) SetTaskID(taskID int)         { r.TaskID = taskID }
