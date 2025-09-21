package volc

import (
	"context"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
)

type Provider struct {
	Ctx       context.Context
	Token     ai.Token
	Observers []observer.Observer
}

func NewProvider(ctx context.Context, token ai.Token, observers []observer.Observer) *Provider {
	return &Provider{
		Ctx:       ctx,
		Token:     token,
		Observers: observers,
	}
}

func (p *Provider) Notify(event int, data interface{}) {
	for _, o := range p.Observers {
		o.Update(event, data)
	}
}

type Request struct {
	ImageUrls []string `json:"image_urls"`
	Prompt    string   `json:"prompt"`
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	TaskID    int      `json:"task_id"`
}

func (p *Provider) Create(request Request) {
	go func() {}()
	ret := make([]image.AsyncCreateResponse, 0)
	logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", "volc").
		Str("token_desc", p.Token.Desc).Str("model", "jimeng_t2i_v40").Msg("Attempting JiMeng Create request")
	content := JiMengV40Request{
		ReqKey:    "jimeng_t2i_v40",
		ImageUrls: request.ImageUrls,
		Prompt:    request.Prompt,
		Width:     request.Width,
		Height:    request.Height,
	}
	requester := image.NewAsyncRequester(p.Token, &content, nil)
	requester.SetTaskID(request.TaskID)
	response, err := requester.Do()
	if err != nil {
		logs.Logger.Error().Err(err).Int("task_id", request.TaskID).Str("supplier", p.Token.Supplier.String()).
			Str("model", "jimeng_t2i_v40").Msg("JiMeng Create request failed")
	}
	if response.GetError() != nil {
		logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", p.Token.Supplier.String()).
			Str("model", "jimeng_t2i_v40").Msg("JiMeng Create response failed")
	}
	p.Notify(consts.EventAsyncCreate, ret)
}
