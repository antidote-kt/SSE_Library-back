package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	SessionID uint64         `gorm:"not null;index:idx_session_id" json:"session_id"`
	SenderID  uint64         `gorm:"not null;index:idx_sender_id" json:"sender_id"`
	Content   string         `gorm:"not null" json:"content"`
	Status    string         `gorm:"not null" json:"status"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
