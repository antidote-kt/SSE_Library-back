package models

import (
	"time"

	"gorm.io/gorm"
)

type AIMessage struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	AISessionsID uint64         `gorm:"column:ai_sessions_id;not null;index:idx_session_messages" json:"aiSessionsId"`
	Role         string         `gorm:"size:20;not null" json:"role"`
	Status       string         `gorm:"size:20;not null;default:generating" json:"status"`
	Content      string         `gorm:"type:longtext;not null" json:"content"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"createdAt"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
