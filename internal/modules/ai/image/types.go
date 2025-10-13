package image

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"io"
)

type Request[T any] interface {
	BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error)
	Path(supplier consts.ModelSupplier) string
	InitResponse(supplier string, tokenDesc string) T
}
