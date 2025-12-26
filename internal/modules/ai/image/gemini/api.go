package gemini

import (
	"context"
	"errors"
	"strings"
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

type Request struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Model      string   `json:"model"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

func (p *Provider) Notify(event int, data interface{}) {
	for _, o := range p.Observers {
		o.Update(event, data)
	}
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
	getToken := ai.GTokenManager[request.Model].GetTokenIterator()
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
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.GetSupplier().String()).
			Str("token_desc", token.Desc).Str("model", token.Model).Msg("Attempting Gemini Create request")
		content := FlashImageRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      token.Model,
		}
		if token.Model == "gemini-3-pro-image-preview" && token.GetSupplier() == consts.Geek {
			if request.Model == consts.Gemini3.String() {
				content.GeekSize = "1K"
			} else if request.Model == consts.Gemini32k.String() {
				content.GeekSize = "2K"
			} else if request.Model == consts.Gemini34k.String() {
				content.GeekSize = "4K"
			}
		}
		var parser image.Parser[image.Response]
		parser = NewFlashImageParser()
		if token.Model == "gemini-nano-banana-hd" && token.GetSupplier() == consts.Geek {
			parser = image.NewGenericParser(&image.OpenAIURLStrategy{}, &image.GenericB64Strategy{})
		}
		requester := image.NewRequester(p.Ctx, ai.Token{Token: token.Token.Token, Desc: token.Desc, Supplier: token.Supplier}, &content, parser)
		requester.SetTaskID(request.TaskID) // 设置TaskID
		var retryCount int
	retry:
		response := requester.Do()
		ret = append(ret, response)
		if response.GetError() != nil {
			if strings.Contains(response.GetError().Error(), "Client.Timeout") {
				if retryCount < 1 {
					retryCount++
					logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
						Str("error", response.GetError().Error()).Int("retry_no", retryCount).
						Msg("Request error, retry")
					goto retry
				}
			}
		}
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Strs("image_urls", urls).
				Msg("Gemini Create request succeeded, stopping iteration")
			break
		}
		logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("model", token.Model).Msg("Gemini Create request completed but failed validation, continuing")
		if response.GetError() != nil {
			if errors.Is(response.GetError(), image.PromptError) {
				break
			}
		}
		if image.ShouldBanToken(response) {
			ai.GTokenManager[request.Model].Ban(token.Supplier, time.Now().Add(10*time.Minute))
		}
	}
	once.Do(func() { p.Notify(consts.EventTaskEnd, ret) })
}
