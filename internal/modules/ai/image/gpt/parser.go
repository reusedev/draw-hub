package gpt

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type Image4oParser struct {
	*image.GenericParser
}

func NewImage4oParser() *Image4oParser {
	return &Image4oParser{
		GenericParser: image.NewGenericParser(&image.MarkdownURLStrategy{}, &image.GenericB64Strategy{}),
	}
}

type Image1Parser struct {
	*image.GenericParser
}

func NewImage1Parser() *Image1Parser {
	return &Image1Parser{
		GenericParser: image.NewGenericParser(&image.OpenAIURLStrategy{}, &gptImage1B64Strategy{}),
	}
}

type Image4oResponse struct {
	image.BaseResponse
}

type Image1Response struct {
	image.BaseResponse
}

type gptImage1B64Strategy struct{}

func (g *gptImage1B64Strategy) ExtractB64s(body []byte) ([]string, error) {
	b64s := make([]string, 0)
	var s struct {
		Data []struct {
			URL           string `json:"url,omitempty"`
			B64JSON       string `json:"b64_json,omitempty"`
			RevisedPrompt string `json:"revised_prompt,omitempty"`
		} `json:"data"`
	}
	err := jsoniter.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	for _, v := range s.Data {
		if v.B64JSON != "" {
			b64s = append(b64s, v.B64JSON)
		}
	}
	return b64s, nil
}
