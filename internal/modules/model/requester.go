package model

import (
	"fmt"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
)

type Requester struct {
	Supplier     consts.ModelSupplier
	token        string
	RequestTypes RequestContent
	Parser       Parser
}

func NewRequester(supplier consts.ModelSupplier, token string, requestTypes RequestContent, parser Parser) *Requester {
	return &Requester{
		Supplier:     supplier,
		token:        token,
		RequestTypes: requestTypes,
		Parser:       parser,
	}
}

func (r *Requester) Do() (Response, error) {
	client := NewHttpClient()
	body, contentType, err := r.RequestTypes.BodyContentType(r.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(tools.BaseURLBySupplier(r.Supplier), r.RequestTypes.Path()),
		WithHeader("Authorization", "Bearer "+r.token),
		WithHeader("Content-Type", contentType),
		WithBody(body),
	)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	ret, err := r.Parser.Parse(resp)
	return ret, err
}
