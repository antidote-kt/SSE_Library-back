package models

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	User1ID   uint64         `gorm:"not null" json:"user1Id"`
	User2ID   uint64         `gorm:"not null" json:"user2Id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
