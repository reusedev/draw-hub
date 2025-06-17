package model

import (
	"database/sql"
	"time"
)

type InputImage struct {
	Id                  int       `json:"id" gorm:"primaryKey"`
	Path                string    `json:"path" gorm:"column:path;type:varchar(255)"`
	StorageSupplierName string    `json:"storage_supplier_name" gorm:"column:storage_supplier_name;type:varchar(20)"`
	Key                 string    `json:"key" gorm:"column:key;type:varchar(100)"`
	ACL                 string    `json:"acl" gorm:"column:acl;type:varchar(20)"`
	TTL                 int       `json:"ttl" gorm:"column:ttl;type:int;default:0"` // Time to live in days
	URL                 string    `json:"url" gorm:"column:url;type:varchar(500)"`
	CreatedAt           time.Time `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (InputImage) TableName() string {
	return "input_image"
}

type OutputImage struct {
	Id   int    `json:"id" gorm:"primaryKey"`
	Path string `json:"path" gorm:"column:path;type:varchar(255)"`
	// TODO save local thumbnail image
	ThumbNailPath       string          `json:"thumbnail_path" gorm:"column:thumbnail_path;type:varchar(255)"` // Path to the thumbnail image
	StorageSupplierName string          `json:"storage_supplier_name" gorm:"column:storage_supplier_name;type:varchar(20)"`
	Key                 string          `json:"key" gorm:"column:key;type:varchar(100)"`
	ACL                 string          `json:"acl" gorm:"column:acl;type:varchar(20)"`
	TTL                 int             `json:"ttl" gorm:"column:ttl;type:int;default:0"` // Time to live in days
	URL                 string          `json:"url" gorm:"column:url;type:varchar(500)"`  // oss URL in MySQL
	Type                string          `json:"type" gorm:"column:type;type:enum('normal', 'compressed')"`
	CompressionRatio    sql.NullFloat64 `json:"compression_ratio" gorm:"column:compression_ratio;type:float"`
	ModelSupplierURL    string          `json:"model_supplier_url" gorm:"column:model_supplier_url;type:varchar(500)"`
	ModelSupplierName   string          `json:"model_supplier_name" gorm:"column:model_supplier_name;type:varchar(20)"`
	ModelName           string          `json:"model_name" gorm:"column:model_name;type:varchar(20)"`
	CreatedAt           time.Time       `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
}

func (OutputImage) TableName() string {
	return "output_image"
}

type OutputImageType string

const (
	OuputImageTypeNormal     OutputImageType = "normal"
	OuputImageTypeCompressed OutputImageType = "compressed"
)

func (o OutputImageType) String() string {
	return string(o)
}
