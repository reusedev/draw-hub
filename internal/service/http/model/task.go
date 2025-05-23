package model

import "time"

type Task struct {
	Id           int       `json:"id" gorm:"primaryKey"`
	TaskGroupId  string    `json:"task_group_id" gorm:"column:task_group_id;type:varchar(50)"`
	Type         string    `json:"type" gorm:"column:type;type:enum('generate', 'edit')"`
	Prompt       string    `json:"prompt" gorm:"column:prompt;type:varchar(5000)"`
	Model        string    `json:"ai_model" gorm:"column:ai_model;type:varchar(20)"`
	Quality      string    `json:"quality" gorm:"column:quality;type:varchar(20)"`
	Size         string    `json:"size" gorm:"column:size;type:varchar(20)"`
	Status       string    `json:"status" gorm:"column:status;type:enum('queued', 'running', 'succeed', 'failed')"`
	FailedReason string    `json:"failed_reason" gorm:"column:failed_reason;type:varchar(1000)"`
	Progress     float32   `json:"progress" gorm:"column:progress;type:float"`
	CreatedAt    time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}
type TaskType string

const (
	TaskTypeGenerate TaskType = "generate"
	TaskTypeEdit     TaskType = "edit"
)

func (t TaskType) String() string {
	return string(t)
}

type TaskStatus string

const (
	TaskStatusQueued  TaskStatus = "queued"
	TaskStatusRunning TaskStatus = "running"
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
	ModelName      string    `json:"model_name" gorm:"column:model_name;type:varchar(20)"`
	StatusCode     int       `json:"status_code" gorm:"column:status_code;type:int"`
	FailedRespBody string    `json:"failed_resp_body" gorm:"column:failed_resp_body;type:varchar(2000)"`
	DurationMs     int64     `json:"duration_ms" gorm:"column:duration_ms;type:int"`
	CreatedAt      time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

type TaskImage struct {
	TaskId  int    `json:"task_id" gorm:"column:task_id;type:int;primaryKey"`
	ImageId int    `json:"image_id" gorm:"column:image_id;type:int;primaryKey"`
	Type    string `json:"type" gorm:"column:type;type:enum('input', 'output')"`
}

type TaskImageType string

const (
	TaskImageTypeInput  TaskImageType = "input"
	TaskImageTypeOutput TaskImageType = "output"
)

func (t TaskImageType) String() string {
	return string(t)
}
