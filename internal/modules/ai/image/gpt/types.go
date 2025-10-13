package gpt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

type Image4oRequest struct {
	Model      string   `json:"model"`
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
}

func (g *Image4oRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	body := make(map[string]any)
	body["model"] = g.Model
	body["stream"] = false
	body["messages"] = []map[string]interface{}{
		{
			"role": "user",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": g.Prompt,
				},
			},
		},
	}
	for _, img := range g.ImageBytes {
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
func (g *Image4oRequest) Path(supplier consts.ModelSupplier) string {
	return "v1/chat/completions"
}
func (g *Image4oRequest) InitResponse(supplier string, tokenDesc string) image.Response {
	ret := &Image4oResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		URLs:      []string{},
	}
	ret.Model = g.Model
	return ret
}

type Image1Request struct {
	ImageBytes [][]byte `json:"image_bytes"`
	Prompt     string   `json:"prompt"`
	Quality    string   `json:"quality"`
	Size       string   `json:"size"`
}

func (g *Image1Request) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	if supplier == consts.Geek {
		body := map[string]interface{}{}
		body["model"] = "gpt-image-1"
		body["n"] = 1
		body["prompt"] = g.Prompt
		var images []string
		for _, img := range g.ImageBytes {
			imageByte := base64.StdEncoding.EncodeToString(img)
			images = append(images, imageByte)
		}
		body["image"] = images
		if g.Size != "" {
			body["size"] = g.Size
		}
		if g.Quality != "" {
			body["quality"] = g.Quality
		}
		b, err := jsoniter.Marshal(body)
		if err != nil {
			return nil, "", err
		}
		payload := bytes.NewBuffer(b)
		return payload, "application/json", nil
	} else {
		payload := &bytes.Buffer{}
		writer := multipart.NewWriter(payload)

		for _, b := range g.ImageBytes {
			header := make(textproto.MIMEHeader)
			header.Set("Content-Type", http.DetectContentType(b))
			header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, "image.png"))
			filePart, err := writer.CreatePart(header)
			if err != nil {
				return nil, "", err
			}
			_, err = filePart.Write(b)
			if err != nil {
				return nil, "", err
			}
		}
		_ = writer.WriteField("prompt", g.Prompt)
		_ = writer.WriteField("model", "gpt-image-1")
		if g.Quality != "" {
			_ = writer.WriteField("quality", g.Quality)
		}
		if g.Size != "" {
			_ = writer.WriteField("size", g.Size)
		}
		err := writer.Close()
		if err != nil {
			return nil, "", err
		}
		return payload, writer.FormDataContentType(), nil
	}
}
func (g *Image1Request) Path(supplier consts.ModelSupplier) string {
	return "v1/images/edits"
}
func (g *Image1Request) InitResponse(supplier string, tokenDesc string) image.Response {
	ret := &Image1Response{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		URLs:      []string{},
	}
	ret.Model = consts.GPTImage1.String()
	return ret
}
