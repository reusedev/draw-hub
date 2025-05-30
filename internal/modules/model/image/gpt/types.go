package gpt

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/model/image"
	"github.com/reusedev/draw-hub/tools"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"
)

type Image4oRequest struct {
	Vip       bool   `json:"vip"`
	ImageURL  string `json:"image_url"` // URL Bytes 二选一
	ImageByte []byte `json:"image_byte"`
	Prompt    string `json:"prompt"`
}

func (g *Image4oRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	var imageByte []byte
	imageByte = g.ImageByte
	if len(imageByte) == 0 {
		b, _, err := tools.GetOnlineImage(g.ImageURL)
		if err != nil {
			return nil, "", err
		}
		imageByte = b
	}
	body := make(map[string]any)
	body["model"] = "gpt-4o-image"
	if g.Vip {
		body["model"] = "gpt-4o-image-vip"
	}
	body["stream"] = false
	body["messages"] = []map[string]interface{}{
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
						"url": "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageByte),
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
func (g *Image4oRequest) Path() string {
	return "v1/chat/completions"
}
func (g *Image4oRequest) InitResponse(supplier string, duration time.Duration, tokenDesc string) image.Response {
	ret := &Image4oResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		Duration:  duration,
		URLs:      []string{},
	}
	if g.Vip {
		ret.Model = consts.GPT4oImageVip.String()
	} else {
		ret.Model = consts.GPT4oImage.String()
	}
	return ret
}

type Image1Request struct {
	ImageURLs  []string `json:"image_urls"`
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
		if len(g.ImageBytes) != 0 {
			body["image"] = base64.StdEncoding.EncodeToString(g.ImageBytes[0])
		} else if len(g.ImageURLs) != 0 {
			body["image"] = g.ImageURLs[0]
		}
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
func (g *Image1Request) Path() string {
	return "v1/images/edits"
}
func (g *Image1Request) InitResponse(supplier string, duration time.Duration, tokenDesc string) image.Response {
	ret := &Image1Response{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		Duration:  duration,
		Base64:    []string{},
		URLs:      []string{},
	}
	ret.Model = consts.GPTImage1.String()
	return ret
}
