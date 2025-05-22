package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"github.com/reusedev/draw-hub/internal/service/http/response"
	"net/http"
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
	r := model.InputImage{
		StorageSupplierName: config.GConfig.StorageSupplier,
		Key:                 key,
	}
	mysql.DB.Model(&model.InputImage{}).Create(&r)

}
