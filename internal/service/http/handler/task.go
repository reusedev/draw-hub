package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/reusedev/draw-hub/internal/modules/ai/image/gemini"
	"github.com/reusedev/draw-hub/internal/modules/ai/image/mj"
	"github.com/reusedev/draw-hub/internal/modules/ai/image/volc"
	"github.com/reusedev/draw-hub/internal/modules/observer"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/consts"
	"github.com/reusedev/draw-hub/internal/modules/ai/image"
	"github.com/reusedev/draw-hub/internal/modules/ai/image/gpt"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model"
	"github.com/reusedev/draw-hub/internal/modules/queue"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/modules/storage/local"
	"github.com/reusedev/draw-hub/internal/service/http/handler/request"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/tools"
)

type TaskHandler struct {
	ctx           *gin.Context
	task          *model.Task
	imageResponse []image.Response
	wg            *sync.WaitGroup
}

func newTaskHandler(c *gin.Context) (*TaskHandler, error) {
	h := &TaskHandler{ctx: c}
	return h, nil
}

func EnqueueUnfinishedTask() error {
	tasks := make([]model.Task, 0)
	err := mysql.DB.Model(&model.Task{}).
		Preload("TaskImages").
		Preload("TaskImages.InputImage").
		Preload("TaskImages.OutputImage").
		Where("status = ?", model.TaskStatusAborted.String()).Find(&tasks).Error
	if err != nil {
		return err
	}
	for _, task := range tasks {
		h := TaskHandler{task: &task}
		h.enqueue()
		logs.Logger.Info().Int("task_id", task.Id).Msg("Re-enqueued task")
	}
	return nil
}

func (h *TaskHandler) enqueue() {
	mysql.DB.Model(&model.Task{}).Where("id = ?", h.task.Id).Updates(map[string]interface{}{
		"status": model.TaskStatusQueued.String(),
	})
	queue.ImageTaskQueue <- h
}

func (h *TaskHandler) fail(err error) {
	// 记录任务异常失败日志
	logs.Logger.Error().
		Err(err).
		Int("task_id", h.task.Id).
		Str("model", h.task.Model).
		Str("status", "failed").
		Msg("Task execution failed with exception")

	mysql.DB.Model(&model.Task{}).Where("id = ?", h.task.Id).Updates(map[string]interface{}{
		"status":        model.TaskStatusFailed.String(),
		"failed_reason": "任务执行失败，请稍后重试",
	})
}

func (h *TaskHandler) Execute(ctx context.Context, wg *sync.WaitGroup) {
	// 记录任务开始执行日志
	logs.Logger.Info().
		Int("task_id", h.task.Id).
		Str("task_type", h.task.Type).
		Str("model", h.task.Model).
		Str("status", "started").
		Msg("Task execution started")
	mysql.DB.Model(&model.Task{}).Where("id = ?", h.task.Id).Updates(map[string]interface{}{
		"status": model.TaskStatusRunning.String(),
	})
	h.wg = wg
	switch h.task.Type {
	case consts.TaskTypeEdit.String():
		h.edit(ctx)
	case consts.TaskTypeGenerate.String():
		h.generate(ctx)
	}
}

func (h *TaskHandler) edit(ctx context.Context) {
	bs, err := h.inputImageBytes()
	if err != nil {
		h.fail(err)
		return
	}
	urls, _ := h.inputImageURLs()
	logs.Logger.Info().
		Int("task_id", h.task.Id).
		Str("model", h.task.Model).
		Msg("Calling image supplier")
	if h.task.Speed.Valid && h.task.Speed.String == consts.SlowSpeed.String() {
		editRequest := gpt.SlowRequest{
			ImageBytes: bs,
			Prompt:     h.task.Prompt,
			Model:      h.task.Model,
			TaskID:     h.task.Id,
		}
		gpt.NewProvider(ctx, []observer.Observer{h}).SlowSpeed(editRequest)
	} else if h.task.Speed.Valid && h.task.Speed.String == consts.FastSpeed.String() {
		editRequest := gpt.FastRequest{
			ImageBytes: bs,
			Prompt:     h.task.Prompt,
			Quality:    h.task.Quality,
			Size:       h.task.Size,
			TaskID:     h.task.Id,
		}
		gpt.NewProvider(ctx, []observer.Observer{h}).FastSpeed(editRequest)
	} else if strings.HasPrefix(h.task.Model, "gemini") {
		req := gemini.Request{
			ImageBytes: bs,
			Prompt:     h.task.Prompt,
			Model:      h.task.Model,
			TaskID:     h.task.Id,
		}
		gemini.NewProvider(ctx, []observer.Observer{h}).Create(req)
	} else if strings.HasPrefix(h.task.Model, "jimeng") {
		req := volc.Request{
			ImageURLs:  urls,
			ImageBytes: bs,
			Prompt:     h.task.Prompt,
			Size:       h.task.Size,
			TaskID:     h.task.Id,
		}
		volc.NewProvider(ctx, []observer.Observer{h}).Create(req)
	} else if h.task.Model == consts.MidJourney.String() {
		req := mj.Request{
			ImageBytes: bs,
			Prompt:     h.task.Prompt,
			TaskID:     h.task.Id,
		}
		mj.NewProvider(ctx, []observer.Observer{h}).Create(req)
	}
}

func (h *TaskHandler) generate(ctx context.Context) {
	logs.Logger.Info().
		Int("task_id", h.task.Id).
		Str("model", h.task.Model).
		Msg("Calling image supplier")
	if strings.HasPrefix(h.task.Model, "gemini") {
		req := gemini.Request{
			Prompt: h.task.Prompt,
			Model:  h.task.Model,
			TaskID: h.task.Id, // 传递TaskID
		}
		gemini.NewProvider(ctx, []observer.Observer{h}).Create(req)
	} else if strings.HasPrefix(h.task.Model, "jimeng") {
		req := volc.Request{
			Prompt: h.task.Prompt,
			Size:   h.task.Size,
			TaskID: h.task.Id,
		}
		volc.NewProvider(ctx, []observer.Observer{h}).Create(req)
	} else if h.task.Model == consts.MidJourney.String() {
		req := mj.Request{
			Prompt: h.task.Prompt,
			TaskID: h.task.Id,
		}
		mj.NewProvider(ctx, []observer.Observer{h}).Create(req)
	} else {
		genRequest := gpt.SlowRequest{
			Prompt: h.task.Prompt,
			Model:  h.task.Model,
			TaskID: h.task.Id, // 传递TaskID
		}
		gpt.NewProvider(ctx, []observer.Observer{h}).SlowSpeed(genRequest)
	}
}

func (h *TaskHandler) inputImageBytes() (ret [][]byte, err error) {
	for _, img := range h.task.TaskImages {
		if img.Type != model.TaskImageTypeInput.String() {
			continue
		}
		var path, key string
		if img.Origin.String == model.TaskImageOriginOutput.String() {
			path = img.OutputImage.Path
			key = img.OutputImage.Key
		} else {
			path = img.InputImage.Path
			key = img.InputImage.Key
		}
		b, err := tools.ReadFile(filepath.Join(config.GConfig.LocalStorageDirectory, path))
		if err != nil {
			logs.Logger.Err(err).Msg("Read-LocalFile")
		} else {
			ret = append(ret, b)
			continue
		}

		if !config.GConfig.CloudStorageEnabled {
			return nil, fmt.Errorf("cloud storage is not enabled, cannot get input image bytes")
		}
		url, err := ali.OssClient.URL(key, time.Hour)
		if err != nil {
			logs.Logger.Err(err).Msg("Get-OSS-URL")
			return nil, err
		}
		b, _, err = tools.GetOnlineImage(url)
		if err != nil {
			logs.Logger.Err(err).Msg("Get-OnlineImage")
			return nil, err
		}
		ret = append(ret, b)
	}
	return
}

func (h *TaskHandler) inputImageURLs() ([]string, error) {
	ret := make([]string, 0)
	for _, img := range h.task.TaskImages {
		if img.Type != model.TaskImageTypeInput.String() {
			continue
		}
		var key string
		if img.Origin.String == model.TaskImageOriginOutput.String() {
			key = img.OutputImage.Key
		} else {
			key = img.InputImage.Key
		}
		if !config.GConfig.CloudStorageEnabled {
			return nil, fmt.Errorf("cloud storage is not enabled, cannot get input image bytes")
		}
		url, err := ali.OssClient.URL(key, time.Hour)
		if err != nil {
			logs.Logger.Err(err).Msg("Get-OSS-URL")
			return nil, err
		}
		ret = append(ret, url)
	}
	return ret, nil
}

func (h *TaskHandler) createEditTask(form request.TaskForm) error {
	now := time.Now()
	taskRecord := model.Task{
		TaskGroupId: form.GetGroupId(),
		Type:        consts.TaskTypeEdit.String(),
		Prompt:      form.GetPrompt(),
		Speed:       sql.NullString{Valid: true, String: form.GetSpeed().String()},
		Status:      model.TaskStatusPending.String(),
		Quality:     form.GetQuality(),
		Size:        form.GetSize(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if h.ctx.FullPath() == "/v2/task/slow/4oVip-four" {
		taskRecord.Model = consts.GPT4oImageVip.String()
	}
	err := mysql.DB.Model(&model.Task{}).Create(&taskRecord).Error
	if err != nil {
		return err
	}
	for _, ii := range form.GetImageIds() {
		taskImageR := model.TaskImage{
			ImageId: ii,
			TaskId:  taskRecord.Id,
			Type:    model.TaskImageTypeInput.String(),
		}
		if form.GetImageOrigin() != "" {
			taskImageR.Origin = sql.NullString{Valid: true, String: form.GetImageOrigin()}
		} else {
			taskImageR.Origin = sql.NullString{Valid: false}
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageR).Error
		if err != nil {
			return err
		}
	}
	var task model.Task
	err = mysql.DB.Model(&model.Task{}).
		Preload("TaskImages").
		Preload("TaskImages.InputImage").
		Preload("TaskImages.OutputImage").
		Where("id = ?", taskRecord.Id).First(&task).Error
	if err != nil {
		return err
	}
	h.task = &task
	return nil
}

func (h *TaskHandler) createGenerateTask(form *request.Generate) error {
	now := time.Now()
	task := model.Task{
		TaskGroupId: form.GroupId,
		Type:        consts.TaskTypeGenerate.String(),
		Prompt:      form.Prompt,
		Status:      model.TaskStatusPending.String(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if h.ctx.FullPath() == "/v2/task/generate/4oVip-four" {
		task.Model = consts.GPT4oImageVip.String()
	}
	err := mysql.DB.Model(&model.Task{}).Create(&task).Error
	if err != nil {
		return err
	}
	h.task = &task
	return nil
}

func (h *TaskHandler) createTask(form *request.Create) error {
	now := time.Now()
	taskRecord := model.Task{
		TaskGroupId: form.GroupId,
		Type:        form.TaskType(),
		Prompt:      form.Prompt,
		Model:       form.Model,
		Size:        form.Size,
		Status:      model.TaskStatusPending.String(),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := mysql.DB.Model(&model.Task{}).Create(&taskRecord).Error
	if err != nil {
		return err
	}
	for _, ii := range form.ImageIds {
		taskImageR := model.TaskImage{
			ImageId: ii,
			TaskId:  taskRecord.Id,
			Type:    model.TaskImageTypeInput.String(),
		}
		if form.ImageType != "" {
			taskImageR.Origin = sql.NullString{Valid: true, String: form.ImageType}
		} else {
			taskImageR.Origin = sql.NullString{Valid: false}
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageR).Error
		if err != nil {
			return err
		}
	}
	var task model.Task
	err = mysql.DB.Model(&model.Task{}).
		Preload("TaskImages").
		Preload("TaskImages.InputImage").
		Preload("TaskImages.OutputImage").
		Where("id = ?", taskRecord.Id).First(&task).Error
	if err != nil {
		return err
	}
	h.task = &task
	return nil
}

func (h *TaskHandler) createTaskRecord(form any) error {
	if _, ok := form.(request.TaskForm); ok {
		return h.createEditTask(form.(request.TaskForm))
	} else if _, ok := form.(*request.Generate); ok {
		return h.createGenerateTask(form.(*request.Generate))
	} else if _, ok := form.(*request.Create); ok {
		return h.createTask(form.(*request.Create))
	}
	return fmt.Errorf("unknown task form type: %T", form)
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
	type imageData struct {
		URL  string `json:"url"`
		Byte []byte `json:"byte"`
	}
	result := make([]imageData, 0)
	for _, v := range imageResp.GetURLs() {
		b, _, err := tools.GetOnlineImage(v)
		if err != nil {
			return err
		}
		result = append(result, imageData{URL: v, Byte: b})
	}
	for _, v := range imageResp.GetB64s() {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		result = append(result, imageData{URL: "", Byte: decoded})
	}

	for _, v := range result {
		path, err := saveNormalImage(v.Byte, h.task.CreatedAt, imageResp.GetSupplier())
		if err != nil {
			return err
		}
		thumbnailPath, err := saveThumbnailImage(v.Byte, h.task.CreatedAt, imageResp.GetSupplier())
		if err != nil {
			logs.Logger.Err(err).Msg("save thumbnail image error")
		}
		imageRecord := model.OutputImage{
			Path:              path,
			ThumbNailPath:     thumbnailPath,
			TTL:               0,
			Type:              string(model.OuputImageTypeNormal),
			ModelSupplierURL:  v.URL,
			ModelSupplierName: imageResp.GetSupplier(),
			ModelName:         imageResp.GetModel(),
		}
		if config.GConfig.CloudStorageEnabled {
			normal, err := uploadNormalImage(v.Byte)
			if err != nil {
				return err
			}
			imageRecord.StorageSupplierName = config.GConfig.CloudStorageSupplier
			imageRecord.Key = normal.Key
			imageRecord.ACL = "private"
			imageRecord.URL = normal.URL
		}
		err = mysql.DB.Model(&model.OutputImage{}).Create(&imageRecord).Error
		if err != nil {
			return err
		}
		taskImageRecord := model.TaskImage{
			TaskId:  h.task.Id,
			ImageId: imageRecord.Id,
			Type:    model.TaskImageTypeOutput.String(),
			Origin:  sql.NullString{Valid: false},
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) createCompressionRecords(imageResp image.Response) error {
	type imageData struct {
		URL  string `json:"url"`
		Byte []byte `json:"byte"`
	}
	result := make([]imageData, 0)
	for _, v := range imageResp.GetURLs() {
		b, _, err := tools.GetOnlineImage(v)
		if err != nil {
			return err
		}
		result = append(result, imageData{URL: v, Byte: b})
	}
	for _, v := range imageResp.GetB64s() {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return err
		}
		result = append(result, imageData{URL: "", Byte: decoded})
	}

	for _, v := range result {
		path, ratio, err := saveCompressionImage(v.Byte, 95, h.task.CreatedAt, imageResp.GetSupplier())
		if err != nil {
			return err
		}
		thumbnailPath, err := saveCompressionThumbnailImage(v.Byte, 95, h.task.CreatedAt, imageResp.GetSupplier())
		if err != nil {
			logs.Logger.Err(err).Msg("save thumbnail image error")
		}
		imageRecord := model.OutputImage{
			Path:              path,
			ThumbNailPath:     thumbnailPath,
			TTL:               0,
			Type:              string(model.OuputImageTypeCompressed),
			CompressionRatio:  sql.NullFloat64{Valid: true, Float64: ratio},
			ModelSupplierURL:  v.URL,
			ModelSupplierName: imageResp.GetSupplier(),
			ModelName:         imageResp.GetModel(),
		}
		if config.GConfig.CloudStorageEnabled {
			compression, _, err := uploadCompressionImage(v.Byte, 95)
			if err != nil {
				return err
			}
			imageRecord.StorageSupplierName = config.GConfig.CloudStorageSupplier
			imageRecord.Key = compression.Key
			imageRecord.ACL = "private"
			imageRecord.URL = compression.URL
		}
		err = mysql.DB.Model(&model.OutputImage{}).Create(&imageRecord).Error
		if err != nil {
			return err
		}
		taskImageRecord := model.TaskImage{
			TaskId:  h.task.Id,
			ImageId: imageRecord.Id,
			Type:    model.TaskImageTypeOutput.String(),
			Origin:  sql.NullString{Valid: false},
		}
		err = mysql.DB.Model(&model.TaskImage{}).Create(&taskImageRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) recordSupplierInvoke() error {
	for _, v := range h.imageResponse {
		exeRecord := model.SupplierInvokeHistory{
			TaskId:       h.task.Id,
			SupplierName: v.GetSupplier(),
			TokenDesc:    v.GetTokenDesc(),
			ModelName:    v.GetModel(),
			StatusCode:   v.GetStatusCode(),
			DurationMs:   v.TaskConsumeMs(),
			CreatedAt:    v.GetRespAt(),
		}
		respBody := v.GetRespBody()
		if v.GetError() != nil {
			if len(respBody) < 10000 {
				exeRecord.FailedRespBody = respBody
			}
			exeRecord.Error = v.GetError().Error()
		}
		err := mysql.DB.Model(&model.SupplierInvokeHistory{}).Create(&exeRecord).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *TaskHandler) Update(event int, data interface{}) {
	defer h.wg.Done()
	if event == consts.EventTaskEnd {
		h.imageResponse = data.([]image.Response)
		err := h.endWork()
		if err != nil {
			h.fail(err)
		}
	} else if event == consts.EventSysExit {
		data := data.(image.SysExitResponse)
		err := mysql.DB.Model(&model.Task{}).Where("id = ?", data.GetTaskID()).Update("status", model.TaskStatusAborted).Error
		if err != nil {
			logs.Logger.Error().Err(err).Msg("Update task status error")
		}
	}
}

func (h *TaskHandler) endWork() error {
	err := h.recordSupplierInvoke()
	if err != nil {
		return err
	}

	var succeed bool
	errs := make([]error, 0)
	for _, v := range h.imageResponse {
		if v.Succeed() {
			succeed = true
			err := h.createImageRecords(v)
			if err != nil {
				return err
			}
			taskRecord := model.Task{
				Id:       h.task.Id,
				Status:   model.TaskStatusSucceed.String(),
				Progress: 100,
			}
			if h.task.Model == "" {
				taskRecord.Model = v.GetModel()
			}
			err = mysql.DB.Updates(&taskRecord).Error
			if err != nil {
				return err
			}

			// 记录任务成功完成日志
			logs.Logger.Info().
				Int("task_id", h.task.Id).
				Str("supplier", v.GetSupplier()).
				Str("model", v.GetModel()).
				Str("status", "success").
				Int64("task_consume_ms", v.TaskConsumeMs()).
				Msg("Task completed successfully")
		} else {
			err := v.GetError()
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	if !succeed {
		var failReason string
		for _, err := range errs {
			if errors.Is(err, image.PromptError) {
				failReason = "该任务的输入可能违反了相关服务政策，请调整后进行重试"
				break
			}
		}
		taskRecord := model.Task{
			Id:           h.task.Id,
			Status:       model.TaskStatusFailed.String(),
			FailedReason: failReason,
		}
		err := mysql.DB.Updates(&taskRecord).Error
		if err != nil {
			return err
		}

		// 记录任务失败日志
		logs.Logger.Error().
			Int("task_id", h.task.Id).
			Str("model", h.task.Model).
			Str("status", "failed").
			Str("failed_reason", failReason).
			Errs("errors", errs).
			Msg("Task execution failed")
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
	for i := range tasks {
		tasks[i].TidyImage()
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
	if c.FullPath() == "/v2/task/slow/4oVip-four" {
		form.Prompt = form.Prompt + consts.FourImagePrompt
	}
	h, err := newTaskHandler(c)
	if err != nil {
		logs.Logger.Err(err).Msg("task-SlowSpeed-NewTaskHandler")
		c.JSON(http.StatusInternalServerError, response.ParamError)
	}
	err = h.createTaskRecord(&form)
	if err != nil {
		logs.Logger.Err(err).Msg("task-SlowSpeed")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	h.enqueue()
	c.JSON(http.StatusOK, response.SuccessWithData(h.task.TidyImageTask()))
}

func FastSpeed(c *gin.Context) {
	form := request.FastSpeed{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h, err := newTaskHandler(c)
	if err != nil {
		logs.Logger.Err(err).Msg("task-FastSpeed-NewTaskHandler")
		c.JSON(http.StatusInternalServerError, response.ParamError)
		return
	}
	err = h.createTaskRecord(&form)
	if err != nil {
		logs.Logger.Err(err).Msg("task-FastSpeed")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	h.enqueue()
	c.JSON(http.StatusOK, response.SuccessWithData(h.task.TidyImageTask()))
}

func Generate(c *gin.Context) {
	form := request.Generate{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	if c.FullPath() == "/v2/task/generate/4oVip-four" {
		form.Prompt = form.Prompt + consts.FourImagePrompt
	}
	h, err := newTaskHandler(c)
	if err != nil {
		logs.Logger.Err(err).Msg("task-Generate-NewTaskHandler")
		c.JSON(http.StatusInternalServerError, response.ParamError)
		return
	}
	err = h.createTaskRecord(&form)
	if err != nil {
		logs.Logger.Err(err).Msg("task-Generate")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	h.enqueue()
	c.JSON(http.StatusOK, response.SuccessWithData(h.task.TidyImageTask()))
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

func Create(c *gin.Context) {
	form := request.Create{}
	err := c.ShouldBind(&form)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}
	h, err := newTaskHandler(c)
	if err != nil {
		logs.Logger.Err(err).Msg("task-Generate-NewTaskHandler")
		c.JSON(http.StatusInternalServerError, response.ParamError)
		return
	}
	err = h.createTaskRecord(&form)
	if err != nil {
		logs.Logger.Err(err).Msg("task-Generate")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	h.enqueue()
	c.JSON(http.StatusOK, response.SuccessWithData(h.task.TidyImageTask()))
}

func saveNormalImage(image []byte, t time.Time, supplier string) (relativePath string, err error) {
	relativePath = filepath.Join("output", "o", t.Format("20060102"), supplier, uuid.New().String()+"."+tools.DetectImageType(image).String())
	path := filepath.Join(config.GConfig.LocalStorageDirectory, relativePath)
	err = local.SaveFile(bytes.NewReader(image), path)
	return
}

func saveCompressionImage(image []byte, quality int, t time.Time, supplier string) (relativePath string, ratio float64, err error) {
	compressionBytes, err := tools.ConvertAndCompressToJPEG(image, quality)
	if err != nil {
		return
	}
	ratio = float64(len(compressionBytes)) / float64(len(image))
	relativePath = filepath.Join("output", "c", t.Format("20060102"), supplier, uuid.New().String()+"."+tools.DetectImageType(compressionBytes).String())
	path := filepath.Join(config.GConfig.LocalStorageDirectory, relativePath)
	err = local.SaveFile(bytes.NewReader(compressionBytes), path)
	return
}

func saveThumbnailImage(image []byte, t time.Time, supplier string) (relativePath string, err error) {
	format, err := tools.DetectImageType(image).ImagingFormat()
	if err != nil {
		return "", err
	}
	thumbnail, err := tools.Thumbnail(bytes.NewReader(image), 0.5, format)
	relativePath = filepath.Join("output", "ot", t.Format("20060102"), supplier, uuid.New().String()+"."+strings.ToLower(format.String()))
	path := filepath.Join(config.GConfig.LocalStorageDirectory, relativePath)
	err = local.SaveFile(thumbnail, path)
	return
}

func saveCompressionThumbnailImage(image []byte, quality int, t time.Time, supplier string) (relativePath string, err error) {
	compressionBytes, err := tools.ConvertAndCompressToJPEG(image, quality)
	if err != nil {
		return
	}
	format, err := tools.DetectImageType(compressionBytes).ImagingFormat()
	if err != nil {
		return "", err
	}
	thumbnail, err := tools.Thumbnail(bytes.NewReader(compressionBytes), 0.5, format)
	relativePath = filepath.Join("output", "ct", t.Format("20060102"), supplier, uuid.New().String()+"."+strings.ToLower(format.String()))
	path := filepath.Join(config.GConfig.LocalStorageDirectory, relativePath)
	err = local.SaveFile(thumbnail, path)
	return
}

func uploadNormalImage(image []byte) (normal ali.OSSObject, err error) {
	key, err := ali.OssClient.UploadPrivateImage(image)
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

func uploadCompressionImage(image []byte, quality int) (compression ali.OSSObject, ratio float64, err error) {
	compressionBytes, err := tools.ConvertAndCompressToJPEG(image, quality)
	if err != nil {
		return
	}
	ratio = float64(len(compressionBytes)) / float64(len(image))
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
