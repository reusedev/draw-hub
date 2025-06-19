package response

import jsoniter "github.com/json-iterator/go"

type GetImage struct {
	Path string `json:"path"`
	URL  string `json:"url"`
}

func (g *GetImage) Marsh() (string, error) {
	return jsoniter.MarshalToString(g)
}

func UnmarshalGetImage(data string) (*GetImage, error) {
	var result GetImage
	err := jsoniter.Unmarshal([]byte(data), &result)
	if err != nil {
		return nil, err
	}
	if result.URL == "" {
		result.URL = result.Path
	}
	return &result, nil
}
