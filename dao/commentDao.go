package dao

import (
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

func preloadCommentRelations(db *gorm.DB) *gorm.DB {
	return db.Preload("User").Preload("Document").Preload("Post")
}

func GetCommentWithUserAndDocument(sourceID uint64, sourceType string) ([]models.Comment, error) {
	db := config.GetDB()
	var comments []models.Comment

	err := preloadCommentRelations(db).
		Where("source_id = ? AND source_type = ? AND deleted_at IS NULL", sourceID, sourceType).
		Order("created_at DESC").
		Find(&comments).Error

	return comments, err
}

func GetAllCommentsWithUserAndDocument() ([]models.Comment, error) {
	db := config.GetDB()
	var comments []models.Comment

	err := preloadCommentRelations(db).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&comments).Error

	return comments, err
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

func GetUserCommentsWithUserAndDocument(userID uint64) ([]models.Comment, error) {
	db := config.GetDB()
	var comments []models.Comment

	err := preloadCommentRelations(db).
		Where("user_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&comments).Error

	return comments, err
}

func GetCommentDetailByID(commentID uint64) (*models.Comment, error) {
	db := config.GetDB()
	var comment models.Comment

	err := preloadCommentRelations(db).
		Where("id = ? AND deleted_at IS NULL", commentID).
		First(&comment).Error

	if err != nil {
		return nil, err
	}

	return &comment, nil
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
