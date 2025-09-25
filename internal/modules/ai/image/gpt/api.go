package gpt

import (
	"context"
	"errors"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
	"net/http"
	"sync"
	"time"
)

type Provider struct {
	Ctx       context.Context
	Observers []observer.Observer
}

func NewProvider(ctx context.Context, observers []observer.Observer) *Provider {
	return &Provider{
		Ctx:       ctx,
		Observers: observers,
	}
}

func (p *Provider) Notify(event int, data interface{}) {
	for _, o := range p.Observers {
		o.Update(event, data)
	}
}

type FastRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Quality    string   `json:"quality"`
	Size       string   `json:"size"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

type SlowRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Model      string   `json:"model"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

func (p *Provider) SlowSpeed(request SlowRequest) {
	var once sync.Once
	down := make(chan struct{})
	defer func() { down <- struct{}{} }()
	go func() {
		select {
		case <-p.Ctx.Done():
			once.Do(func() {
				p.Notify(consts.EventSysExit, &image.GenericSysExitResponse{
					TaskID: request.TaskID,
				})
			})
			return
		case <-down:
			return
		}
	}()
	ret := make([]image.Response, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumeSignal := make(chan struct{})
	go func() {
		consumeSignal <- struct{}{}
	}()
	for tokenWithModel := range ai.GTokenManager["slow_speed"].GetToken(ctx, consumeSignal) {
		if request.Model != "" && tokenWithModel.Model != request.Model {
			continue
		}
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
			Str("token_desc", tokenWithModel.Desc).Str("model", tokenWithModel.Model).Msg("Attempting GPT SlowSpeed request")

		content := Image4oRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      tokenWithModel.Model,
		}
		requester := image.NewRequester(ai.Token{Token: tokenWithModel.Token.Token, Desc: tokenWithModel.Desc, Supplier: tokenWithModel.Supplier}, &content, NewImage4oParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Error().Err(err).Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Msg("GPT SlowSpeed request failed")
			go func() {
				consumeSignal <- struct{}{}
			}()
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Strs("image_urls", urls).Msg("GPT SlowSpeed request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Msg("GPT SlowSpeed request completed but failed validation, continuing")
			if response.GetError() != nil {
				if errors.Is(response.GetError(), image.PromptError) {
					break
				}
			}
			if response.GetStatusCode() == http.StatusBadGateway {
				ai.GTokenManager["slow_speed"].Ban(tokenWithModel.Supplier, time.Now().Add(10*time.Minute))
			}
			go func() {
				consumeSignal <- struct{}{}
			}()
		}
	}
	once.Do(func() { p.Notify(consts.EventSyncCreate, ret) })
}

func (p *Provider) FastSpeed(request FastRequest) {
	var once sync.Once
	down := make(chan struct{})
	defer func() { down <- struct{}{} }()
	go func() {
		select {
		case <-p.Ctx.Done():
			once.Do(func() {
				p.Notify(consts.EventSysExit, &image.GenericSysExitResponse{
					TaskID: request.TaskID,
				})
			})
			return
		case <-down:
			return
		}
	}()
	// 记录方法开始执行日志
	logs.Logger.Info().
		Int("task_id", request.TaskID).
		Str("method", "FastSpeed").
		Str("quality", request.Quality).
		Str("size", request.Size).
		Int("available_orders", len(config.GConfig.RequestOrder.FastSpeed)).
		Msg("GPT FastSpeed method started")

	ret := make([]image.Response, 0)
	attemptCount := 0

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumeSignal := make(chan struct{})
	go func() {
		consumeSignal <- struct{}{}
	}()
	for tokenWithModel := range ai.GTokenManager["fast_speed"].GetToken(ctx, consumeSignal) {
		attemptCount++
		logs.Logger.Info().
			Int("task_id", request.TaskID).
			Int("attempt", attemptCount).
			Str("supplier", tokenWithModel.Supplier.String()).
			Str("token_desc", tokenWithModel.Desc).
			Str("model", tokenWithModel.Model).
			Str("quality", request.Quality).
			Str("size", request.Size).
			Msg("Attempting GPT FastSpeed request")

		content := Image1Request{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Quality:    request.Quality,
			Size:       request.Size,
		}
		requester := image.NewRequester(ai.Token{Token: tokenWithModel.Token.Token, Desc: tokenWithModel.Desc, Supplier: tokenWithModel.Supplier}, &content, NewImage1Parser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Error().
				Err(err).
				Int("task_id", request.TaskID).
				Int("attempt", attemptCount).
				Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).
				Msg("GPT FastSpeed request failed")
			go func() {
				consumeSignal <- struct{}{}
			}()
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			logs.Logger.Info().
				Int("task_id", request.TaskID).
				Int("attempt", attemptCount).
				Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).
				Int("total_attempts", attemptCount).
				Msg("GPT FastSpeed request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().
				Int("task_id", request.TaskID).
				Int("attempt", attemptCount).
				Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).
				Msg("GPT FastSpeed request completed but failed validation, continuing")
			if response.GetError() != nil {
				if errors.Is(response.GetError(), image.PromptError) {
					break
				}
			}
			if response.GetStatusCode() == http.StatusBadGateway {
				ai.GTokenManager["fast_speed"].Ban(tokenWithModel.Supplier, time.Now().Add(10*time.Minute))
			}
			go func() {
				consumeSignal <- struct{}{}
			}()
		}
	}
	once.Do(func() { p.Notify(consts.EventSyncCreate, ret) })
}
