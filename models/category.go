package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(50);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	ParentID    *uint64        `gorm:"index:idx_parent_id" json:"parent_id"`
	Documents   []Document     `gorm:"foreignKey:CategoryID"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Category) TableName() string {
	return "categories"
}
