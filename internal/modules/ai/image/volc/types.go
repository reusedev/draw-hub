package volc

import (
	"bytes"
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
	"time"
)

// JiMengV40Request Reference: https://tuzi-api.apifox.cn/349741169e0
type JiMengV40Request struct {
	Model      string
	ImageURLs  []string
	ImageBytes [][]byte
	Prompt     string
	Size       string
}

func (j *JiMengV40Request) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	body := make(map[string]any)
	body["model"] = j.Model
	if supplier == consts.Tuzi {
		if len(j.ImageBytes) != 0 {
			body["image"] = j.ImageBytes
		}
	} else if supplier == consts.Geek {
		if len(j.ImageURLs) != 0 {
			body["image"] = j.ImageURLs
		}
	}
	body["prompt"] = j.Prompt
	if j.Size != "" {
		body["size"] = j.Size
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}

func (j *JiMengV40Request) Path() string {
	return "/v1/images/generations"
}

func (j *JiMengV40Request) InitResponse(supplier string, duration time.Duration, tokenDesc string) image.Response {
	return &CreateResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		Model:     "jimeng_t2i_v40",
		Duration:  duration,
	}
}
