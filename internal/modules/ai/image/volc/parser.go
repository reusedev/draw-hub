package volc

import (
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"time"
)

type JiMengParser struct {
	*image.GenericParser
}

func NewJiMengParser() *JiMengParser {
	return &JiMengParser{
		GenericParser: image.NewGenericParser(&JiMengParserStrategy{}),
	}
}

type JiMengParserStrategy struct{}

func (j *JiMengParserStrategy) ExtractURLs(body []byte) ([]string, error) {
	var responseBody struct {
		Data []struct {
			Url string `json:"url"`
		} `json:"data"`
		Created int `json:"created"`
		Usage   struct {
			PromptTokens        int `json:"prompt_tokens"`
			CompletionTokens    int `json:"completion_tokens"`
			TotalTokens         int `json:"total_tokens"`
			PromptTokensDetails struct {
				CachedTokensDetails struct {
				} `json:"cached_tokens_details"`
			} `json:"prompt_tokens_details"`
			CompletionTokensDetails struct {
			} `json:"completion_tokens_details"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	for _, data := range responseBody.Data {
		if data.Url != "" {
			ret = append(ret, data.Url)
		}
	}
	return ret, nil
}

type CreateResponse struct {
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

func (r *CreateResponse) GetSupplier() string {
	return r.Supplier
}
func (r *CreateResponse) GetTokenDesc() string {
	return r.TokenDesc
}
func (r *CreateResponse) GetModel() string {
	return r.Model
}
func (r *CreateResponse) GetStatusCode() int {
	return r.StatusCode
}
func (r *CreateResponse) GetRespAt() time.Time {
	return r.RespAt
}
func (r *CreateResponse) GetRespBody() string {
	return r.RespBody
}
func (r *CreateResponse) TaskConsumeMs() int64 {
	return r.EndAt.Sub(r.StartAt).Milliseconds()
}
func (r *CreateResponse) ReqConsumeMs() int64 {
	return r.RespAt.Sub(r.ReqAt).Milliseconds()
}
func (r *CreateResponse) Succeed() bool {
	return len(r.URLs) != 0
}
func (r *CreateResponse) GetURLs() []string {
	return r.URLs
}
func (r *CreateResponse) GetB64s() []string {
	return nil
}
func (r *CreateResponse) GetError() error {
	return r.Error
}

func (r *CreateResponse) SetBasicResponse(statusCode int, respBody string) {
	r.StatusCode = statusCode
	r.RespBody = respBody
}

func (r *CreateResponse) SetStartAt(startAt time.Time) {
	r.StartAt = startAt
}

func (r *CreateResponse) SetEndAt(endAt time.Time) {
	r.EndAt = endAt
}

func (r *CreateResponse) SetReqAt(reqAt time.Time) {
	r.ReqAt = reqAt
}

func (r *CreateResponse) SetRespAt(respAt time.Time) {
	r.RespAt = respAt
}

func (r *CreateResponse) SetURLs(urls []string) {
	r.URLs = urls
}

func (r *CreateResponse) SetB64s(b64 []string) {}

func (r *CreateResponse) SetError(err error) {
	r.Error = err
}

func (r *CreateResponse) GetTaskID() int {
	return r.TaskID
}

func (r *CreateResponse) SetTaskID(taskID int) {
	r.TaskID = taskID
}
