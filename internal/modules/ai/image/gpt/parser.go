package gpt

import (
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type Image4oParser struct {
	*image.GenericParser
}

func NewImage4oParser() *Image4oParser {
	return &Image4oParser{
		GenericParser: image.NewGenericParser(&image.MarkdownURLStrategy{}),
	}
}

type Image1Parser struct {
	*image.GenericParser
}

func NewImage1Parser() *Image1Parser {
	return &Image1Parser{
		GenericParser: image.NewGenericParser(&image.OpenAIURLStrategy{}),
	}
}

type Image4oResponse struct {
	image.BaseResponse
}

type Image1Response struct {
	image.BaseResponse
}
