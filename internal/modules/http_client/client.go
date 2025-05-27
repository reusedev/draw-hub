package http_client

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type HttpClient struct {
	HttpClient *http.Client
}

type RequestOption func(options *RequestOptions)

type RequestOptions struct {
	body   any
	header http.Header
}

func WithBody(body any) RequestOption {
	return func(c *RequestOptions) {
		c.body = body
	}
}

func WithHeader(key, value string) RequestOption {
	return func(c *RequestOptions) {
		c.header.Set(key, value)
	}
}

func New() *HttpClient {
	return &HttpClient{
		HttpClient: http.DefaultClient,
	}
}

func (c *HttpClient) NewRequest(method string, url string, option ...RequestOption) (*http.Request, error) {
	options := &RequestOptions{header: http.Header{}}
	for _, opt := range option {
		opt(options)
	}
	var body io.Reader
	if options.body != nil {
		switch v := options.body.(type) {
		case io.Reader:
			body = v
		default:
			data, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			body = bytes.NewBuffer(data)
		}
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if options.header != nil {
		req.Header = options.header
	}
	return req, nil
}

func (c *HttpClient) Do(req *http.Request) (*http.Response, error) {
	return c.HttpClient.Do(req)
}
