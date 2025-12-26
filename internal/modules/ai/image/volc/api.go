package volc

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
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
	getToken := ai.GTokenManager[consts.JiMengV40.String()].GetTokenIterator()
	for {
		select {
		case <-p.Ctx.Done():
			return
		default:
		}
		token := getToken()
		if token == nil {
			break
		}
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("token_desc", token.Desc).Str("model", token.Model).Msg("Attempting JiMeng Create request")
		content := JiMengV40Request{
			ImageURLs:  request.ImageURLs,
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      token.Model,
			Size:       request.Size,
		}
		requester := image.NewRequester(p.Ctx, ai.Token{Token: token.Token.Token, Desc: token.Desc, Supplier: token.Supplier}, &content, NewJiMengParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response := requester.Do()
		ret = append(ret, response)
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Strs("image_urls", urls).
				Msg("JiMeng Create request succeeded, stopping iteration")
			break
		}
		logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("model", token.Model).Msg("JiMeng Create request completed but failed validation, continuing")
		if response.GetError() != nil {
			if errors.Is(response.GetError(), image.PromptError) {
				break
			}
		}
		if image.ShouldBanToken(response) {
			ai.GTokenManager[consts.JiMengV40.String()].Ban(token.Supplier, time.Now().Add(10*time.Minute))
		}
	}
	once.Do(func() { p.Notify(consts.EventTaskEnd, ret) })
}
