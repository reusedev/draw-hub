package gemini

import (
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type FlashImageParser struct {
	*image.StreamParser
}

func NewFlashImageParser() *FlashImageParser {
	return &FlashImageParser{
		StreamParser: image.NewStreamParser(&image.MarkdownURLStrategy{}, &image.GenericB64Strategy{}),
	}
}

type FlashImageResponse struct {
	image.BaseResponse
}
