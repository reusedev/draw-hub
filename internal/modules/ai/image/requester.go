package image

import (
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/http_client"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"time"
)

type SyncRequester struct {
	token       ai.Token
	RequestType Request[Response]
	Parser      Parser[Response]
	TaskID      int // 添加TaskID字段用于日志跟踪
}

func NewRequester(token ai.Token, requestTypes Request[Response], parser Parser[Response]) *SyncRequester {
	return &SyncRequester{
		token:       token,
		RequestType: requestTypes,
		Parser:      parser,
		TaskID:      0, // 默认值，需要调用方设置
	}
}

func (r *SyncRequester) SetTaskID(taskID int) *SyncRequester {
	r.TaskID = taskID
	return r
}

func (r *SyncRequester) Do() (Response, error) {
	client := http_client.New()
	body, contentType, err := r.RequestType.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.RequestType.Path()),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logs.Logger.Info().
		Int("task_id", r.TaskID).
		Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.RequestType.Path()).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("duration", duration).
		Msg("image request")
	ret := r.RequestType.InitResponse(r.token.Supplier.String(), duration, r.token.Desc)
	ret.SetTaskID(r.TaskID)
	err = r.Parser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type AsyncCreateRequester struct {
	token       ai.Token
	RequestType Request[AsyncCreateResponse]
	Parser      Parser[AsyncCreateResponse]
	TaskID      int
}

func NewAsyncRequester(token ai.Token, requestTypes Request[AsyncCreateResponse], parser Parser[AsyncCreateResponse]) *AsyncCreateRequester {
	return &AsyncCreateRequester{
		token:       token,
		RequestType: requestTypes,
		Parser:      parser,
		TaskID:      0, // 默认值，需要调用方设置
	}
}

func (r *AsyncCreateRequester) SetTaskID(taskID int) *AsyncCreateRequester {
	r.TaskID = taskID
	return r
}

func (r *AsyncCreateRequester) Do() (AsyncCreateResponse, error) {
	client := http_client.New()
	body, contentType, err := r.RequestType.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.RequestType.Path()),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logs.Logger.Info().
		Int("task_id", r.TaskID).
		Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.RequestType.Path()).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("duration", duration).
		Msg("image request")
	ret := r.RequestType.InitResponse(r.token.Supplier.String(), duration, r.token.Desc)
	ret.SetTaskID(r.TaskID) // 设置TaskID到响应中
	err = r.Parser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}
	return ret, nil
}
