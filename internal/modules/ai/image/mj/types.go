package mj

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"io"
)

type ImagineRequest struct {
	Prompt      string   `json:"prompt"`
	Base64Array []string `json:"base64Array"`
}

func (i *ImagineRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	b, err := json.Marshal(i)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(b), "application/json", nil
}

func (i *ImagineRequest) Path(supplier consts.ModelSupplier) string {
	return "/mj/submit/imagine"
}

func (i *ImagineRequest) InitResponse(supplier string, tokenDesc string) image.SubmitResponse {
	return &ImagineResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
	}
}

type FetchRequest struct {
	ID string `json:"id"`
}

func (f *FetchRequest) BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error) {
	return nil, "application/json", nil
}

func (f *FetchRequest) Path(supplier consts.ModelSupplier) string {
	return fmt.Sprintf("/mj/task/%s/fetch", f.ID)
}

func (f *FetchRequest) InitResponse(supplier string, tokenDesc string) image.Response {
	return &FetchResponse{
		Supplier:  supplier,
		TokenDesc: tokenDesc,
	}
}
