package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai_model"
	"github.com/reusedev/draw-hub/internal/modules/ai_model/image"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"github.com/reusedev/draw-hub/tools"
	"net/http"
	"time"
)

func SLowSpeed(c *gin.Context) {
	form := request.SlowTask{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	var imageRecord model.InputImage
	err = mysql.DB.Model(&model.InputImage{}).Where("id = ?", form.ImageId).First(&imageRecord).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	imageURL, err := ali.OssClient.URL(imageRecord.Key, time.Hour)
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	taskRecord := model.Task{
		TaskGroupId: form.GroupId,
		Type:        model.TaskTypeEdit.String(),
		Prompt:      form.Prompt,
		Status:      model.TaskStatusRunning.String(),
	}
	err = mysql.DB.Model(&model.Task{}).Create(&taskRecord).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	taskImageRecord := model.TaskImage{
		ImageId: form.ImageId,
		TaskId:  taskRecord.Id,
		Type:    model.TaskImageTypeInput.String(),
	}
	err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	go func() {
		editRequest := image.SlowRequest{
			ImageURL: imageURL,
			Prompt:   form.Prompt,
		}
		editResponse := image.SlowSpeed(editRequest)
		err = exeEndWork(editResponse, taskRecord)
		if err != nil {
			logs.Logger.Err(err)
		}
	}()
	c.JSON(http.StatusOK, response.SuccessWithData(taskRecord))
}

func exeEndWork(response []ai_model.Response, task model.Task) error {
	for _, v := range response {
		exeRecord := model.SupplierInvokeHistory{
			TaskId:         task.Id,
			SupplierName:   v.GetSupplier(),
			ModelName:      v.GetModel(),
			StatusCode:     v.GetStatusCode(),
			FailedRespBody: v.FailedRespBody(),
			DurationMs:     v.DurationMs(),
			CreatedAt:      v.GetRespAt(),
		}
		err := mysql.DB.Model(&model.SupplierInvokeHistory{}).Create(&exeRecord).Error
		if err != nil {
			return err
		}
	}

	var succeed bool
	for i, v := range response {
		if v.Succeed() {
			succeed = true
			if i != len(response)-1 {
				return fmt.Errorf("not the last response, but succeed. task_id: %d", task.Id)
			}
			taskRecord := model.Task{
				Id:       task.Id,
				Model:    v.GetModel(),
				Status:   model.TaskStatusSucceed.String(),
				Progress: 100,
			}
			err := mysql.DB.Model(&model.Task{}).Updates(&taskRecord).Error
			if err != nil {
				return err
			}
			var imgBytes []byte
			if v.GetModel() == consts.GPTImage1.String() {
				b64 := v.GetBase64()
				imgBytes, err = base64.StdEncoding.DecodeString(b64)
				if err != nil {
					return err
				}
			} else if v.GetModel() == consts.GPT4oImage.String() || v.GetModel() == consts.GPT4oImageVip.String() {
				URLs := v.GetURLs()
				if len(URLs) == 0 {
					return fmt.Errorf("no image URL found in response %s", v.FailedRespBody())
				}
				// todo 生成多张，保存多张
				imgBytes, _, err = tools.GetOnlineImage(URLs[0])
				if err != nil {
					return err
				}
			} else {
				return fmt.Errorf("unknown model %s", v.GetModel())
			}

			key, err := ali.OssClient.UploadImage(imgBytes)
			if err != nil {
				return err
			}
			duration, _ := time.ParseDuration(config.GConfig.URLExpires)
			url, err := ali.OssClient.URL(key, duration)
			if err != nil {
				return err
			}
			imageRecord := model.OutputImage{
				StorageSupplierName: config.GConfig.StorageSupplier,
				Key:                 key,
				URL:                 url,
				Type:                model.OuputImageTypeNormal.String(),
			}
			if v.GetModel() == consts.GPT4oImageVip.String() || v.GetModel() == consts.GPT4oImage.String() {
				imageRecord.OriginalURL = v.GetURLs()[0]
			}
			err = mysql.DB.Model(&model.OutputImage{}).Create(&imageRecord).Error
			if err != nil {
				return err
			}
			taskImageRecord := model.TaskImage{
				TaskId:  taskRecord.Id,
				ImageId: imageRecord.Id,
				Type:    model.TaskImageTypeOutput.String(),
			}
			err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
			if err != nil {
				return err
			}
			// todo 保存压缩图片
		}
	}
	// todo 总结失败原因
	if !succeed {
		taskRecord := model.Task{
			Id:     task.Id,
			Status: model.TaskStatusFailed.String(),
		}
		err := mysql.DB.Model(&model.Task{}).Updates(&taskRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func FastSpeed(c *gin.Context) {
	form := request.FastSpeed{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
}
