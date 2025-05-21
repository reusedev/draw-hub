package image

import (
	"fmt"
	"github.com/reusedev/draw-hub/internal/modules/consts"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
)

type Requester struct {
	Supplier     consts.ImageSupplier
	token        string
	RequestTypes RequestTypes
	Parser       Parser
}

func NewRequester(supplier consts.ImageSupplier, token string, requestTypes RequestTypes, parser Parser) *Requester {
	return &Requester{
		Supplier:     supplier,
		token:        token,
		RequestTypes: requestTypes,
		Parser:       parser,
	}
}

func (r *Requester) Do() (Response, error) {
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
