package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/dao"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/modules/storage/local"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

type StorageHandler struct {
	inputImage  model.InputImage
	outputImage model.OutputImage
}

func NewStorageHandler() *StorageHandler {
	return &StorageHandler{}
}
func (s *StorageHandler) InputImage() model.InputImage {
	result := s.inputImage
	if result.URL == "" {
		result.URL = config.GConfig.LocalStorageDomain + "/" + strings.ReplaceAll(result.Path, string(filepath.Separator), "/")
	}
	return result
}
func (s *StorageHandler) OutputImage() model.OutputImage {
	return s.outputImage
}
func (s *StorageHandler) Upload(request request.UploadRequest) error {
	s.initInputImage(request)
	err := s.upload(request)
	if err != nil {
		return err
	}
	err = s.createInputImage()
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageHandler) Query(request request.GetImageRequest) (response.GetImage, error) {
	err := s.selectImage(request)
	if err != nil {
		return response.GetImage{}, err
	}
	return s.getImageResponse(request)
}

func (s *StorageHandler) selectImage(request request.GetImageRequest) error {
	if request.Type == "input" {
		inputImage, err := dao.InputImageById(request.ID)
		if err != nil {
			return err
		}
		s.inputImage = inputImage
	} else {
		outputImage, err := dao.OutputImageById(request.ID)
		if err != nil {
			return err
		}
		s.outputImage = outputImage
	}
	return nil
}
func (s *StorageHandler) getImageResponse(request request.GetImageRequest) (response.GetImage, error) {
	var ret response.GetImage
	var key, acl, url string
	if request.Type == "input" {
		ret.Path = s.inputImage.Path
		key = s.inputImage.Key
		acl = s.inputImage.ACL
		url = s.inputImage.URL
	} else {
		ret.Path = s.outputImage.Path
		key = s.outputImage.Key
		acl = s.outputImage.ACL
		url = s.outputImage.URL
	}
	if !request.ThumbNail && request.Type == "output" && strings.HasSuffix(s.outputImage.ModelSupplierURL, ".png") && s.outputImage.Type == model.OuputImageTypeNormal.String() {
		ret.URL = s.outputImage.ModelSupplierURL
		return ret, nil
	}
	if config.GConfig.CloudStorageEnabled {
		if request.ThumbNail {
			d, _ := time.ParseDuration(config.GConfig.URLExpires)
			ossURL, err := ali.OssClient.Resize50(key, d)
			if err != nil {
				return ret, err
			}
			ret.URL = ossURL.URL
			return ret, nil
		}
		if acl == "private" {
			d, _ := time.ParseDuration(config.GConfig.URLExpires)
			ossURL, err := ali.OssClient.URL(key, d)
			if err != nil {
				return ret, err
			}
			ret.URL = ossURL
			return ret, nil
		}
		ret.URL = url
		return ret, nil
	}
	if request.ThumbNail {
		if s.outputImage.ThumbNailPath != "" {
			ret.Path = s.outputImage.ThumbNailPath
			ret.URL = config.GConfig.LocalStorageDomain + "/" + strings.ReplaceAll(ret.Path, string(filepath.Separator), "/")
			return ret, nil
		}
	}
	ret.URL = config.GConfig.LocalStorageDomain + "/" + strings.ReplaceAll(ret.Path, string(filepath.Separator), "/")
	return ret, nil
}

func (s *StorageHandler) initInputImage(request request.UploadRequest) {
	now := time.Now()
	s.inputImage.TTL = request.TTL
	s.inputImage.CreatedAt = now
	s.inputImage.Path = s.localStoragePath(request.File.Filename, now)
	if config.GConfig.CloudStorageEnabled {
		s.inputImage.Key = s.ossStorageKey(request.File.Filename)
		s.inputImage.ACL = request.ACL
		s.inputImage.StorageSupplierName = config.GConfig.CloudStorageSupplier
	}
}

func (s *StorageHandler) upload(request request.UploadRequest) error {
	err := s.localSave(request)
	if err != nil {
		return err
	}
	if config.GConfig.CloudStorageEnabled {
		err = s.uploadToOSS(request)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *StorageHandler) uploadToOSS(request request.UploadRequest) error {
	ossReq, err := s.transformToOssUpload(request)
	if err != nil {
		return err
	}
	ossObject, err := ali.OssClient.UploadFile(&ossReq)
	if err != nil {
		return err
	}
	s.inputImage.URL = ossObject.URL
	return nil
}
func (s *StorageHandler) localSave(request request.UploadRequest) error {
	f, err := request.File.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	absPath := filepath.Join(config.GConfig.LocalStorageDirectory, s.inputImage.Path)
	err = local.SaveFile(f, absPath)
	if err != nil {
		return err
	}
	return nil
}
func (s *StorageHandler) localStoragePath(filename string, t time.Time) string {
	ext := filepath.Ext(filename)
	path := filepath.Join("input", t.Format("20060102"), uuid.New().String()+ext)
	return path
}
func (s *StorageHandler) ossStorageKey(filename string) string {
	ext := filepath.Ext(filename)
	key := config.GConfig.AliOss.Directory + uuid.New().String() + ext
	return key
}

func (s *StorageHandler) createInputImage() error {
	err := mysql.DB.Create(&s.inputImage).Error
	if err != nil {
		return err
	}
	return nil
}

func (s *StorageHandler) transformToOssUpload(request request.UploadRequest) (ali.UploadRequest, error) {
	f, err := request.File.Open()
	if err != nil {
		return ali.UploadRequest{}, err
	}
	urlExpire, _ := time.ParseDuration(config.GConfig.URLExpires)
	ret := ali.UploadRequest{
		Key:       s.inputImage.Key,
		Filename:  request.File.Filename,
		File:      f,
		Acl:       s.inputImage.ACL,
		URLExpire: urlExpire,
	}
	return ret, nil
}

func UploadImage(c *gin.Context) {
	var req request.UploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	err := req.Valid()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	req.FullWithDefault()
	handler := NewStorageHandler()
	err = handler.Upload(req)
	if err != nil {
		logs.Logger.Err(err).Msg("Image-Upload")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(handler.InputImage()))
}

func GetImage(c *gin.Context) {
	var req request.GetImageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	err := req.Valid()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	req.FullWithDefault()
	handler := NewStorageHandler()
	resp, err := handler.Query(req)
	if err != nil {
		logs.Logger.Err(err).Msg("Image-Get")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(resp))
}
