package image

import (
	"github.com/reusedev/draw-hub/internal/modules/consts"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
)

type Request struct {
	Supplier     consts.ImageSupplier
	token        string
	RequestTypes RequestTypes
	client       Client
}

func (r *Request) Do() (*http.Response, error) {
	client := NewClient()
	body, contentType, err := r.RequestTypes.BodyContentType(r.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(tools.BaseURLBySupplier(r.Supplier), r.RequestTypes.Path()),
		withHeader("Authorization", "Bearer "+r.token),
		withHeader("Content-Type", contentType),
		withBody(body),
	)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
