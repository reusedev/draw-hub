package ai_model

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"time"
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
	start := time.Now()
	resp, err := client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret := r.RequestTypes.InitResponse(r.Supplier.String(), duration)
	err = r.Parser.Parse(resp, ret)

	return ret, nil
}
