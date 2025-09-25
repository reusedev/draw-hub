package volc

import (
	"context"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
	"sync"
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

type Request struct {
	ImageURLs  []string `json:"image_urls"`
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Size       string   `json:"size"`
	TaskID     int      `json:"task_id"`
}

func (p *Provider) Create(request Request) {
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
	for tokenWithModel := range ai.GTokenManager[consts.JiMengV40.String()].GetToken(ctx, consumeSignal) {
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
			Str("token_desc", tokenWithModel.Desc).Str("model", tokenWithModel.Model).Msg("Attempting JiMeng Create request")
		content := JiMengV40Request{
			ImageURLs:  request.ImageURLs,
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      tokenWithModel.Model,
			Size:       request.Size,
		}
		requester := image.NewRequester(ai.Token{Token: tokenWithModel.Token.Token, Desc: tokenWithModel.Desc, Supplier: tokenWithModel.Supplier}, &content, NewJiMengParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		go func() {
			consumeSignal <- struct{}{}
		}()
		if err != nil {
			logs.Logger.Error().Err(err).Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Msg("JiMeng Create request failed")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Strs("image_urls", urls).
				Msg("JiMeng Create request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", tokenWithModel.Supplier.String()).
				Str("model", tokenWithModel.Model).Msg("JiMeng Create request completed but failed validation, continuing")
		}
	}
	once.Do(func() { p.Notify(consts.EventSyncCreate, ret) })
}
