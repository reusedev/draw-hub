package handler

import (
	"database/sql"
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
	"net/http"
	"time"
)

type TaskHandler struct {
	speed         consts.TaskSpeed
	inputImage    *model.InputImage
	task          *model.Task
	taskImage     *model.TaskImage
	imageResponse []image.Response
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
			h.imageResponse = gpt.SlowSpeed(editRequest)
			err = h.endWork()
			if err != nil {
				logs.Logger.Err(err).Msg("task-SlowSpeed")
			}
		} else if h.speed == consts.FastSpeed {
			editRequest := gpt.FastRequest{
				ImageURLs: []string{imageURL},
				Prompt:    form.GetPrompt(),
				Quality:   form.GetQuality(),
				Size:      form.GetSize(),
			}
			h.imageResponse = gpt.FastSpeed(editRequest)
			err = h.endWork()
			if err != nil {
				logs.Logger.Err(err).Msg("task-FastSpeed")
			}
		}
	}()
	return nil
}
func (h *TaskHandler) createTaskRecord(form request.TaskForm) error {
	now := time.Now()
	taskRecord := model.Task{
		TaskGroupId: form.GetGroupId(),
		Type:        model.TaskTypeEdit.String(),
		Prompt:      form.GetPrompt(),
		Status:      model.TaskStatusRunning.String(),
		Quality:     form.GetQuality(),
		Size:        form.GetSize(),
		CreatedAt:   now,
		UpdatedAt:   now,
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
func (h *TaskHandler) createImageRecords(imageResp image.Response) error {
	err := h.createNormalRecords(imageResp)
	if err != nil {
		return err
	}
	err = h.createCompressionRecords(imageResp)
	if err != nil {
		return err
	}
	return nil
}
func (h *TaskHandler) createNormalRecords(imageResp image.Response) error {
	for _, v := range imageResp.GetURLs() {
		normal, err := uploadNormalImage(v)
		if err != nil {
			return err
		}
		imageRecord := model.OutputImage{
			StorageSupplierName: config.GConfig.StorageSupplier,
			Key:                 normal.Key,
			ACL:                 "private",
			TTL:                 0,
			URL:                 normal.URL,
			Type:                string(model.OuputImageTypeNormal),
			ModelSupplierURL:    v,
			ModelSupplierName:   imageResp.GetSupplier(),
			ModelName:           imageResp.GetModel(),
		}
		err = mysql.DB.Model(&model.OutputImage{}).Create(&imageRecord).Error
		if err != nil {
			return err
		}
		taskImageRecord := model.TaskImage{
			TaskId:  h.task.Id,
			ImageId: imageRecord.Id,
			Type:    model.TaskImageTypeOutput.String(),
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) createCompressionRecords(imageResp image.Response) error {
	for _, v := range imageResp.GetURLs() {
		compression, ratio, err := uploadCompressionImage(v, 95)
		if err != nil {
			return err
		}
		imageRecord := model.OutputImage{
			StorageSupplierName: config.GConfig.StorageSupplier,
			Key:                 compression.Key,
			ACL:                 "private",
			TTL:                 0,
			URL:                 compression.URL,
			Type:                string(model.OuputImageTypeCompressed),
			CompressionRatio:    sql.NullFloat64{Valid: true, Float64: ratio},
			ModelSupplierURL:    v,
			ModelSupplierName:   imageResp.GetSupplier(),
			ModelName:           imageResp.GetModel(),
		}
		err = mysql.DB.Model(&model.OutputImage{}).Create(&imageRecord).Error
		if err != nil {
			return err
		}
		taskImageRecord := model.TaskImage{
			TaskId:  h.task.Id,
			ImageId: imageRecord.Id,
			Type:    model.TaskImageTypeOutput.String(),
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) endWork() error {
	for _, v := range h.imageResponse {
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
	for i, v := range h.imageResponse {
		if v.Succeed() {
			succeed = true
			if i != len(h.imageResponse)-1 {
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
			err = h.createImageRecords(v)
			if err != nil {
				return err
			}
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
	for t := range tasks {
		for i := range tasks[t].TaskImages {
			if tasks[t].TaskImages[i].Type == model.TaskImageTypeInput.String() {
				tasks[t].TaskImages[i].OutputImage = model.OutputImage{}
			}
			if tasks[t].TaskImages[i].Type == model.TaskImageTypeOutput.String() {
				tasks[t].TaskImages[i].InputImage = model.InputImage{}
			}
		}
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
		logs.Logger.Err(err).Msg("task-SlowSpeed")
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h := TaskHandler{speed: consts.SlowSpeed, inputImage: &inputImage}
	err = h.run(&form)

	if err != nil {
		logs.Logger.Err(err).Msg("task-SlowSpeed")
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
		logs.Logger.Err(err).Msg("task-FastSpeed")
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h := TaskHandler{speed: consts.FastSpeed, inputImage: &inputImage}
	err = h.run(&form)

	if err != nil {
		logs.Logger.Err(err).Msg("task-FastSpeed")
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
		logs.Logger.Err(err).Msg("task-TaskQuery")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithData(tasks))
}

func uploadNormalImage(supplierURL string) (normal ali.OSSObject, err error) {
	bytes, _, err := tools.GetOnlineImage(supplierURL)
	if err != nil {
		return
	}
	key, err := ali.OssClient.UploadPrivateImage(bytes)
	if err != nil {
		return
	}
	// 配置初始化时已校验
	duration, _ := time.ParseDuration(config.GConfig.URLExpires)
	presignRet, err := ali.OssClient.Presign(key, duration)
	if err != nil {
		return
	}
	normal.URLExpire = &presignRet.Expiration
	normal.URL = presignRet.URL
	normal.Key = key
	return
}

func uploadCompressionImage(supplierURL string, quality int) (compression ali.OSSObject, ratio float64, err error) {
	bytes, _, err := tools.GetOnlineImage(supplierURL)
	if err != nil {
		return
	}
	compressionBytes, err := tools.ConvertAndCompressPNGtoJPEG(bytes, quality)
	if err != nil {
		return
	}
	ratio = float64(len(compressionBytes)) / float64(len(bytes))
	key, err := ali.OssClient.UploadPrivateImage(compressionBytes)
	if err != nil {
		return
	}
	// 配置初始化时已校验
	duration, _ := time.ParseDuration(config.GConfig.URLExpires)
	presignRet, err := ali.OssClient.Presign(key, duration)
	if err != nil {
		return
	}
	compression.URLExpire = &presignRet.Expiration
	compression.URL = presignRet.URL
	compression.Key = key
	return
}
