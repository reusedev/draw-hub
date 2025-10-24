package mj

import (
	"encoding/json"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"io"
	"net/http"
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
	image.BaseResponse
}

type tuziUrlStrategy struct{}

func (t *tuziUrlStrategy) ExtractURLs(body []byte) ([]string, error) {
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

type v3UrlStrategy struct{}

func (v *v3UrlStrategy) ExtractURLs(body []byte) ([]string, error) {
	urls := make([]string, 0)
	var response struct {
		ImageURL string `json:"imageUrl"`
	}
	err := json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	if response.ImageURL != "" {
		urls = append(urls, response.ImageURL)
	}
	return urls, nil
}

type parser struct {
	urlStrategy image.URLParseStrategy
}

func (p parser) Parse(resp *http.Response, response image.Response) error {
	if resp.StatusCode != http.StatusOK {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		response.SetBasicResponse(resp.StatusCode, string(data))
		response.SetError(image.StatusCodeError)
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	response.SetBasicResponse(resp.StatusCode, string(body))
	urls, err := p.urlStrategy.ExtractURLs(body)
	if err != nil {
		return err
	}
	response.SetURLs(urls)
	if !response.Succeed() {
		logs.Logger.Warn().
			Int("task_id", response.GetTaskID()).
			Str("supplier", response.GetSupplier()).
			Str("token_desc", response.GetTokenDesc()).
			Str("model", response.GetModel()).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Int64("req_consume_ms", response.ReqConsumeMs()).
			Str("body", string(body)).
			Msg("image resp error")
		// tuzi v3
		failReason := jsoniter.Get(body, "failReason").ToString()
		if failReason != "" {
			response.SetError(errors.New(failReason))
		}
		// geek
		taskStatus := jsoniter.Get(body, "task_status").ToString()
		if taskStatus == "failed" {
			response.SetError(errors.New(taskStatus))
		}
	}
	return nil
}

type geekGenerateResponse struct {
	image.BaseResponse
}

type geekGenerateURLStrategy struct{}

func (e *geekGenerateURLStrategy) ExtractURLs(body []byte) ([]string, error) {
	var resp struct {
		Model      string `json:"model"`
		Created    int    `json:"created"`
		TaskId     string `json:"task_id"`
		TaskStatus string `json:"task_status"`
		Data       []struct {
			Url string `json:"url"`
		} `json:"data"`
	}
	err := json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	for _, v := range resp.Data {
		ret = append(ret, v.Url)
	}
	return ret, nil
}
