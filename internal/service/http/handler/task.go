package handler

import (
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model/image"
	"github.com/reusedev/draw-hub/internal/modules/model/image/gpt"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/internal/service/http/model"
	"github.com/reusedev/draw-hub/tools"
	"github.com/shopspring/decimal"
	"net/http"
	"time"
)

type TaskHandler struct {
	speed      consts.TaskSpeed
	inputImage *model.InputImage
	task       *model.Task
	taskImage  *model.TaskImage
}

func (h *TaskHandler) run(form request.TaskForm) error {
	imageURL, err := ali.OssClient.URL(h.inputImage.Key, time.Hour)
	if err != nil {
		return err
	}
	err = h.createTaskRecord(form)
	if err != nil {
		return err
	}
	go func() {
		if h.speed == consts.SlowSpeed {
			editRequest := gpt.SlowRequest{
				ImageURL: imageURL,
				Prompt:   form.GetPrompt(),
			}
			editResponse := gpt.SlowSpeed(editRequest)
			err = h.endWork(editResponse)
			if err != nil {
				logs.Logger.Err(err)
			}
		} else if h.speed == consts.FastSpeed {
			editRequest := gpt.FastRequest{
				ImageURLs: []string{imageURL},
				Prompt:    form.GetPrompt(),
				Quality:   form.GetQuality(),
				Size:      form.GetSize(),
			}
			editResponse := gpt.FastSpeed(editRequest)
			err = h.endWork(editResponse)
			if err != nil {
				logs.Logger.Err(err)
			}
		}
	}()
	return nil
}
func (h *TaskHandler) createTaskRecord(form request.TaskForm) error {
	taskRecord := model.Task{
		TaskGroupId: form.GetGroupId(),
		Type:        model.TaskTypeEdit.String(),
		Prompt:      form.GetPrompt(),
		Status:      model.TaskStatusRunning.String(),
		Quality:     form.GetQuality(),
		Size:        form.GetSize(),
	}
	err := mysql.DB.Model(&model.Task{}).Create(&taskRecord).Error
	if err != nil {
		return err
	}
	h.task = &taskRecord
	taskImageRecord := model.TaskImage{
		ImageId: form.GetImageId(),
		TaskId:  taskRecord.Id,
		Type:    model.TaskImageTypeInput.String(),
	}
	err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
	if err != nil {
		return err
	}
	h.taskImage = &taskImageRecord
	return nil
}
func (h *TaskHandler) endWork(response []image.Response) error {
	for _, v := range response {
		exeRecord := model.SupplierInvokeHistory{
			TaskId:         h.task.Id,
			SupplierName:   v.GetSupplier(),
			TokenDesc:      v.GetTokenDesc(),
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
				return fmt.Errorf("not the last response, but succeed. task_id: %d", h.task.Id)
			}
			taskRecord := model.Task{
				Id:       h.task.Id,
				Model:    v.GetModel(),
				Status:   model.TaskStatusSucceed.String(),
				Progress: 100,
			}
			err := mysql.DB.Updates(&taskRecord).Error
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
				CompressionRatio:    decimal.NullDecimal{Valid: false},
				ModelSupplierName:   v.GetSupplier(),
				ModelName:           v.GetModel(),
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
			Id:     h.task.Id,
			Status: model.TaskStatusFailed.String(),
		}
		err := mysql.DB.Updates(&taskRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}
func (h *TaskHandler) list(groupId, id string) ([]model.Task, error) {
	var tasks []model.Task
	query := mysql.DB.Model(&model.Task{}).
		Preload("TaskImages").
		Preload("TaskImages.InputImage").
		Preload("TaskImages.OutputImage")
	if groupId != "" {
		query = query.Where("task_group_id = ?", groupId)
	}
	if id != "" {
		query = query.Where("id = ?", id)
	}
	err := query.Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func SlowSpeed(c *gin.Context) {
	form := request.SlowTask{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	var inputImage model.InputImage
	err = mysql.DB.Model(&model.InputImage{}).Where("id = ?", form.ImageId).First(&inputImage).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h := TaskHandler{speed: consts.SlowSpeed, inputImage: &inputImage}
	err = h.run(&form)

	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(h.task))
}

func FastSpeed(c *gin.Context) {
	form := request.FastSpeed{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	var inputImage model.InputImage
	err = mysql.DB.Model(&model.InputImage{}).Where("id = ?", form.ImageId).First(&inputImage).Error
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h := TaskHandler{speed: consts.FastSpeed, inputImage: &inputImage}
	err = h.run(&form)

	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(h.task))
}

func TaskQuery(c *gin.Context) {
	id := c.Query("id")
	groupId := c.Query("group_id")
	h := TaskHandler{}
	tasks, err := h.list(groupId, id)
	if err != nil {
		logs.Logger.Err(err)
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(tasks))
}
