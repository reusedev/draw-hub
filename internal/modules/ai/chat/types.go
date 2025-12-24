package chat

import (
	"bytes"
	"encoding/json"
	"io"
	"time"
)

type RequestContent interface {
	Body() (io.Reader, error)
	ContentType() string
	Path() string
	InitResponse(supplier string, duration time.Duration, tokenDesc string) Response
}

type CommonRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

type Content struct {
	Type     string   `json:"type"`
	Text     string   `json:"text,omitempty"`
	ImageURL ImageURL `json:"image_url,omitempty"`
}

type ImageURL struct {
	URL string `json:"url"`
}

func (c *CommonRequest) Body() (io.Reader, error) {
	b, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func (c *CommonRequest) ContentType() string {
	return "application/json"
}

func (c *CommonRequest) Path() string {
	return "/v1/chat/completions"
}

func (c *CommonRequest) InitResponse(supplier string, duration time.Duration, tokenDesc string) Response {
	return &CommonResponse{
		Supplier:  supplier,
		Duration:  duration,
		TokenDesc: tokenDesc,
	}
}
