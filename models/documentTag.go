package models

import (
	"time"

	"gorm.io/gorm"
)

type DocumentTag struct {
	ID         uint64         `gorm:"primary_key;autoIncrement" json:"id"`
	DocumentID uint64         `json:"document_id"`
	TagID      uint64         `json:"tag_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DocumentTag) TableName() string {
	return "document_tag"
}
