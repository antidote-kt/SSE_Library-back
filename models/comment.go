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
	DocumentID uint64         `gorm:"not null;index:idx_comment" json:"document_id"`
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	User       *User          `gorm:"foreignKey:UserID;references:ID" json:"-"`
	Document   *Document      `gorm:"foreignKey:DocumentID;references:ID" json:"-"`
}
