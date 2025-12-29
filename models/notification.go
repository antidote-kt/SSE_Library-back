package models

import (
	"time"

	"gorm.io/gorm"
)

type Notification struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	ReceiverID uint64         `gorm:"not null;index:idx_receiver_status,idx_receiver_type" json:"receiver_id"`
	Type       string         `gorm:"not null;size:50;index:idx_receiver_type" json:"type"`
	Content    string         `gorm:"type:text" json:"content"`
	IsRead     bool           `gorm:"not null;default:0;index:idx_receiver_status" json:"is_read"`
	SourceID   uint64         `gorm:"not null" json:"source_id"`
	SourceType string         `gorm:"not null;size:50" json:"source_type"`
	CreatedAt  time.Time      `gorm:"autoCreateTime;index:idx_created_at" json:"created_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
