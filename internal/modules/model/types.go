package model

import (
	"github.com/reusedev/draw-hub/internal/consts"
	"io"
)

type RequestContent interface {
	BodyContentType(supplier consts.ModelSupplier) (io.Reader, string, error)
	Path() string
}
