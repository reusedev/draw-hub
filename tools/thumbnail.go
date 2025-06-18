package tools

import (
	"bytes"
	"github.com/disintegration/imaging"
	"io"
)

func Thumbnail(r io.Reader, ratio float64, format imaging.Format) (io.Reader, error) {
	img, err := imaging.Decode(r)
	if err != nil {
		return nil, err
	}
	b := img.Bounds()
	width := int(float64(b.Dx()) * ratio)
	height := int(float64(b.Dy()) * ratio)
	thumbnail := imaging.Thumbnail(img, width, height, imaging.Lanczos)
	if thumbnail == nil {
		return nil, io.ErrUnexpectedEOF
	}
	var buf bytes.Buffer
	err = imaging.Encode(&buf, thumbnail, format)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}
