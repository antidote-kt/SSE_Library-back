package models

import (
	"time"

	"gorm.io/gorm"
)

type Course struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string         `gorm:"type:varchar(200);not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Instructor  string         `gorm:"type:varchar(50);not null" json:"instructor"`
	CategoryID  uint64         `json:"category_id"`
	Duration    int            `json:"duration"`
	Thumbnail   string         `gorm:"type:varchar(500)" json:"thumbnail"`
	URL         string         `gorm:"type:varchar(500)" json:"url"`
	Documents   []Document     `gorm:"foreignKey:CourseID"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
