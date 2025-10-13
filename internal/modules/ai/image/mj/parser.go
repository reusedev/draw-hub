package mj

import (
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
	"time"
)

type ImagineResponse struct {
	Supplier       string    `json:"supplier"`
	TokenDesc      string    `json:"token_desc"`
	RespBody       string    `json:"resp_body"`
	StatusCode     int       `json:"status_code"`
	ReqAt          time.Time `json:"req_at"`
	RespAt         time.Time `json:"resp_at"`
	TaskID         int       `json:"task_id"`
	ProviderTaskID int64     `json:"provider_task_id"`
}

func (i *ImagineResponse) GetProviderTaskID() int64 {
	return i.ProviderTaskID
}
func (i *ImagineResponse) ReqConsumeMs() int64 {
	return i.RespAt.Sub(i.ReqAt).Milliseconds()
}
func (i *ImagineResponse) GetReqAt() time.Time {
	return i.ReqAt
}
func (i *ImagineResponse) GetRespAt() time.Time {
	return i.RespAt
}
func (i *ImagineResponse) GetTaskID() int {
	return i.TaskID
}
func (i *ImagineResponse) Succeed() bool {
	return jsoniter.Get([]byte(i.RespBody), "code").ToInt() == 1
}

func (i *ImagineResponse) SetBasicResponse(statusCode int, respBody string) {
	i.StatusCode = statusCode
	i.RespBody = respBody
}

func (i *ImagineResponse) SetReqAt(reqAt time.Time) {
	i.ReqAt = reqAt
}

func (i *ImagineResponse) SetRespAt(respAt time.Time) {
	i.RespAt = respAt
}

func (i *ImagineResponse) SetProviderTaskID(id int64) {
	i.ProviderTaskID = id
}

func (i *ImagineResponse) SetTaskID(id int) {
	i.TaskID = id
}

type providerTaskIDStrategy struct{}

func (p *providerTaskIDStrategy) ExtractProviderTaskID(body []byte) (int64, error) {
	return jsoniter.Get(body, "result").ToInt64(), nil
}

type FetchResponse struct {
	Supplier   string    `json:"supplier"`
	TokenDesc  string    `json:"token_desc"`
	Model      string    `json:"model"`
	StatusCode int       `json:"status_code"`
	RespBody   string    `json:"resp_body"`
	StartAt    time.Time `json:"start_at"`
	EndAt      time.Time `json:"end_at"`
	ReqAt      time.Time `json:"req_at"`
	RespAt     time.Time `json:"resp_at"`
	Error      error     `json:"error,omitempty"`
	TaskID     int       `json:"task_id"`
	URLs       []string  `json:"urls"`
}

func (f *FetchResponse) GetSupplier() string {
	return f.Supplier
}
func (f *FetchResponse) GetTokenDesc() string {
	return f.TokenDesc
}
func (f *FetchResponse) GetModel() string {
	return f.Model
}
func (f *FetchResponse) GetStatusCode() int {
	return f.StatusCode
}
func (f *FetchResponse) GetRespAt() time.Time {
	return f.RespAt
}
func (f *FetchResponse) GetRespBody() string {
	return f.RespBody
}
func (f *FetchResponse) TaskConsumeMs() int64 {
	return f.EndAt.Sub(f.StartAt).Milliseconds()
}
func (f *FetchResponse) ReqConsumeMs() int64 {
	return f.RespAt.Sub(f.ReqAt).Milliseconds()
}
func (f *FetchResponse) Succeed() bool {
	return len(f.URLs) != 0
}
func (f *FetchResponse) GetURLs() []string {
	return f.URLs
}
func (f *FetchResponse) GetB64s() []string {
	return nil
}
func (f *FetchResponse) GetError() error {
	return f.Error
}

func (f *FetchResponse) SetBasicResponse(statusCode int, respBody string) {
	f.StatusCode = statusCode
	f.RespBody = respBody
}

func (f *FetchResponse) SetStartAt(startAt time.Time) {
	f.StartAt = startAt
}

func (f *FetchResponse) SetEndAt(endAt time.Time) {
	f.EndAt = endAt
}

func (f *FetchResponse) SetReqAt(reqAt time.Time) {
	f.ReqAt = reqAt
}

func (f *FetchResponse) SetRespAt(respAt time.Time) {
	f.RespAt = respAt
}

func (f *FetchResponse) SetURLs(urls []string) {
	f.URLs = urls
}

func (f *FetchResponse) SetB64s(b64 []string) {}

func (f *FetchResponse) SetError(err error) {
	f.Error = err
}

func (f *FetchResponse) GetTaskID() int {
	return f.TaskID
}

func (f *FetchResponse) SetTaskID(taskID int) {
	f.TaskID = taskID
}

type urlStrategy struct{}

func (u *urlStrategy) ExtractURLs(body []byte) ([]string, error) {
	urls := make([]string, 0)
	var response struct {
		ImageURLs []struct {
			URL string `json:"url"`
		} `json:"imageUrls"`
	}
	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	for _, url := range response.ImageURLs {
		urls = append(urls, url.URL)
	}
	return urls, nil
}
