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
	GetBase64() string // image1
	GetURLs() []string // 4o-image
}

type Parser interface {
	Parse(resp *http.Response, response Response) error
}
