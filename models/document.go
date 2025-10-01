package models

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	Type         string         `gorm:"type:varchar(20);not null" json:"type"`
	Name         string         `gorm:"type:varchar(200);not null" json:"name"`
	BookISBN     string         `gorm:"type:varchar(20)" json:"book_isbn"`
	Author       string         `gorm:"type:varchar(100);not null" json:"author"`
	UploaderID   uint64         `gorm:"not null;index:idx_uploader_id" json:"uploader_id"`
	CourseID     uint64         `gorm:"not null;index:idx_category_id" json:"course_id"`
	Cover        string         `gorm:"type:varchar(500)" json:"cover"`
	Introduction string         `gorm:"type:text" json:"introduction"`
	CreateYear   string         `gorm:"type:varchar(10)" json:"create_year"`
	Status       string         `gorm:"type:varchar(20);default:'audit'" json:"status"`
	ReadCounts   int            `gorm:"default:0" json:"read_counts"`
	Collections  int            `gorm:"default:0" json:"collections"`
	URL          string         `gorm:"type:varchar(500);not null" json:"url"`
	CreatedAt    time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
