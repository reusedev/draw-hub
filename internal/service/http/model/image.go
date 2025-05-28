package model

import (
	"github.com/shopspring/decimal"
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
	Id                  int                 `json:"id" gorm:"primaryKey"`
	Path                string              `json:"path" gorm:"column:path;type:varchar(255)"`
	StorageSupplierName string              `json:"storage_supplier_name" gorm:"column:storage_supplier_name;type:varchar(20)"`
	Key                 string              `json:"key" gorm:"column:key;type:varchar(100)"`
	ACL                 string              `json:"acl" gorm:"column:acl;type:varchar(20)"`
	TTL                 int                 `json:"ttl" gorm:"column:ttl;type:int;default:0"` // Time to live in days
	URL                 string              `json:"url" gorm:"column:url;type:varchar(500)"`
	Type                string              `json:"type" gorm:"column:type;type:enum('normal', 'compressed')"`
	CompressionRatio    decimal.NullDecimal `json:"compression_ratio" gorm:"column:compression_ratio;type:decimal(5,2)"`
	OriginalURL         string              `json:"original_url" gorm:"column:original_url;type:varchar(500)"`
	ModelSupplierName   string              `json:"model_supplier_name" gorm:"column:model_supplier_name;type:varchar(20)"`
	ModelName           string              `json:"model_name" gorm:"column:model_name;type:varchar(20)"`
	CreatedAt           time.Time           `json:"created_at" gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP"`
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
