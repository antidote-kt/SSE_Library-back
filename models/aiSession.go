package models

import (
	"time"

	"gorm.io/gorm"
)

type AISession struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64         `gorm:"not null;index:idx_user_sessions" json:"userId"`
	Title     string         `gorm:"size:255;default:'新对话'" json:"title"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime;index:idx_user_sessions" json:"updatedAt"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
