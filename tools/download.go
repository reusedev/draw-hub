package tools

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

func GetOnlineImage(url string) (bytes []byte, fName string, err error) {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
		return
	}

	bytes, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if resp.Header.Get("Content-Disposition") != "" {
		parts := strings.Split(resp.Header.Get("Content-Disposition"), ";")
		for _, part := range parts {
			if strings.Contains(part, "filename=") {
				fName = strings.Trim(strings.Split(part, "=")[1], "\"")
				break
			}
		}
	}
	return
}
