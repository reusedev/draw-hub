package image

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"io"
	"time"
)

type Request[T any] interface {
	BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error)
	Path() string
	InitResponse(supplier string, duration time.Duration, tokenDesc string) T
}
