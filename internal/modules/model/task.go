package model

import (
	"database/sql"
	"time"

	"github.com/jinzhu/copier"
)

type Task struct {
	Id           int            `json:"id" gorm:"primaryKey"`
	TaskGroupId  string         `json:"task_group_id" gorm:"column:task_group_id;type:varchar(50)"`
	Type         string         `json:"type" gorm:"column:type;type:enum('generate', 'edit')"`
	Prompt       string         `json:"prompt" gorm:"column:prompt;type:varchar(5000)"`
	Speed        sql.NullString `json:"speed" gorm:"column:speed;type:enum('fast', 'slow')"`
	Model        string         `json:"model" gorm:"column:model;type:varchar(30)"`
	Quality      string         `json:"quality" gorm:"column:quality;type:varchar(20)"`
	Size         string         `json:"size" gorm:"column:size;type:varchar(20)"`
	Status       string         `json:"status" gorm:"column:status;type:enum('pending', 'queued', 'running', 'succeed', 'aborted', 'failed')"`
	FailedReason string         `json:"failed_reason" gorm:"column:failed_reason;type:varchar(1000)"`
	Progress     float32        `json:"progress" gorm:"column:progress;type:float"`
	CreatedAt    time.Time      `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	TaskImages   []TaskImage    `json:"task_images" gorm:"foreignKey:TaskId"`
}

func (*Task) TableName() string {
	return "task"
}

func (t *Task) TidyImageTask() *Task {
	c := t.DeepCopy()
	c.TidyImage()
	return c
}

func (t *Task) DeepCopy() *Task {
	newT := Task{}
	copier.CopyWithOption(&newT, &t, copier.Option{
		DeepCopy: true,
	})
	return &newT
}

func (t *Task) TidyImage() {
	for i := range t.TaskImages {
		if t.TaskImages[i].Type == TaskImageTypeInput.String() {
			if t.TaskImages[i].Origin.String == TaskImageOriginOutput.String() {
				inputImage := t.TaskImages[i].OutputImage
				t.TaskImages[i].InputImage = InputImage{
					Id:                  inputImage.Id,
					Path:                inputImage.Path,
					StorageSupplierName: inputImage.StorageSupplierName,
					Key:                 inputImage.Key,
					ACL:                 inputImage.ACL,
					TTL:                 inputImage.TTL,
					URL:                 inputImage.URL,
					CreatedAt:           inputImage.CreatedAt,
				}
			}
			t.TaskImages[i].OutputImage = OutputImage{}
		} else if t.TaskImages[i].Type == TaskImageTypeOutput.String() {
			t.TaskImages[i].InputImage = InputImage{}
		}
	}
}

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusQueued  TaskStatus = "queued"
	TaskStatusRunning TaskStatus = "running"
	TaskStatusAborted TaskStatus = "aborted"
	TaskStatusSucceed TaskStatus = "succeed"
	TaskStatusFailed  TaskStatus = "failed"
)

func (t TaskStatus) String() string {
	return string(t)
}

type SupplierInvokeHistory struct {
	Id             int       `json:"id" gorm:"primaryKey"`
	TaskId         int       `json:"task_id" gorm:"column:task_id;type:int"`
	SupplierName   string    `json:"supplier_name" gorm:"column:supplier_name;type:varchar(20)"`
	TokenDesc      string    `json:"token_desc" gorm:"column:token_desc;type:varchar(20)"`
	ModelName      string    `json:"model_name" gorm:"column:model_name;type:varchar(30)"`
	StatusCode     int       `json:"status_code" gorm:"column:status_code;type:int"`
	FailedRespBody string    `json:"failed_resp_body" gorm:"column:failed_resp_body;type:varchar(2000)"`
	DurationMs     int64     `json:"duration_ms" gorm:"column:duration_ms;type:int"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (SupplierInvokeHistory) TableName() string {
	return "supplier_invoke_history"
}

type TaskImage struct {
	TaskId      int            `json:"task_id" gorm:"column:task_id;type:int;primaryKey"`
	ImageId     int            `json:"image_id" gorm:"column:image_id;type:int;primaryKey"`
	Type        string         `json:"type" gorm:"column:type;type:enum('input', 'output');primaryKey"` // 类型
	Origin      sql.NullString `json:"origin" gorm:"column:origin;type:enum('input', 'output')"`        // 来源
	InputImage  InputImage     `json:"input_image" gorm:"foreignKey:ImageId;references:Id"`
	OutputImage OutputImage    `json:"output_image" gorm:"foreignKey:ImageId;references:Id"`
}

func (TaskImage) TableName() string {
	return "task_image"
}

type TaskImageType string

const (
	TaskImageTypeInput  TaskImageType = "input"
	TaskImageTypeOutput TaskImageType = "output"
)

func (t TaskImageType) String() string {
	return string(t)
}

type TaskImageOrigin string

const (
	TaskImageOriginInput  TaskImageOrigin = "input"
	TaskImageOriginOutput TaskImageOrigin = "output"
)

func (t TaskImageOrigin) String() string {
	return string(t)
}
