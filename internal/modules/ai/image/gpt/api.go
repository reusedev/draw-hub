package gpt

import (
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type FastRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Quality    string   `json:"quality"`
	Size       string   `json:"size"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

type SlowRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Model      string   `json:"model"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

func SlowSpeed(request SlowRequest) []image.Response {
	ret := make([]image.Response, 0)
	for _, order := range config.GConfig.RequestOrder.SlowSpeed {
		if request.Model != "" && order.Model != request.Model {
			continue
		}
		content := Image4oRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      order.Model,
		}
		requester := image.NewRequester(ai.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, NewImage4oParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
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
			Prompt:     request.Prompt,
			Quality:    request.Quality,
			Size:       request.Size,
		}
		requester := image.NewRequester(ai.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, NewImage1Parser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
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
