package ali

import (
	"bytes"
	"context"
	"fmt"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
	"github.com/google/uuid"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/tools"
	"io"
	"path/filepath"
	"time"
)

var (
	OssClient *ossClient
)

type ossClient struct {
	client     *oss.Client
	endpoint   string
	bucketName string
	directory  string
}

func InitOSS(config config.AliOss) {
	credential := credentials.NewStaticCredentialsProvider(config.AccessKeyId, config.AccessKeySecret, "")
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credential).
		WithEndpoint(config.Endpoint).WithRegion(config.Region)
	client := oss.NewClient(cfg)
	if client == nil {
		panic("create oss client failed")
	}
	OssClient = &ossClient{
		client:     client,
		endpoint:   config.Endpoint,
		bucketName: config.Bucket,
		directory:  config.Directory,
	}
}

func (o *ossClient) UploadFileWithName(fName string, file io.Reader) (string, error) {
	ext := filepath.Ext(fName)
	key := o.fullPath(uuid.New().String() + ext)
	return key, o.upload(fName, key, file)
}

func (o *ossClient) UploadImage(b []byte) (string, error) {
	fName := uuid.New().String() + "." + tools.DetectImageType(b)
	key := o.fullPath(fName)
	return key, o.upload(fName, key, bytes.NewReader(b))
}

func (o *ossClient) URL(key string, expire time.Duration) (string, error) {
	ret, err := o.client.Presign(context.TODO(), &oss.GetObjectRequest{Bucket: oss.Ptr(o.bucketName), Key: oss.Ptr(key)}, oss.PresignExpires(expire))
	if err != nil {
		return "", err
	}
	return ret.URL, nil
}

func (o *ossClient) fullPath(fName string) string {
	return o.directory + fName
}

func (o *ossClient) upload(fName, key string, reader io.Reader) error {
	request := &oss.PutObjectRequest{
		Bucket:             oss.Ptr(o.bucketName),
		Key:                oss.Ptr(key),
		Body:               reader,
		ContentDisposition: oss.Ptr(fmt.Sprintf("attachment; filename=\"%s\"", fName)),
	}
	_, err := o.client.PutObject(context.TODO(), request)
	if err != nil {
		return err
	}
	return nil
}
