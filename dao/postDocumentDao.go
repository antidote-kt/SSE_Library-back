package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
)

// GetPostsByDocumentID 获取与指定文档关联的帖子列表
func GetPostsByDocumentID(documentID uint64) ([]models.Post, error) {
	db := config.GetDB()
	var posts []models.Post

	err := db.Table("posts").
		Joins("JOIN post_documents ON posts.id = post_documents.post_id").
		Where("post_documents.document_id = ?", documentID).
		Find(&posts).Error

	if err != nil {
		return nil, err
	}

	return posts, nil
}
