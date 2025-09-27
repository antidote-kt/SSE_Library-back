package models

import (
	"time"

	"gorm.io/gorm"
)

type Resource struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	FileName  string         `json:"file_name" gorm:"size:255;not null"`
	FileType  string         `json:"file_type" gorm:"size:50;not null"`
	FileURL   string         `json:"file_url" gorm:"size:512;not null"` // 相对url
	FileSize  int64          `json:"file_size" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (Resource) TableName() string {
	return "resources"
}
