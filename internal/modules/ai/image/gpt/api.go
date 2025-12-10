package gpt

import (
	"context"
	"errors"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
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
	var model string
	var getToken func() *ai.TokenWithModel
	if request.Model == consts.GPT4oImageVip.String() {
		getToken = ai.GTokenManager[consts.GPT4oImageVip.String()].GetTokenIterator()
		model = consts.GPT4oImageVip.String()
	} else {
		getToken = ai.GTokenManager[consts.GPT4oImage.String()].GetTokenIterator()
		model = consts.GPT4oImage.String()
	}
	for {
		token := getToken()
		if token == nil {
			break
		}
		if request.Model != "" && token.Model != request.Model {
			continue
		}
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("token_desc", token.Desc).Str("model", token.Model).Msg("Attempting GPT SlowSpeed request")

		content := Image4oRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      token.Model,
		}
		requester := image.NewRequester(ai.Token{Token: token.Token.Token, Desc: token.Desc, Supplier: token.Supplier}, &content, NewImage4oParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response := requester.Do()
		ret = append(ret, response)
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Strs("image_urls", urls).Msg("GPT SlowSpeed request succeeded, stopping iteration")
			break
		}
		logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("model", token.Model).Msg("GPT SlowSpeed request completed but failed validation, continuing")
		if response.GetError() != nil {
			if errors.Is(response.GetError(), image.PromptError) {
				break
			}
		}
		if image.ShouldBanToken(response) {
			ai.GTokenManager[model].Ban(token.Supplier, time.Now().Add(10*time.Minute))
		}
	}
	once.Do(func() { p.Notify(consts.EventTaskEnd, ret) })
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
		Msg("GPT FastSpeed method started")

	ret := make([]image.Response, 0)
	attemptCount := 0

	getToken := ai.GTokenManager[consts.GPTImage1.String()].GetTokenIterator()
	for {
		token := getToken()
		if token == nil {
			break
		}
		attemptCount++
		logs.Logger.Info().
			Int("task_id", request.TaskID).
			Int("attempt", attemptCount).
			Str("supplier", token.Supplier.String()).
			Str("token_desc", token.Desc).
			Str("model", token.Model).
			Str("quality", request.Quality).
			Str("size", request.Size).
			Msg("Attempting GPT FastSpeed request")

		content := Image1Request{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Quality:    request.Quality,
			Size:       request.Size,
		}
		requester := image.NewRequester(ai.Token{Token: token.Token.Token, Desc: token.Desc, Supplier: token.Supplier}, &content, NewImage1Parser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response := requester.Do()
		ret = append(ret, response)
		if response.Succeed() {
			logs.Logger.Info().
				Int("task_id", request.TaskID).
				Int("attempt", attemptCount).
				Str("supplier", token.Supplier.String()).
				Str("model", token.Model).
				Int("total_attempts", attemptCount).
				Msg("GPT FastSpeed request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().
				Int("task_id", request.TaskID).
				Int("attempt", attemptCount).
				Str("supplier", token.Supplier.String()).
				Str("model", token.Model).
				Msg("GPT FastSpeed request completed but failed validation, continuing")
			if response.GetError() != nil {
				if errors.Is(response.GetError(), image.PromptError) {
					break
				}
			}
			if image.ShouldBanToken(response) {
				ai.GTokenManager[consts.GPTImage1.String()].Ban(token.Supplier, time.Now().Add(10*time.Minute))
			}
		}
	}
	once.Do(func() { p.Notify(consts.EventTaskEnd, ret) })
}
