package image

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai_model"
	"github.com/reusedev/draw-hub/tools"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

type GPT4oImageRequest struct {
	Vip      bool   `json:"vip"`
	ImageURL string `json:"image_url"`
	Prompt   string `json:"prompt"`
}

func (g *GPT4oImageRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	imageBytes, _, err := tools.GetOnlineImage(g.ImageURL)
	if err != nil {
		return nil, "", err
	}
	body := make(map[string]any)
	body["ai_model"] = "gpt-4o-image"
	if g.Vip {
		body["ai_model"] = "gpt-4o-image-vip"
	}
	body["stream"] = false
	body["message"] = []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": g.Prompt,
				},
				{
					"type": "image_url",
					"image_url": map[string]string{
						"url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageBytes),
					},
				},
			},
		},
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}
func (g *GPT4oImageRequest) Path() string {
	return "v1/chat/completions"
}
func (g *GPT4oImageRequest) InitResponse(supplier string, duration time.Duration) ai_model.Response {
	ret := &GPT4oImageResponse{
		Supplier: supplier,
		Duration: duration,
	}
	if g.Vip {
		ret.Model = consts.GPT4oImageVip.String()
	} else {
		ret.Model = consts.GPT4oImage.String()
	}
	return ret
}

type GPTImage1Request struct {
	ImageURLs []string `json:"image_urls"`
	Prompt    string   `json:"prompt"`
	Quality   string   `json:"quality"`
	Size      string   `json:"size"`
}

func (g *GPTImage1Request) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	for _, f := range g.ImageURLs {
		imageBytes, fName, err := tools.GetOnlineImage(f)
		header := make(textproto.MIMEHeader)
		header.Set("Content-Type", http.DetectContentType(imageBytes))
		header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, fName))
		filePart, err := writer.CreatePart(header)
		if err != nil {
			return nil, "", err
		}
		_, err = filePart.Write(imageBytes)
		if err != nil {
			return nil, "", err
		}
	}
	_ = writer.WriteField("prompt", g.Prompt)
	_ = writer.WriteField("ai_model", "gpt-image-1")
	_ = writer.WriteField("quality", g.Quality)
	_ = writer.WriteField("size", g.Size)
	err := writer.Close()
	if err != nil {
		return nil, "", err
	}
	return payload, writer.FormDataContentType(), nil
}
func (g *GPTImage1Request) Path() string {
	return "v1/images/edits"
}
