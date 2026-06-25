package common

import (
	"context"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/chat"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

func Chat(request chat.CommonRequest) []chat.Response {
	ret := make([]chat.Response, 0)
	tokenManager := ai.GTokenManager[request.Model]
	if tokenManager == nil {
		logs.Logger.Error().Str("model", request.Model).Msg("chat model token manager not found")
		return ret
	}
	getToken := tokenManager.GetTokenIterator()
	for {
		token := getToken()
		if token == nil {
			break
		}
		requester := chat.NewRequester(context.Background(), ai.Token{Token: token.Token.Token, Desc: token.Desc, Supplier: token.Supplier}, &request, &chat.CommonParser{})
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err).Msg("common-Chat")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}
