package ali

import (
	"github.com/reusedev/draw-hub/config"
	"strings"
	"testing"
	"time"
)

func init() {
	aliOssConfig := config.AliOss{
		AccessKeyId:     "",
		AccessKeySecret: "",
		Endpoint:        "https://oss-ap-southeast-1.aliyuncs.com",
		Region:          "ap-southeast-1",
		Bucket:          "",
		Directory:       "draw_hub/",
	}
	InitOSS(aliOssConfig)
}

func TestUpload(t *testing.T) {
	req := UploadRequest{
		Filename:  "test.txt",
		File:      strings.NewReader("123"),
		Acl:       "public-read",
		URLExpire: time.Minute,
	}
	resp, err := OssClient.UploadFile(&req)
	if err != nil {
		t.Error(err)
	}
	t.Log(resp)
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
