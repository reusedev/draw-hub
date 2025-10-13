package mj

import (
	"context"
	"encoding/base64"
	"errors"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/observer"
	"strconv"
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

type Request struct {
	ImageURLs  []string `json:"image_urls"`
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
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
	getToken := ai.GTokenManager[consts.MidJourney.String()].GetTokenIterator()
	b64s := make([]string, 0)
	if len(request.ImageBytes) != 0 {
		for _, v := range request.ImageBytes {
			b64s = append(b64s, base64.StdEncoding.EncodeToString(v))
		}
	}
	for {
		token := getToken()
		if token == nil {
			break
		}
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
			Str("token_desc", token.Desc).Str("model", token.Model).Msg("Attempting Midjourney Create request")
		content := ImagineRequest{
			Prompt:      request.Prompt,
			Base64Array: b64s,
		}
		pollingContent := FetchRequest{}
		requester := image.NewAsyncRequester(
			ai.Token{
				Token:    token.Token.Token,
				Desc:     token.Desc,
				Supplier: token.Supplier,
			},
			&content,
			image.NewSubmitParser(&providerTaskIDStrategy{}),
			&pollingContent,
			image.NewGenericParser(&urlStrategy{}),
			func(response image.SubmitResponse) {
				pollingContent.ID = strconv.FormatInt(response.GetProviderTaskID(), 10)
			},
		)
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Error().Err(err).Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Msg("Midjourney Create request failed")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			urls := response.GetURLs()
			logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Strs("image_urls", urls).
				Msg("Midjourney Create request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().Int("task_id", request.TaskID).Str("supplier", token.Supplier.String()).
				Str("model", token.Model).Msg("Midjourney Create request completed but failed validation, continuing")
			if response.GetError() != nil {
				if errors.Is(response.GetError(), image.PromptError) {
					break
				}
			}
			if image.ShouldBanToken(response) {
				ai.GTokenManager[consts.MidJourney.String()].Ban(token.Supplier, time.Now().Add(10*time.Minute))
			}
		}
	}
	once.Do(func() { p.Notify(consts.EventSyncCreate, ret) })
}
