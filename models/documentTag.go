package models

import (
	"time"

	"gorm.io/gorm"
)

type DocumentTag struct {
	DocumentID uint64         `gorm:"primaryKey" json:"document_id"`
	TagID      uint64         `gorm:"primaryKey" json:"tag_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DocumentTag) TableName() string {
	return "document_tag"
}
