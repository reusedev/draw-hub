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
	"strings"
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

type UploadRequest struct {
	Key       string        `json:"key,omitempty"` // Optional, if provided, will use this as the object key
	Filename  string        `json:"filename"`
	File      io.Reader     `json:"file"`
	Acl       string        `json:"acl"`
	URLExpire time.Duration `json:"url_expire,omitempty"`
}
type OSSObject struct {
	Key       string     `json:"key"`
	URL       string     `json:"url"`
	URLExpire *time.Time `json:"url_expire,omitempty"`
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

func (o *ossClient) validACL(acl string) bool {
	if oss.ObjectACLType(acl) != oss.ObjectACLPrivate &&
		oss.ObjectACLType(acl) != oss.ObjectACLPublicRead {
		return false
	}
	return true
}

func (o *ossClient) UploadFile(request *UploadRequest) (OSSObject, error) {
	ret := OSSObject{}
	var key string
	if request.Key != "" {
		key = request.Key
	} else {
		ext := filepath.Ext(request.Filename)
		key = o.fullPath(uuid.New().String() + ext)
	}
	ret.Key = key

	err := o.upload(request.Filename, key, request.Acl, request.File)
	if err != nil {
		return ret, err
	}
	if oss.ObjectACLType(request.Acl) == oss.ObjectACLPublicRead {
		ret.URL = "https://" + o.bucketName + "." + strings.TrimPrefix(o.endpoint, "https://") + "/" + key
		ret.URLExpire = nil
		return ret, nil
	}
	if request.URLExpire <= 0 {
		request.URLExpire = time.Hour * 24 * 7 // default 7 days
	}
	presignRet, err := o.Presign(key, request.URLExpire)
	if err != nil {
		return ret, err
	}
	ret.URL = presignRet.URL
	ret.URLExpire = &presignRet.Expiration
	return ret, nil
}

func (o *ossClient) UploadPrivateImage(b []byte) (string, error) {
	fName := uuid.New().String() + "." + tools.DetectImageType(b).String()
	key := o.fullPath(fName)
	return key, o.upload(fName, key, string(oss.ObjectACLPrivate), bytes.NewReader(b))
}

func (o *ossClient) URL(key string, expire time.Duration) (string, error) {
	presignResult, err := o.Presign(key, expire)
	if err != nil {
		return "", err
	}
	return presignResult.URL, nil
}

func (o *ossClient) fullPath(fName string) string {
	return o.directory + fName
}

func (o *ossClient) Presign(key string, expire time.Duration) (*oss.PresignResult, error) {
	request := &oss.GetObjectRequest{
		Bucket: oss.Ptr(o.bucketName),
		Key:    oss.Ptr(key),
	}
	return o.client.Presign(context.TODO(), request, oss.PresignExpires(expire))
}

func (o *ossClient) Resize50(key string, expire time.Duration) (*oss.PresignResult, error) {
	request := &oss.GetObjectRequest{
		Bucket:  oss.Ptr(o.bucketName),
		Key:     oss.Ptr(key),
		Process: oss.Ptr("image/resize,p_50"),
	}
	return o.client.Presign(context.TODO(), request, oss.PresignExpires(expire))
}

func (o *ossClient) upload(fName, key, acl string, reader io.Reader) error {
	request := &oss.PutObjectRequest{
		Bucket:             oss.Ptr(o.bucketName),
		Acl:                oss.ObjectACLType(acl),
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
