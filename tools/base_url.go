package tools

import (
	draw_hub "github.com/reusedev/draw-hub/internal/consts"
	"strings"
)

func BaseURLBySupplier(supplier draw_hub.ModelSupplier) string {
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
	return strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
}
