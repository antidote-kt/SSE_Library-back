package models

import (
	"time"

	"gorm.io/gorm"
)

type PostLike struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64         `gorm:"not null;uniqueIndex:uk_user_post_like;index" json:"user_id"`
	PostID    uint64         `gorm:"not null;uniqueIndex:uk_user_post_like;index:idx_post_id" json:"post_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
