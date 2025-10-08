package models

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	TagName   string         `gorm:"type:varchar(50);not null" json:"tag_name"`
	Documents []Document     `gorm:"many2many:document_tag"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
