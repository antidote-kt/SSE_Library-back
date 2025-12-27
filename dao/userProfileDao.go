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
	// favoriteDao.go第44行的GetFavoriteDocumentsByUserID(userID)采用的是N+1循环查找模式，这里采用另一种方式————Join联表查询
	err := db.Joins("JOIN view_histories ON view_histories.source_id = documents.id").
		Where("view_histories.user_id = ?", userID).
		Where("view_histories.source_type = ?", "document").
		Order("view_histories.created_at DESC"). // 通常按最近浏览排序
		Find(&documents).Error
	return documents, err
}
