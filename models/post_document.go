package models

import (
	"time"

	"gorm.io/gorm"
)

type PostDocument struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	PostID     uint64         `gorm:"not null;uniqueIndex:uk_post_document;index" json:"post_id"`
	DocumentID uint64         `gorm:"not null;uniqueIndex:uk_post_document;index:idx_document_id" json:"document_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
