package grok

import (
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/chat"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

func DeepSearch(request chat.CommonRequest) []chat.Response {
	ret := make([]chat.Response, 0)
	for _, order := range config.GConfig.RequestOrder.DeepSearch {
		requester := chat.NewRequester(ai.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &request, &chat.CommonParser{})
		response, err := requester.Do()
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
