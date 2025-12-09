package image

import (
	"fmt"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/http_client"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"strings"
	"time"
)

type SyncRequester struct {
	token   ai.Token
	Request Request[Response]
	Parser  Parser[Response]
	TaskID  int // 添加TaskID字段用于日志跟踪
}

func NewRequester(token ai.Token, requestTypes Request[Response], parser Parser[Response]) *SyncRequester {
	return &SyncRequester{
		token:   token,
		Request: requestTypes,
		Parser:  parser,
		TaskID:  0, // 默认值，需要调用方设置
	}
}

func (r *SyncRequester) SetTaskID(taskID int) *SyncRequester {
	r.TaskID = taskID
	return r
}

func (r *SyncRequester) Do() (Response, error) {
	retryTimes := 0
retry:
	client := http_client.NewWithTimeout(20 * time.Minute)
	body, contentType, err := r.Request.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.Request.Path(r.token.Supplier)),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	reqAt := time.Now()
	resp, err := client.Do(req)
	respAt := time.Now()
	if err != nil {
		// tuzi 收到请求后，长时间未响应也未计费，导致任务一直running
		if strings.Contains(err.Error(), "Client.Timeout") {
			if retryTimes < 2 {
				logs.Logger.Info().
					Int("task_id", r.TaskID).
					Str("supplier", r.token.Supplier.String()).
					Str("token_desc", r.token.Desc).
					Str("path", r.Request.Path(r.token.Supplier)).
					Str("method", req.Method).
					Msg("Client.Timeout retry")
				retryTimes++
				goto retry
			}
		}
		return nil, err
	}
	defer resp.Body.Close()
	logs.Logger.Info().
		Int("task_id", r.TaskID).
		Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.Request.Path(r.token.Supplier)).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("req_consume_ms", respAt.Sub(reqAt)).
		Msg("image request")
	ret := r.Request.InitResponse(r.token.Supplier.String(), r.token.Desc)
	ret.SetTaskID(r.TaskID)
	ret.SetReqAt(reqAt)
	ret.SetRespAt(respAt)
	ret.SetStartAt(reqAt)
	ret.SetEndAt(respAt)
	err = r.Parser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type AsyncRequester struct {
	token           ai.Token
	SubmitRequest   Request[SubmitResponse]
	SubmitParser    Parser[SubmitResponse]
	PollingRequest  Request[Response]
	PollingParser   Parser[Response]
	OnSubmitSucceed func(response SubmitResponse)
	TaskID          int // 添加TaskID字段用于日志跟踪
}

func NewAsyncRequester(
	token ai.Token, submitRequest Request[SubmitResponse], submitParser Parser[SubmitResponse],
	pollingRequest Request[Response], pollingParser Parser[Response], onsubmitSucceed func(response SubmitResponse),
) *AsyncRequester {
	return &AsyncRequester{
		token:           token,
		SubmitRequest:   submitRequest,
		SubmitParser:    submitParser,
		PollingRequest:  pollingRequest,
		PollingParser:   pollingParser,
		OnSubmitSucceed: onsubmitSucceed,
		TaskID:          0, // 默认值，需要调用方设置
	}
}

func (r *AsyncRequester) SetTaskID(taskID int) *AsyncRequester {
	r.TaskID = taskID
	return r
}

func (r *AsyncRequester) Do() (Response, error) {
	submitRet, err := r.submit()
	if err != nil {
		return nil, err
	}
	if !submitRet.Succeed() {
		return nil, fmt.Errorf("submit task failed")
	}
	r.OnSubmitSucceed(submitRet)

	for {
		pollingRet, err := r.polling()
		if err != nil {
			return nil, err
		}
		if pollingRet.Succeed() {
			pollingRet.SetStartAt(submitRet.GetReqAt())
			pollingRet.SetEndAt(pollingRet.GetRespAt())
			return pollingRet, nil
		} else {
			if pollingRet.GetError() != nil {
				return pollingRet, nil
			}
		}
		time.Sleep(3 * time.Second)
	}
}

func (r *AsyncRequester) submit() (SubmitResponse, error) {
	client := http_client.New()
	body, contentType, err := r.SubmitRequest.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.SubmitRequest.Path(r.token.Supplier)),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	reqAt := time.Now()
	resp, err := client.Do(req)
	respAt := time.Now()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logs.Logger.Info().
		Int("task_id", r.TaskID).
		Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.SubmitRequest.Path(r.token.Supplier)).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("req_consume_ms", respAt.Sub(reqAt)).
		Msg("image request")
	ret := r.SubmitRequest.InitResponse(r.token.Supplier.String(), r.token.Desc)
	ret.SetTaskID(r.TaskID)
	ret.SetReqAt(reqAt)
	ret.SetRespAt(respAt)
	err = r.SubmitParser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (r *AsyncRequester) polling() (Response, error) {
	client := http_client.New()
	_, contentType, err := r.PollingRequest.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodGet,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.PollingRequest.Path(r.token.Supplier)),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
	)
	if err != nil {
		return nil, err
	}
	reqAt := time.Now()
	resp, err := client.Do(req)
	respAt := time.Now()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logs.Logger.Info().
		Int("task_id", r.TaskID).
		Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.PollingRequest.Path(r.token.Supplier)).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("req_consume_ms", respAt.Sub(reqAt)).
		Msg("image request")
	ret := r.PollingRequest.InitResponse(r.token.Supplier.String(), r.token.Desc)
	ret.SetTaskID(r.TaskID)
	ret.SetReqAt(reqAt)
	ret.SetRespAt(respAt)
	err = r.PollingParser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
