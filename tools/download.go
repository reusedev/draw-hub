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
	retry := 3
label:
	retry--
	bytes, fName, err = getOnlineImage(url)
	if err != nil {
		if retry > 0 {
			time.Sleep(time.Second)
			goto label
		}
	}
	return
}

func getOnlineImage(url string) (bytes []byte, fName string, err error) {
	client := http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
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
