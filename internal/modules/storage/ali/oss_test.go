package ali

import (
	"github.com/reusedev/draw-hub/config"
	"strings"
	"testing"
	"time"
)

func init() {
	aliOssConfig := config.AliOss{}
	InitOSS(aliOssConfig)
}

func TestUpload(t *testing.T) {
	err := OssClient.upload("test.txt", "cloud_test/draw_hub/test.txt", strings.NewReader("123"))
	if err != nil {
		t.Error(err)
	}
}

func TestSignURL(t *testing.T) {
	key := "cloud_test/draw_hub/test.txt"
	url, err := OssClient.URL(key, time.Minute)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("PreSign URL: %s", url)
	}
}
