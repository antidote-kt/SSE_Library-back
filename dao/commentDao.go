package dao

import (
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

func CreateComment(comment *models.Comment) error {
	db := config.GetDB()
	return db.Create(comment).Error
}

func GetCommentsByDocumentID(documentID uint64) ([]models.Comment, error) {
	db := config.GetDB()
	var comments []models.Comment

	err := db.Where("document_id = ? AND deleted_at IS NULL", documentID).
		Order("created_at DESC").
		Find(&comments).Error

	return comments, err
}

type CommentWithDetails struct {
	CommentID          int64     `gorm:"column:comment_id"`
	ParentID           *uint64   `gorm:"column:parent_id"`
	CommentContent     string    `gorm:"column:comment_content"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UserID             int64     `gorm:"column:user_id"`
	Username           string    `gorm:"column:username"`
	UserAvatar         string    `gorm:"column:user_avatar"`
	Status             string    `gorm:"column:status"`
	UserCreateTime     time.Time `gorm:"column:user_create_time"`
	Email              string    `gorm:"column:email"`
	Role               string    `gorm:"column:role"`
	DocumentName       string    `gorm:"column:document_name"`
	DocumentID         int64     `gorm:"column:document_id"`
	Type               string    `gorm:"column:type"`
	DocumentUploadTime time.Time `gorm:"column:document_upload_time"`
	DocumentStatus     string    `gorm:"column:document_status"`
	Category           string    `gorm:"column:category"`
	Course             string    `gorm:"column:course"`
	Collections        int64     `gorm:"column:collections"`
	ReadCounts         int64     `gorm:"column:read_counts"`
	URL                string    `gorm:"column:url"`
	DocumentContent    string    `gorm:"column:document_content"`
	DocumentCreateTime time.Time `gorm:"column:document_create_time"`
}

func GetCommentWithUserAndDocument(documentID uint64) ([]CommentWithDetails, error) {
	db := config.GetDB()
	var results []CommentWithDetails

	err := db.Table("comments c").
		Select(`
			c.id as comment_id,
			c.parent_id as parent_id,
			c.content as comment_content,
			c.created_at,
			u.id as user_id,
			u.username,
			u.avatar as user_avatar,
			u.status,
			u.created_at as user_create_time,
			u.email,
			u.role,
			d.name as document_name,
			d.id as document_id,
			d.type,
			d.created_at as document_upload_time,
			d.status as document_status,
			cat.name as category,
			cat.name as course,
			d.collections,
			d.read_counts,
			d.url,
			d.introduction as document_content,
			d.created_at as document_create_time
		`).
		Joins("LEFT JOIN users u ON c.user_id = u.id").
		Joins("LEFT JOIN documents d ON c.document_id = d.id").
		Joins("LEFT JOIN categories cat ON d.category_id = cat.id").
		Where("c.document_id = ? AND c.deleted_at IS NULL", documentID).
		Order("c.id DESC").
		Scan(&results).Error

	return results, err
}

func GetAllCommentsWithUserAndDocument() ([]CommentWithDetails, error) {
	db := config.GetDB()
	var results []CommentWithDetails

	err := db.Table("comments c").
		Select(`
			c.id as comment_id,
			c.parent_id as parent_id,
			c.content as comment_content,
			c.created_at,
			u.id as user_id,
			u.username,
			u.avatar as user_avatar,
			u.status,
			u.created_at as user_create_time,
			u.email,
			u.role,
			d.name as document_name,
			d.id as document_id,
			d.type,
			d.created_at as document_upload_time,
			d.status as document_status,
			cat.name as category,
			cat.name as course,
			d.collections,
			d.read_counts,
			d.url,
			d.introduction as document_content,
			d.created_at as document_create_time
		`).
		Joins("LEFT JOIN users u ON c.user_id = u.id").
		Joins("LEFT JOIN documents d ON c.document_id = d.id").
		Joins("LEFT JOIN categories cat ON d.category_id = cat.id").
		Where("c.deleted_at IS NULL").
		Order("c.id DESC").
		Scan(&results).Error

	return results, err
}

func GetCommentByID(commentID uint64) (*models.Comment, error) {
	db := config.GetDB()
	var comment models.Comment

	err := db.Where("id = ? AND deleted_at IS NULL", commentID).
		First(&comment).Error

	if err != nil {
		return nil, err
	}

	return &comment, nil
}

func DeleteComment(commentID uint64) error {
	db := config.GetDB()
	return db.Where("id = ?", commentID).Delete(&models.Comment{}).Error
}

func UpdateComment(comment *models.Comment) error {
	db := config.GetDB()
	return db.Save(comment).Error
}

func GetUserCommentsWithUserAndDocument(userID uint64) ([]CommentWithDetails, error) {
	db := config.GetDB()
	var results []CommentWithDetails

	err := db.Table("comments c").
		Select(`
			c.id as comment_id,
			c.parent_id as parent_id,
			c.content as comment_content,
			c.created_at,
			u.id as user_id,
			u.username,
			u.avatar as user_avatar,
			u.status,
			u.created_at as user_create_time,
			u.email,
			u.role,
			d.name as document_name,
			d.id as document_id,
			d.type,
			d.created_at as document_upload_time,
			d.status as document_status,
			cat.name as category,
			cat.name as course,
			d.collections,
			d.read_counts,
			d.url,
			d.introduction as document_content,
			d.created_at as document_create_time
		`).
		Joins("LEFT JOIN users u ON c.user_id = u.id").
		Joins("LEFT JOIN documents d ON c.document_id = d.id").
		Joins("LEFT JOIN categories cat ON d.category_id = cat.id").
		Where("c.user_id = ? AND c.deleted_at IS NULL", userID).
		Order("c.id DESC").
		Scan(&results).Error

	return results, err
}

func GetCommentDetailByID(commentID uint64) (*CommentWithDetails, error) {
	db := config.GetDB()
	var result CommentWithDetails

	err := db.Table("comments c").
		Select(`
			c.id as comment_id,
			c.parent_id as parent_id,
			c.content as comment_content,
			c.created_at,
			u.id as user_id,
			u.username,
			u.avatar as user_avatar,
			u.status,
			u.created_at as user_create_time,
			u.email,
			u.role,
			d.name as document_name,
			d.id as document_id,
			d.type,
			d.created_at as document_upload_time,
			d.status as document_status,
			cat.name as category,
			cat.name as course,
			d.collections,
			d.read_counts,
			d.url,
			d.introduction as document_content,
			d.created_at as document_create_time
		`).
		Joins("LEFT JOIN users u ON c.user_id = u.id").
		Joins("LEFT JOIN documents d ON c.document_id = d.id").
		Joins("LEFT JOIN categories cat ON d.category_id = cat.id").
		Where("c.id = ? AND c.deleted_at IS NULL", commentID).
		Scan(&result).Error

	if err != nil {
		return nil, err
	}

	if result.CommentID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return &result, nil
}

func DeleteUserComment(userID, commentID uint64) error {
	db := config.GetDB()

	var comment models.Comment
	err := db.Where("id = ? AND user_id = ? AND deleted_at IS NULL", commentID, userID).
		First(&comment).Error

	if err != nil {
		return err
	}

	return db.Where("id = ? AND user_id = ?", commentID, userID).
		Delete(&models.Comment{}).Error
}
