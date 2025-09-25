package grok

import (
	"context"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/chat"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

func DeepSearch(request chat.CommonRequest) []chat.Response {
	ret := make([]chat.Response, 0)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumeSignal := make(chan struct{})
	go func() {
		consumeSignal <- struct{}{}
	}()
	for tokenWithModel := range ai.GTokenManager[request.Model].GetToken(ctx, consumeSignal) {
		requester := chat.NewRequester(ai.Token{Token: tokenWithModel.Token.Token, Desc: tokenWithModel.Desc, Supplier: tokenWithModel.Supplier}, &request, &chat.CommonParser{})
		response, err := requester.Do()
		go func() {
			consumeSignal <- struct{}{}
		}()
		if err != nil {
			logs.Logger.Err(err).Msg("grok-DeepSearch")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}
