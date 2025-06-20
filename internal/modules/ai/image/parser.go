package image

import (
	"net/http"
	"time"
)

type Response interface {
	GetModel() string
	GetSupplier() string
	GetTokenDesc() string
	GetStatusCode() int
	GetRespAt() time.Time
	FailedRespBody() string // != 200
	DurationMs() int64

	Succeed() bool
	GetURLs() []string
	GetError() error
}

type Parser interface {
	Parse(resp *http.Response, response Response) error
}
