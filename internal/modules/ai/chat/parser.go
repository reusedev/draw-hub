package chat

import (
	"encoding/json"
	"fmt"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"io"
	"net/http"
	"time"
)

type Response interface {
	RawBody() string
	Succeed() bool
	Marsh() ([]byte, error)
}

type Parser interface {
	Parse(resp *http.Response, response Response) error
}

type CommonResponse struct {
	Supplier   string        `json:"supplier"`
	TokenDesc  string        `json:"token_desc"`
	Duration   time.Duration `json:"duration"`
	Body       string        `json:"body"`
	StatusCode int           `json:"status_code"`
}

func (c *CommonResponse) RawBody() string {
	return c.Body
}
func (c *CommonResponse) Succeed() bool {
	return c.StatusCode == http.StatusOK
}
func (c *CommonResponse) Marsh() ([]byte, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return data, nil
}

type CommonParser struct{}

func (c *CommonParser) Parse(resp *http.Response, response Response) error {
	realResp := response.(*CommonResponse)
	realResp.StatusCode = resp.StatusCode
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		logs.Logger.Warn().Str("supplier", realResp.Supplier).
			Str("token_desc", realResp.TokenDesc).
			Str("path", resp.Request.URL.Path).
			Str("method", resp.Request.Method).
			Int("status_code", resp.StatusCode).
			Dur("duration", realResp.Duration).
			Str("body", string(body)).
			Msg("chat request failed")
	}
	realResp.Body = string(body)
	return nil
}
