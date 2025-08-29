package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
	"time"
)

type FlashImageRequest struct {
	Model      string   `json:"model"`
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
}

func (f *FlashImageRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	body := make(map[string]any)
	body["model"] = f.Model
	body["stream"] = false
	body["messages"] = []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": f.Prompt,
				},
			},
		},
	}
	for _, img := range f.ImageBytes {
		imageByte := base64.StdEncoding.EncodeToString(img)
		body["messages"].([]map[string]interface{})[0]["content"] = append(body["messages"].([]map[string]interface{})[0]["content"].([]map[string]interface{}), map[string]interface{}{
			"type": "image_url",
			"image_url": map[string]string{
				"url": "data:image/png;base64," + imageByte,
			},
		})
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}
func (f *FlashImageRequest) Path() string {
	return "v1/chat/completions"
}
func (f *FlashImageRequest) InitResponse(supplier string, duration time.Duration, tokenDesc string) image.Response {
	return &FlashImageResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		Model:     f.Model,
		Duration:  duration,
		URLs:      []string{},
	}
}
