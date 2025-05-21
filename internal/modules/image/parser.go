package image

import "net/http"

type Response interface {
	URLs() ([]string, error)
	Base64s() ([]string, error)
}

type Parser interface {
	Parse(resp *http.Response) (Response, error)
}
