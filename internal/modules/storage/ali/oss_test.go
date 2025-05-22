package ali

import (
	"github.com/reusedev/draw-hub/config"
	"strings"
	"testing"
)

func TestUpload(t *testing.T) {
	aliOssConfig := config.AliOss{}
	InitOSS(aliOssConfig)
	err := OssClient.upload("test.txt", "cloud_test/draw_hub/123213.txt", strings.NewReader("123"))
	if err != nil {
		t.Error(err)
	}
}
