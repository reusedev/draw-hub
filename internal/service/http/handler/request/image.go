package request

import (
	"fmt"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"mime/multipart"
	"time"
)

type UploadRequest struct {
	File   *multipart.FileHeader `form:"file" binding:"required"` // 文件字段（必须）
	ACL    string                `form:"acl"`                     // 访问控制权限，默认 "public-read"
	TTL    int                   `form:"ttl"`                     // TTL（存活时间），默认 "0"
	Expire string                `form:"expire"`                  // 过期时间，默认 "168h"
}

func (u *UploadRequest) Valid() error {
	file, err := u.File.Open()
	if err != nil {
		return err
	}
	defer file.Close()
	if u.ACL != "" && u.ACL != "public-read" && u.ACL != "private" {
		return fmt.Errorf("invalid ACL: %s, must be 'public-read' or 'private'", u.ACL)
	}
	if u.TTL < 0 {
		return fmt.Errorf("invalid TTL: %d, must be non-negative", u.TTL)
	}
	if u.Expire != "" {
		if _, err := time.ParseDuration(u.Expire); err != nil {
			return fmt.Errorf("invalid expire duration: %s", u.Expire)
		}
	}
	return nil
}

func (u *UploadRequest) FullWithDefault() {
	if u.ACL == "" {
		u.ACL = "public-read"
	}
	if u.Expire == "" {
		u.Expire = "168h" // 默认 7 天
	}
}

func (u *UploadRequest) TransformOSSUpload() (ali.UploadRequest, error) {
	file, err := u.File.Open()
	if err != nil {
		return ali.UploadRequest{}, err
	}
	defer file.Close()
	d, _ := time.ParseDuration(u.Expire)

	return ali.UploadRequest{
		Filename:  u.File.Filename,
		File:      file,
		Acl:       u.ACL,
		URLExpire: d,
	}, nil
}

type GetImageRequest struct {
	ID     int    `form:"id"`     // 图片 ID
	Type   string `form:"type"`   // 图片类型，input 或 output
	Expire string `form:"expire"` // 过期时间，默认 "168h"
}

func (g *GetImageRequest) Valid() error {
	if g.ID <= 0 {
		return fmt.Errorf("invalid ID: %d, must be greater than 0", g.ID)
	}
	if g.Type != "input" && g.Type != "output" {
		return fmt.Errorf("invalid type: %s, must be 'input' or 'output'", g.Type)
	}
	if g.Expire != "" {
		if _, err := time.ParseDuration(g.Expire); err != nil {
			return fmt.Errorf("invalid expire duration: %s", g.Expire)
		}
	}
	return nil
}

func (g *GetImageRequest) FullWithDefault() {
	if g.Expire == "" {
		g.Expire = "168h" // 默认 7 天
	}
}
