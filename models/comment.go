package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64         `gorm:"not null;index:idx_user_id" json:"user_id"`
	Content    string         `gorm:"type:text;not null" json:"content"`
	ParentID   *uint64        `gorm:"index:idx_parent_id" json:"parent_id"`
	SourceID   uint64         `gorm:"not null;index:idx_source" json:"source_id"`
	SourceType string         `gorm:"type:varchar(20);not null;default:'document';index:idx_source" json:"source_type"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	User       *User          `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Document   *Document      `gorm:"foreignKey:SourceID;references:ID" json:"-"` // 当 SourceType 为 document 时使用
	Post       *Post          `gorm:"foreignKey:SourceID;references:ID" json:"-"` // 当 SourceType 为 post 时使用
}
