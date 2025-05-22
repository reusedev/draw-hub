package model

import "net/http"

type Response interface {
	String() string
}

type Parser interface {
	Parse(resp *http.Response) (Response, error)
}
