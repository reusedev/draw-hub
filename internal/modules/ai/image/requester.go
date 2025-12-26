package image

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/http_client"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/tools"
)

type SyncRequester struct {
	ctx     context.Context
	token   ai.Token
	Request Request[Response]
	Parser  Parser[Response]
	TaskID  int // 添加TaskID字段用于日志跟踪
}

func NewRequester(ctx context.Context, token ai.Token, requestTypes Request[Response], parser Parser[Response]) *SyncRequester {
	return &SyncRequester{
		ctx:     ctx,
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

func (r *SyncRequester) Do() Response {
	ret := r.Request.InitResponse(r.token.Supplier.String(), r.token.Desc)
	ret.SetTaskID(r.TaskID)

	// 不设置会2分钟超时，超时断开有时仍计费。延长为6分钟
	client := http_client.NewWithTimeout(6 * time.Minute)
	body, contentType, err := r.Request.BodyContentType(r.token.Supplier)
	if err != nil {
		ret.SetError(err)
		return ret
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.Request.Path(r.token.Supplier)),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
		http_client.WithContext(r.ctx),
	)
	if err != nil {
		ret.SetError(err)
		return ret
	}
	reqAt := time.Now()
	resp, err := client.Do(req)
	respAt := time.Now()
	ret.SetReqAt(reqAt)
	ret.SetRespAt(respAt)
	ret.SetStartAt(reqAt)
	ret.SetEndAt(respAt)
	if err != nil {
		ret.SetError(err)
		return ret
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
	err = r.Parser.Parse(resp, ret)
	if err != nil {
		ret.SetError(err)
		return ret
	}
	return ret
}

type AsyncRequester struct {
	ctx             context.Context
	token           ai.Token
	SubmitRequest   Request[SubmitResponse]
	SubmitParser    Parser[SubmitResponse]
	PollingRequest  Request[Response]
	PollingParser   Parser[Response]
	OnSubmitSucceed func(response SubmitResponse)
	TaskID          int // 添加TaskID字段用于日志跟踪
}

func NewAsyncRequester(
	ctx context.Context, token ai.Token, submitRequest Request[SubmitResponse], submitParser Parser[SubmitResponse],
	pollingRequest Request[Response], pollingParser Parser[Response], onsubmitSucceed func(response SubmitResponse),
) *AsyncRequester {
	return &AsyncRequester{
		ctx:             ctx,
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
		select {
		case <-r.ctx.Done():
			return nil, r.ctx.Err()
		default:
		}

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

		select {
		case <-time.After(3 * time.Second):
		case <-r.ctx.Done():
			return nil, r.ctx.Err()
		}
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
		http_client.WithContext(r.ctx),
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
		http_client.WithContext(r.ctx),
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
