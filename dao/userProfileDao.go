package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
)

// GetFavoriteDocumentsByUserID 获取用户收藏的文档列表(该函数在favoriteDao.go已经有一个定义了)
//func GetFavoriteDocumentsByUserID(userID uint64) ([]models.Document, error) {
//	db := config.GetDB()
//	var documents []models.Document
//	// 使用Join查询，通过favorites表筛选出documents
//	err := db.Joins("JOIN favorites ON favorites.document_id = documents.id").
//		Where("favorites.user_id = ?", userID).
//		Find(&documents).Error
//	return documents, err
//}

// GetViewHistoryDocumentsByUserID 获取用户浏览历史的文档列表
func GetViewHistoryDocumentsByUserID(userID uint64) ([]models.Document, error) {
	db := config.GetDB()
	var documents []models.Document
	// 使用Join查询，通过view_histories表筛选出documents
	err := db.Joins("JOIN view_histories ON view_histories.document_id = documents.id").
		Where("view_histories.user_id = ?", userID).
		Order("view_histories.created_at DESC"). // 通常按最近浏览排序
		Find(&documents).Error
	return documents, err
}
