package image

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Client struct {
	HttpClient *http.Client
}

type requestOption func(options *requestOptions)

type requestOptions struct {
	body   any
	header http.Header
}

func withBody(body any) requestOption {
	return func(c *requestOptions) {
		c.body = body
	}
}

func withHeader(key, value string) requestOption {
	return func(c *requestOptions) {
		c.header.Set(key, value)
	}
}

func NewClient() *Client {
	return &Client{
		HttpClient: http.DefaultClient,
	}
}

func (c *Client) NewRequest(method string, url string, option ...requestOption) (*http.Request, error) {
	options := &requestOptions{}
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

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.HttpClient.Do(req)
}
