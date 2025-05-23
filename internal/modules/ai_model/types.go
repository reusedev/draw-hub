package ai_model

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"io"
	"time"
)

type RequestContent interface {
	BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error)
	Path() string
	InitResponse(supplier string, duration time.Duration, tokenDesc string) Response
}
