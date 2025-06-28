package tools

import (
	"strings"
)

func FullURL(baseURL, path string) string {
	if baseURL == "" {
		return ""
	}
	return strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
}
