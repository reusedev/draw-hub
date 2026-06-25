//go:build !cgo

package tools

import (
	"fmt"
	"image"
	"io"
)

func decodeHEIC(reader io.Reader) (image.Image, error) {
	return nil, fmt.Errorf("HEIC decoding requires cgo and libheif")
}
