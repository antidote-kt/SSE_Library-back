package models

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SenderID     uint64         `gorm:"not null;index:idx_user_id" json:"sender_id"`
	Title        string         `gorm:"type:varchar(200)" json:"title"`
	Content      string         `gorm:"type:text;not null" json:"content"`
	ReadCount    uint32         `gorm:"default:0" json:"read_count"`
	LikeCount    uint32         `gorm:"default:0" json:"like_count"`
	CollectCount uint32         `gorm:"default:0" json:"collect_count"`
	CommentCount uint32         `gorm:"default:0" json:"comment_count"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
