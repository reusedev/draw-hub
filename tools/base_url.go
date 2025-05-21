package tools

import (
	draw_hub "github.com/reusedev/draw-hub/internal/modules/consts"
)

func BaseURLBySupplier(supplier draw_hub.ImageSupplier) string {
	switch supplier {
	case draw_hub.Geek:
		return draw_hub.GeekBaseURL
	case draw_hub.Tuzi:
		return draw_hub.TuziBaseURL
	case draw_hub.V3:
		return draw_hub.V3BaseUrl
	default:
		return ""
	}
}

func FullURL(baseURL, path string) string {
	if baseURL == "" {
		return ""
	}
	if baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	if path == "" {
		return baseURL
	}
	if path[0] == '/' {
		path = path[1:]
	}
	return baseURL + "/" + path
}
