package ai_model

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"time"
)

type Requester struct {
	Supplier     consts.ModelSupplier
	token        Token
	RequestTypes RequestContent
	Parser       Parser
}
type Token struct {
	Token string
	Desc  string
}

func NewRequester(supplier consts.ModelSupplier, token Token, requestTypes RequestContent, parser Parser) *Requester {
	return &Requester{
		Supplier:     supplier,
		token:        token,
		RequestTypes: requestTypes,
		Parser:       parser,
	}
}

func (r *Requester) Do() (Response, error) {
	client := newHttpClient()
	body, contentType, err := r.RequestTypes.BodyContentType(r.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.newRequest(
		http.MethodPost,
		tools.FullURL(tools.BaseURLBySupplier(r.Supplier), r.RequestTypes.Path()),
		withHeader("Authorization", "Bearer "+r.token.Token),
		withHeader("Content-Type", contentType),
		withBody(body),
	)
	if err != nil {
		return nil, err
	}
	start := time.Now()
	resp, err := client.do(req)
	duration := time.Since(start)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	ret := r.RequestTypes.InitResponse(r.Supplier.String(), duration, r.token.Desc)
	err = r.Parser.Parse(resp, ret)

	return ret, nil
}
