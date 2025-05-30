package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"net/http"
	"time"
)

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
	uploadReq, err := req.TransformOSSUpload()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	resp, err := ali.OssClient.UploadFile(&uploadReq)
	if err != nil {
		logs.Logger.Err(err).Msg("image-UploadImage")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	r := model.InputImage{
		StorageSupplierName: config.GConfig.StorageSupplier,
		Key:                 resp.Key,
		ACL:                 req.ACL,
		TTL:                 req.TTL,
		URL:                 resp.URL,
		CreatedAt:           time.Now(),
	}
	err = mysql.DB.Create(&r).Error
	if err != nil {
		logs.Logger.Err(err).Msg("image-UploadImage")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(r))
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
	var key, acl, url string
	if req.Type == "input" {
		var inputImage model.InputImage
		err = mysql.DB.Where("id = ?", req.ID).First(&inputImage).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.InternalError)
			return
		}
		key = inputImage.Key
		acl = inputImage.ACL
		url = inputImage.URL
	} else {
		var outputImage model.OutputImage
		err = mysql.DB.Where("id = ?", req.ID).First(&outputImage).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.InternalError)
			return
		}
		key = outputImage.Key
		acl = outputImage.ACL
		url = outputImage.URL
	}
	if acl == "private" {
		d, _ := time.ParseDuration(req.Expire)
		url, err = ali.OssClient.URL(key, d)
		if err != nil {
			c.JSON(http.StatusInternalServerError, response.InternalError)
			return
		}
	}
	c.JSON(http.StatusOK, response.SuccessWithData(map[string]string{"url": url}))
}
