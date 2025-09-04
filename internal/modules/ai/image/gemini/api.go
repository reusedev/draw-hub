package gemini

import (
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
)

type Request struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Model      string   `json:"model"`
	TaskID     int      `json:"task_id"` // 添加TaskID字段
}

func Create(request Request) []image.Response {
	ret := make([]image.Response, 0)
	for _, order := range config.GConfig.RequestOrder.Gemini25Flash {
		if request.Model != "" && order.Model != request.Model {
			continue
		}
		logs.Logger.Info().Int("task_id", request.TaskID).Str("supplier", order.Supplier).
			Str("token_desc", order.Desc).Str("model", order.Model).Msg("Attempting Gemini Create request")
		content := FlashImageRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      order.Model,
		}
		requester := image.NewRequester(ai.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, NewFlashImageParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Error().Err(err).Int("task_id", request.TaskID).Str("supplier", order.Supplier).
				Str("model", order.Model).Msg("Gemini Create request failed")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			logs.Logger.Info().
				Int("task_id", request.TaskID).
				Str("supplier", order.Supplier).
				Str("model", order.Model).
				Msg("Gemini Create request succeeded, stopping iteration")
			break
		} else {
			logs.Logger.Warn().
				Int("task_id", request.TaskID).
				Str("supplier", order.Supplier).
				Str("model", order.Model).
				Msg("Gemini Create request completed but failed validation, continuing")
		}
	}
		
	return ret
}
