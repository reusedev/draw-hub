package volc

import (
	"encoding/json"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
)

type JiMengParser struct {
	*image.GenericParser
}

func NewJiMengParser() *JiMengParser {
	return &JiMengParser{
		GenericParser: image.NewGenericParser(&JiMengParserStrategy{}, &image.GenericB64Strategy{}),
	}
}

type JiMengParserStrategy struct{}

func (j *JiMengParserStrategy) ExtractURLs(body []byte) ([]string, error) {
	var responseBody struct {
		Data []struct {
			Url string `json:"url"`
		} `json:"data"`
		Created int `json:"created"`
		Usage   struct {
			PromptTokens        int `json:"prompt_tokens"`
			CompletionTokens    int `json:"completion_tokens"`
			TotalTokens         int `json:"total_tokens"`
			PromptTokensDetails struct {
				CachedTokensDetails struct {
				} `json:"cached_tokens_details"`
			} `json:"prompt_tokens_details"`
			CompletionTokensDetails struct {
			} `json:"completion_tokens_details"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}
	ret := make([]string, 0)
	for _, data := range responseBody.Data {
		if data.Url != "" {
			ret = append(ret, data.Url)
		}
	}
	return ret, nil
}

type CreateResponse struct {
	image.BaseResponse
}
