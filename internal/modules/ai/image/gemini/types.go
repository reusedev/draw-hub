package gemini

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
)

type FlashImageRequest struct {
	Model      string   `json:"model"`
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	GeekSize   string   `json:"geek_size"`
}

func (f *FlashImageRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	if supplier == consts.Geek && f.Model == "gemini-nano-banana-hd" {
		body := make(map[string]any)
		body["model"] = f.Model
		body["prompt"] = f.Prompt
		body["image"] = f.ImageBytes
		data, err := json.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		return bytes.NewBuffer(data), "application/json", nil
	}
	body := make(map[string]any)
	body["model"] = f.Model
	body["stream"] = true
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
	if supplier == consts.Geek && f.Model == "gemini-3-pro-image-preview" {
		body["image"] = map[string]interface{}{
			"image_size": f.GeekSize,
		}
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}
func (f *FlashImageRequest) Path(supplier consts.ModelSupplier) string {
	if supplier == consts.Geek && f.Model == "gemini-nano-banana-hd" {
		return "v1/images/edits"
	}
	return "v1/chat/completions"
}
func (f *FlashImageRequest) InitResponse(supplier string, tokenDesc string) image.Response {
	model := f.Model
	if supplier == consts.Geek.String() && f.Model == consts.Gemini3.String() {
		if f.GeekSize == "2K" {
			model = consts.Gemini32k.String()
		} else if f.GeekSize == "4K" {
			model = consts.Gemini34k.String()
		}
	}
	return &FlashImageResponse{
		image.BaseResponse{
			Supplier:  supplier,
			TokenDesc: tokenDesc,
			Model:     model,
			URLs:      []string{},
		},
	}
}
