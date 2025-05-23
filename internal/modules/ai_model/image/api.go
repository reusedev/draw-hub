package image

import (
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai_model"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type FastRequest struct {
	ImageURLs []string `json:"image_urls"`
	Prompt    string   `json:"prompt"`
	Quality   string   `json:"quality"`
	Size      string   `json:"size"`
}

type SlowRequest struct {
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt"`
}

func SlowSpeed(request SlowRequest) []ai_model.Response {
	ret := make([]ai_model.Response, 0)
	for _, order := range config.GConfig.RequestOrder.SlowSpeed {
		content := GPT4oImageRequest{
			ImageURL: request.ImageURL,
			Prompt:   request.Prompt,
		}
		if order.Model == consts.GPT4oImageVip.String() {
			content.Vip = true
		}
		requester := ai_model.NewRequester(consts.ModelSupplier(order.Supplier), ai_model.Token{Token: order.Token, Desc: order.Desc}, &content, &GPT4oImageParser{})
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err)
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}

func FastSpeed(request FastRequest) []ai_model.Response {
	ret := make([]ai_model.Response, 0)
	for _, order := range config.GConfig.RequestOrder.FastSpeed {
		content := GPTImage1Request{
			ImageURLs: request.ImageURLs,
			Prompt:    request.Prompt,
			Quality:   request.Quality,
			Size:      request.Size,
		}
		requester := ai_model.NewRequester(consts.ModelSupplier(order.Supplier), ai_model.Token{Token: order.Token, Desc: order.Desc}, &content, &GPTImage1Parser{})
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err)
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}
