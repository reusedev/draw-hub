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
		content := FlashImageRequest{
			ImageBytes: request.ImageBytes,
			Prompt:     request.Prompt,
			Model:      order.Model,
		}
		requester := image.NewRequester(ai.Token{Token: order.Token, Desc: order.Desc, Supplier: consts.ModelSupplier(order.Supplier)}, &content, NewFlashImageParser())
		requester.SetTaskID(request.TaskID) // 设置TaskID
		response, err := requester.Do()
		if err != nil {
			logs.Logger.Err(err).Msg("gemini-FlashImage")
			continue
		}
		ret = append(ret, response)
		if response.Succeed() {
			break
		}
	}
	return ret
}
