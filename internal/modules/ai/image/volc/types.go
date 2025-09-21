package volc

import (
	"bytes"
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
	"time"
)

// JiMengV40Request Reference: https://www.volcengine.com/docs/85621/1817045
type JiMengV40Request struct {
	ReqKey      string   // 服务标识 jimeng_t2i_v40
	ImageUrls   []string // 0-10张图
	Prompt      string
	Size        int
	Width       int
	Height      int
	Scale       float32
	ForceSingle bool
	MinRatio    float32
	MaxRatio    float32
}

func (j *JiMengV40Request) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	body := make(map[string]any)
	body["req_key"] = j.ReqKey
	if len(j.ImageUrls) != 0 {
		body["image_urls"] = j.ImageUrls
	}
	body["prompt"] = j.Prompt
	if j.Size != 0 {
		body["size"] = j.Size
	}
	if j.Width != 0 {
		body["width"] = j.Width
	}
	if j.Height != 0 {
		body["height"] = j.Height
	}
	if j.Scale != 0 {
		body["scale"] = j.Scale
	}
	body["force_single"] = j.ForceSingle
	if j.MinRatio != 0 {
		body["min_ratio"] = j.MinRatio
	}
	if j.MaxRatio != 0 {
		body["max_ratio"] = j.MaxRatio
	}
	data, err := json.Marshal(body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewBuffer(data), "application/json", nil
}

func (j *JiMengV40Request) Path() string {
	return "?Action=CVSync2AsyncSubmitTask&Version=2022-08-31"
}

func (j *JiMengV40Request) InitResponse(supplier string, duration time.Duration, tokenDesc string) image.AsyncCreateResponse {
	return &CreateResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
		Model:     "jimeng_t2i_v40",
		Duration:  duration,
	}
}
