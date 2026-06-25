package ali

import (
	"github.com/reusedev/draw-hub/config"
	"os"
	"strings"
	"testing"
	"time"
)

func init() {
	aliOssConfig := config.AliOss{
		AccessKeyId:     "test-access-key",
		AccessKeySecret: "test-access-secret",
		Endpoint:        "https://oss-ap-southeast-1.aliyuncs.com",
		Region:          "ap-southeast-1",
		Bucket:          "test-bucket",
		Directory:       "draw_hub/",
	}
	InitOSS(aliOssConfig)
}

func TestUpload(t *testing.T) {
	accessKeyID := os.Getenv("ALI_OSS_ACCESS_KEY_ID")
	accessKeySecret := os.Getenv("ALI_OSS_ACCESS_KEY_SECRET")
	bucket := os.Getenv("ALI_OSS_BUCKET")
	if accessKeyID == "" || accessKeySecret == "" || bucket == "" {
		t.Skip("set ALI_OSS_ACCESS_KEY_ID, ALI_OSS_ACCESS_KEY_SECRET, and ALI_OSS_BUCKET to run live OSS upload test")
	}
	InitOSS(config.AliOss{
		AccessKeyId:     accessKeyID,
		AccessKeySecret: accessKeySecret,
		Endpoint:        getenvDefault("ALI_OSS_ENDPOINT", "https://oss-ap-southeast-1.aliyuncs.com"),
		Region:          getenvDefault("ALI_OSS_REGION", "ap-southeast-1"),
		Bucket:          bucket,
		Directory:       getenvDefault("ALI_OSS_DIRECTORY", "draw_hub/"),
	})

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

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func TestSignURL(t *testing.T) {
	key := "draw_hub/f0952f59-b6be-4f63-b5bb-b956c44ef4c7.jpeg"
	url, err := OssClient.URL(key, time.Minute)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("PreSign URL: %s", url)
	}
}

func TestThumbNailURL(t *testing.T) {
	key := "draw_hub/f0952f59-b6be-4f63-b5bb-b956c44ef4c7.jpeg"
	url, err := OssClient.Resize50(key, time.Minute)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("Thumbnail URL: %s", url)
	}
}
