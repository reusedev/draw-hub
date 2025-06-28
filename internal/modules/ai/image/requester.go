package image

import (
	"github.com/reusedev/draw-hub/internal/modules/ai"
	"github.com/reusedev/draw-hub/internal/modules/http_client"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"time"
)

type Requester struct {
	token        ai.Token
	RequestTypes RequestContent
	Parser       Parser
}

func NewRequester(token ai.Token, requestTypes RequestContent, parser Parser) *Requester {
	return &Requester{
		token:        token,
		RequestTypes: requestTypes,
		Parser:       parser,
	}
}

func (r *Requester) Do() (Response, error) {
	client := http_client.New()
	body, contentType, err := r.RequestTypes.BodyContentType(r.token.Supplier)
	if err != nil {
		return nil, err
	}
	req, err := client.NewRequest(
		http.MethodPost,
		tools.FullURL(r.token.GetSupplier().BaseURL(), r.RequestTypes.Path()),
		http_client.WithHeader("Authorization", "Bearer "+r.token.Token),
		http_client.WithHeader("Content-Type", contentType),
		http_client.WithBody(body),
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
	logs.Logger.Info().Str("supplier", r.token.Supplier.String()).
		Str("token_desc", r.token.Desc).
		Str("path", r.RequestTypes.Path()).
		Str("method", req.Method).
		Int("status_code", resp.StatusCode).
		Dur("duration", duration).
		Msg("image request")
	ret := r.RequestTypes.InitResponse(r.token.Supplier.String(), duration, r.token.Desc)
	err = r.Parser.Parse(resp, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}
