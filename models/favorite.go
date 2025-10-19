package models

import (
	"time"

	"gorm.io/gorm"
)

type Favorite struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64         `gorm:"not null;uniqueIndex:uk_user_favorite" json:"user_id"`
	DocumentID uint64         `gorm:"not null;uniqueIndex:uk_user_favorite;index:idx_document_id" json:"document_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
