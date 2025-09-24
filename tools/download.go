package tools

import (
	"fmt"
	"github.com/disintegration/imaging"
	"io"
	"net/http"
	"strings"
	"time"
)

func GetOnlineImage(url string) (bytes []byte, fName string, err error) {
	if strings.TrimSpace(url) == "" {
		return nil, "", fmt.Errorf("empty URL provided")
	}
	retry := 5
label:
	retry--
	bytes, fName, err = getOnlineImage(url)
	if err != nil {
		if retry > 0 {
			time.Sleep(3 * time.Second)
			goto label
		}
	}
	return
}

func getOnlineImage(url string) (bytes []byte, fName string, err error) {
	// 清理和验证URL
	url = strings.TrimSpace(url)
	url = strings.ReplaceAll(url, "\n", "")
	url = strings.ReplaceAll(url, "\r", "")
	url = strings.ReplaceAll(url, "\t", "")

	// 验证URL格式
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, "", fmt.Errorf("invalid URL format: %s", url)
	}

	// 检查URL长度，如果太短可能有问题
	if len(url) < 12 { // https://x.x 最少需要12个字符
		return nil, "", fmt.Errorf("URL too short, possibly truncated: %s", url)
	}

	// 检查控制字符
	for _, char := range url {
		if char < 32 || char == 127 {
			return nil, "", fmt.Errorf("URL contains invalid control characters: %s", url)
		}
	}

	client := http.Client{
		Timeout: 100 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request for URL '%s': %w", url, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to download image, status code: %d, URL: %s", resp.StatusCode, url)
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

type ImageType string

const (
	ImageTypeJPEG    ImageType = "jpeg"
	ImageTypePNG     ImageType = "png"
	ImageTypeGIF     ImageType = "gif"
	ImageTypeWEBP    ImageType = "webp"
	ImageTypeUnknown ImageType = "unknown"
)

func (i ImageType) String() string {
	return string(i)
}

func (i ImageType) ImagingFormat() (imaging.Format, error) {
	if i == ImageTypeJPEG {
		return imaging.JPEG, nil
	} else if i == ImageTypePNG {
		return imaging.PNG, nil
	} else {
		return imaging.Format(-1), fmt.Errorf("unsupported image type: %s", i)
	}
}

func DetectImageType(data []byte) ImageType {
	switch http.DetectContentType(data) {
	case "image/jpeg":
		return ImageTypeJPEG
	case "image/png":
		return ImageTypePNG
	case "image/gif":
		return ImageTypeGIF
	case "image/webp":
		return ImageTypeWEBP
	default:
		return ImageTypeUnknown
	}
}
