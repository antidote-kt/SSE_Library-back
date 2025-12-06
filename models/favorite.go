package models

import (
	"time"

	"gorm.io/gorm"
)

type Favorite struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64         `gorm:"not null;uniqueIndex:idx_favorites_user_doc_unique" json:"user_id"`
	SourceID   uint64         `gorm:"not null;uniqueIndex:idx_favorites_user_doc_unique" json:"source_id"`
	SourceType string         `gorm:"type:varchar(20);not null;uniqueIndex:idx_favorites_user_doc_unique" json:"source_type"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
