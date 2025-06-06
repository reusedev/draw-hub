package gpt

import (
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type FastRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	ImageURLs  []string `json:"image_urls"`
	Prompt     string   `json:"prompt"`
	Quality    string   `json:"quality"`
	Size       string   `json:"size"`
}

type SlowRequest struct {
	ImageByte []byte `json:"image_byte"`
	ImageURL  string `json:"image_url"`
	Prompt    string `json:"prompt"`
}

func SlowSpeed(request SlowRequest) []image.Response {
	ret := make([]image.Response, 0)
	for _, order := range config.GConfig.RequestOrder.SlowSpeed {
		content := Image4oRequest{
			ImageByte: request.ImageByte,
			ImageURL:  request.ImageURL,
			Prompt:    request.Prompt,
		}
		if order.Model == consts.GPT4oImageVip.String() {
			content.Vip = true
		}
		requester := image.NewRequester(image.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, &Image4oParser{})
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err).Msg("gpt-SlowSpeed")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}

func FastSpeed(request FastRequest) []image.Response {
	ret := make([]image.Response, 0)
	for _, order := range config.GConfig.RequestOrder.FastSpeed {
		content := Image1Request{
			ImageBytes: request.ImageBytes,
			ImageURLs:  request.ImageURLs,
			Prompt:     request.Prompt,
			Quality:    request.Quality,
			Size:       request.Size,
		}
		requester := image.NewRequester(image.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, &Image1Parser{})
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err).Msg("gpt-FastSpeed")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}
