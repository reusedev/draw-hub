package gpt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
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
		image.BaseResponse{
			Supplier:  supplier,
			TokenDesc: tokenDesc,
			Model:     g.Model,
			URLs:      []string{},
		},
	}
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
		if len(g.ImageBytes) == 0 {
			header := make(textproto.MIMEHeader)
			header.Set("Content-Disposition", fmt.Sprintf(`form-data; name="image"; filename="%s"`, "image.png"))
			filePart, err := writer.CreatePart(header)
			if err != nil {
				return nil, "", err
			}
			_, err = filePart.Write(nil)
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
		image.BaseResponse{
			Supplier:  supplier,
			TokenDesc: tokenDesc,
			Model:     consts.GPTImage1.String(),
			URLs:      []string{},
		},
	}
	return ret
}

// GenericRequest supports both generate (JSON) and edit (multipart/form-data) for gpt-image-2 etc.
// For Geek supplier, both generate and edit use unified JSON endpoint (v1/images/generations).
type GenericRequest struct {
	ImageBytes [][]byte `json:"image_bytes"`
	ImageURLs  []string `json:"image_urls"`
	Prompt     string   `json:"prompt"`
	Quality    string   `json:"quality"`
	Size       string   `json:"size"`
	Model      string   `json:"model"`
}

func (g *GenericRequest) isEdit() bool {
	return len(g.ImageBytes) > 0 || len(g.ImageURLs) > 0
}

func (g *GenericRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	if supplier == consts.Geek {
		// Geek supplier: unified JSON endpoint for both generate and edit
		return g.geekJSON()
	}
	if g.isEdit() {
		// edit: multipart/form-data (OpenAI compatible)
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
		_ = writer.WriteField("model", g.Model)
		_ = writer.WriteField("n", "1")
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
	// generate: JSON
	body := map[string]interface{}{
		"model":  g.Model,
		"prompt": g.Prompt,
		"n":      1,
	}
	if g.Quality != "" {
		body["quality"] = g.Quality
	}
	if g.Size != "" {
		body["size"] = g.Size
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}

// geekJSON builds a JSON body for GeekAI's unified gpt-image-2 API.
// Single image -> "image" field, multiple images -> "images" field.
// ImageURLs take priority over ImageBytes.
func (g *GenericRequest) geekJSON() (io.Reader, string, error) {
	body := map[string]interface{}{
		"model":  g.Model,
		"prompt": g.Prompt,
	}
	if g.Quality != "" {
		body["quality"] = g.Quality
	}
	if g.Size != "" {
		body["size"] = g.Size
	}
	if len(g.ImageURLs) > 0 {
		// URL 优先
		if len(g.ImageURLs) == 1 {
			body["image"] = g.ImageURLs[0]
		} else {
			body["images"] = g.ImageURLs
		}
	} else if len(g.ImageBytes) > 0 {
		if len(g.ImageBytes) == 1 {
			body["image"] = base64.StdEncoding.EncodeToString(g.ImageBytes[0])
		} else {
			images := make([]string, 0, len(g.ImageBytes))
			for _, img := range g.ImageBytes {
				images = append(images, base64.StdEncoding.EncodeToString(img))
			}
			body["images"] = images
		}
	}
	data, err := jsoniter.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}

func (g *GenericRequest) Path(supplier consts.ModelSupplier) string {
	if supplier == consts.Geek {
		// GeekAI uses a unified endpoint for both generate and edit
		return "v1/images/generations"
	}
	if g.isEdit() {
		return "v1/images/edits"
	}
	return "v1/images/generations"
}
func (g *GenericRequest) InitResponse(supplier string, tokenDesc string) image.Response {
	return &GenericResponse{
		image.BaseResponse{
			Supplier:  supplier,
			TokenDesc: tokenDesc,
			Model:     g.Model,
			URLs:      []string{},
		},
	}
}

type GenericResponse struct {
	image.BaseResponse
}
