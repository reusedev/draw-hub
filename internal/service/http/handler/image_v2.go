package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/reusedev/draw-hub/config"
	"github.com/reusedev/draw-hub/internal/components/mysql"
	"github.com/reusedev/draw-hub/internal/consts"
	mcache "github.com/reusedev/draw-hub/internal/modules/cache"
	"github.com/reusedev/draw-hub/internal/modules/dao"
	"github.com/reusedev/draw-hub/internal/modules/logs"
	"github.com/reusedev/draw-hub/internal/modules/model"
	"github.com/reusedev/draw-hub/internal/modules/storage/ali"
	"github.com/reusedev/draw-hub/internal/service/http/handler/response"
	"github.com/reusedev/draw-hub/tools"
)

// UploadImageV2New 图片上传 POST /api/v2/image/upload
// 接收 multipart file，如果是 PNG 则压缩为 JPG，两份都上传到 OSS，存入 image 表
func UploadImageV2New(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ParamErrorWithMessage("file is required"))
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-ReadFile")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	if !config.GConfig.CloudStorageEnabled {
		c.JSON(http.StatusInternalServerError, response.ParamErrorWithMessage("cloud storage is not enabled"))
		return
	}

	imageRecord := model.Image{
		Bucket:    ali.OssSgClient.GetBucket(),
		CreatedAt: time.Now(),
	}

	imageType := tools.DetectImageType(fileBytes)
	ext := ".tmp"
	if imageType != tools.ImageTypeUnknown {
		ext = "." + imageType.String()
	}

	// 0. 直接保存原图到 raw_key
	rawKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ext
	err = uploadBytesSToSgOSS(rawKey, header.Filename, fileBytes)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-Raw-OSS")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	imageRecord.RawKey = rawKey

	// 1. 获取 PNG 原图（如果不是 PNG，转换为 PNG，确保存到 png_key 的一定是 PNG）
	pngBytes, err := tools.ConvertToPNG(fileBytes)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-ConvertPNG")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	pngKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ".png"
	err = uploadBytesSToSgOSS(pngKey, header.Filename, pngBytes)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-PNG-OSS")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	imageRecord.PNGKey = pngKey

	// 2. 获取 JPG 压缩图
	jpgBytes, err := tools.ConvertAndCompressToJPEG(fileBytes, 95)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-CompressJPG")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	jpgKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ".jpeg"
	err = uploadBytesSToSgOSS(jpgKey, uuid.New().String()+".jpeg", jpgBytes)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-JPG-OSS")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}
	imageRecord.JPGKey = jpgKey

	err = mysql.DB.Create(&imageRecord).Error
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Upload-DB")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithData(gin.H{
		"id": imageRecord.Id,
	}))
}

// GetImageV2 获取图片链接 POST /api/v2/image/get
// 根据 image 表 ID 生成预签名 URL，24 小时内缓存相同结果，URL 有效期 7 天
func GetImageV2(c *gin.Context) {
	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}

	cacheKey := fmt.Sprintf("image_v2_urls_%d", req.ID)

	// 尝试从缓存获取
	cached, _ := mcache.ImageV2CacheManager().GetValue(cacheKey)
	if cached != "" {
		var data gin.H
		if err := json.Unmarshal([]byte(cached), &data); err == nil {
			c.JSON(http.StatusOK, response.SuccessWithData(data))
			return
		}
	}

	// 缓存未命中，查库并生成 URL
	img, err := dao.ImageById(req.ID)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Image-Get-DB")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	duration, _ := time.ParseDuration(config.GConfig.URLExpires)

	data := gin.H{
		"png_url":       "",
		"jpg_url":       "",
		"raw_url":       "",
		"thumbnail_url": "",
	}

	// Raw URL
	if img.RawKey != "" {
		rawURL, err := ali.OssSgClient.URL(img.RawKey, duration)
		if err != nil {
			logs.Logger.Err(err).Msg("V2-Image-Get-Raw-URL")
		} else {
			data["raw_url"] = rawURL
		}
	}

	// PNG URL
	if img.PNGKey != "" {
		pngURL, err := ali.OssSgClient.URL(img.PNGKey, duration)
		if err != nil {
			logs.Logger.Err(err).Msg("V2-Image-Get-PNG-URL")
		} else {
			data["png_url"] = pngURL
		}
	}

	// JPG URL
	if img.JPGKey != "" {
		jpgURL, err := ali.OssSgClient.URL(img.JPGKey, duration)
		if err != nil {
			logs.Logger.Err(err).Msg("V2-Image-Get-JPG-URL")
		} else {
			data["jpg_url"] = jpgURL
		}
	}

	// Thumbnail URL（基于 raw key，使用 OSS 图片处理缩放 50%）
	thumbnailKey := img.RawKey
	if thumbnailKey != "" {
		presignResult, err := ali.OssSgClient.Resize50(thumbnailKey, duration)
		if err != nil {
			logs.Logger.Err(err).Msg("V2-Image-Get-Thumbnail-URL")
		} else {
			data["thumbnail_url"] = presignResult.URL
		}
	}

	// 写入缓存，24 小时有效
	if jsonBytes, err := json.Marshal(data); err == nil {
		_ = mcache.ImageV2CacheManager().SetWithExpiration(cacheKey, string(jsonBytes), 24*time.Hour)
	}

	c.JSON(http.StatusOK, response.SuccessWithData(data))
}

// CreateTaskV2 创建生图任务 POST /api/v2/task/create
// 使用 image 表中的图片作为输入，创建编辑任务
func CreateTaskV2(c *gin.Context) {
	var req struct {
		ImageIds []int  `json:"image_ids"`
		Prompt   string `json:"prompt" binding:"required"`
		Model    string `json:"model" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}

	now := time.Now()
	taskType := consts.TaskTypeGenerate.String()

	// 如果提供了 image_id，需要获取输入图片
	type inputImageInfo struct {
		inputImageId int
	}
	var inputImages []inputImageInfo

	if len(req.ImageIds) > 0 {
		taskType = consts.TaskTypeEdit.String()
		for _, imageId := range req.ImageIds {
			img, err := dao.ImageById(imageId)
			if err != nil {
				logs.Logger.Err(err).Int("image_id", imageId).Msg("V2-Task-Create-GetImage")
				c.JSON(http.StatusBadRequest, response.ParamErrorWithMessage(fmt.Sprintf("image not found: %d", imageId)))
				return
			}

			// 获取图片 OSS key
			ossKey := img.RawKey
			if ossKey == "" {
				c.JSON(http.StatusBadRequest, response.ParamErrorWithMessage(fmt.Sprintf("image has no OSS key: %d", imageId)))
				return
			}

			// 创建 InputImage 记录，复用现有的 TaskImage 关联机制
			inputImage := model.InputImage{
				StorageSupplierName: config.GConfig.CloudStorageSupplier,
				Bucket:              img.Bucket,
				Key:                 ossKey,
				ACL:                 "private",
				CreatedAt:           now,
			}
			err = mysql.DB.Create(&inputImage).Error
			if err != nil {
				logs.Logger.Err(err).Msg("V2-Task-Create-InputImage")
				c.JSON(http.StatusInternalServerError, response.InternalError)
				return
			}
			inputImages = append(inputImages, inputImageInfo{inputImageId: inputImage.Id})
		}
	}

	// 创建 Task 记录
	taskRecord := model.Task{
		TaskGroupId: uuid.New().String(),
		Type:        taskType,
		Prompt:      req.Prompt,
		Model:       req.Model,
		Status:      model.TaskStatusPending.String(),
		ApiVersion:  "v2",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	err := mysql.DB.Create(&taskRecord).Error
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Task-Create-Task")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	// 关联输入图片
	for _, input := range inputImages {
		taskImageRecord := model.TaskImage{
			TaskId:  taskRecord.Id,
			ImageId: input.inputImageId,
			Type:    model.TaskImageTypeInput.String(),
			Origin:  sql.NullString{Valid: false},
		}
		err = mysql.DB.Create(&taskImageRecord).Error
		if err != nil {
			logs.Logger.Err(err).Msg("V2-Task-Create-TaskImage")
			c.JSON(http.StatusInternalServerError, response.InternalError)
			return
		}
	}

	// 加载完整 task（包含关联的 TaskImages）
	var task model.Task
	err = mysql.DB.Model(&model.Task{}).
		Preload("TaskImages").
		Preload("TaskImages.InputImage").
		Preload("TaskImages.OutputImage").
		Where("id = ?", taskRecord.Id).First(&task).Error
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Task-Create-LoadTask")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	// 入队执行
	h := &TaskHandler{task: &task}
	h.enqueue()

	c.JSON(http.StatusOK, response.SuccessWithData(gin.H{
		"id": task.Id,
	}))
}

// GetTaskV2 获取任务状态 POST /api/v2/task/get
func GetTaskV2(c *gin.Context) {
	var req struct {
		ID int `json:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.ParamError)
		return
	}

	var task model.Task
	err := mysql.DB.Model(&model.Task{}).Where("id = ?", req.ID).First(&task).Error
	if err != nil {
		logs.Logger.Err(err).Msg("V2-Task-Get-DB")
		c.JSON(http.StatusInternalServerError, response.InternalError)
		return
	}

	c.JSON(http.StatusOK, response.SuccessWithData(gin.H{
		"status":   task.Status,
		"image_id": task.GeneratedImageId,
	}))
}

// uploadBytesSToSgOSS 上传字节数据到 OSS
func uploadBytesSToSgOSS(key, filename string, data []byte) error {
	urlExpire, _ := time.ParseDuration(config.GConfig.URLExpires)
	req := ali.UploadRequest{
		Key:       key,
		Filename:  filename,
		File:      bytes.NewReader(data),
		Acl:       "private",
		URLExpire: urlExpire,
	}
	_, err := ali.OssSgClient.UploadFile(&req)
	return err
}

// saveImageRecord 任务完成后，将生成的图片保存到 image 表
// 上传原图（PNG）和压缩图（JPG）到 OSS
func saveImageRecord(imageBytes []byte, supplierURL, supplierName string) (int, error) {
	imageRecord := model.Image{
		Bucket:            ali.OssSgClient.GetBucket(),
		ModelSupplierURL:  supplierURL,
		ModelSupplierName: supplierName,
		CreatedAt:         time.Now(),
	}

	imageType := tools.DetectImageType(imageBytes)
	ext := ".tmp"
	if imageType != tools.ImageTypeUnknown {
		ext = "." + imageType.String()
	}

	// 0. 直接保存原图到 raw_key
	rawKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ext
	err := uploadBytesSToSgOSS(rawKey, uuid.New().String()+ext, imageBytes)
	if err != nil {
		return 0, err
	}
	imageRecord.RawKey = rawKey

	// 无论原图是什么格式，都转换为 PNG 存放到 png_key
	pngBytes, err := tools.ConvertToPNG(imageBytes)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-SaveImage-ConvertPNG")
		// 如果转换失败，退化为直接保存原始格式到 PNGKey，防止丢失原图
		pngBytes = imageBytes
	}
	pngKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ".png"
	err = uploadBytesSToSgOSS(pngKey, uuid.New().String()+".png", pngBytes)
	if err != nil {
		return 0, err
	}
	imageRecord.PNGKey = pngKey

	// 无论原图是什么格式，都压缩为 JPG 存放到 jpg_key
	jpgBytes, err := tools.ConvertAndCompressToJPEG(imageBytes, 95)
	if err != nil {
		logs.Logger.Err(err).Msg("V2-SaveImage-CompressJPG")
	} else {
		jpgKey := config.GConfig.AliOssSg.Directory + uuid.New().String() + ".jpeg"
		err = uploadBytesSToSgOSS(jpgKey, uuid.New().String()+".jpeg", jpgBytes)
		if err != nil {
			logs.Logger.Err(err).Msg("V2-SaveImage-JPG-OSS")
		} else {
			imageRecord.JPGKey = jpgKey
		}
	}

	err = mysql.DB.Create(&imageRecord).Error
	if err != nil {
		return 0, err
	}
	return imageRecord.Id, nil
}
