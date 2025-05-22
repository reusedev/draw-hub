package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"net/http"
	"time"
)

func UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	defer file.Close()
	key, err := ali.OssClient.UploadFile(header.Filename, file)
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	expires, _ := time.ParseDuration(config.GConfig.URLExpires)
	url, err := ali.OssClient.URL(key, expires)
	r := model.InputImage{
		StorageSupplierName: config.GConfig.StorageSupplier,
		Key:                 key,
		URL:                 url,
		CreatedAt:           time.Now(),
	}
	err = mysql.DB.Create(&r).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(r))
}
