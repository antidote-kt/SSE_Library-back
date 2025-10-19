package dao

import (
	"errors"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// CheckFavoriteExist 检查是否已收藏文档
func CheckFavoriteExist(userID, documentID uint64) (bool, error) {
	db := config.GetDB()
	var favorite models.Favorite

	err := db.Where("user_id = ? AND document_id = ?", userID, documentID).First(&favorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New("数据库查询失败")
	}

	return true, nil
}

// GetFavoriteDocumentsByUserID 获取用户收藏的文档列表
func GetFavoriteDocumentsByUserID(userID uint64) ([]models.Document, error) {
	db := config.GetDB()
	var favorites []models.Favorite

	// 获取用户的收藏记录
	err := db.Where("user_id = ?", userID).Find(&favorites).Error
	if err != nil {
		return nil, errors.New("查询收藏记录失败")
	}

	var documents []models.Document
	for _, favorite := range favorites {
		var document models.Document
		// 获取收藏的文档
		err := db.Where("id = ?", favorite.DocumentID).First(&document).Error
		if err != nil {
			continue // 跳过不存在的文档
		}
		documents = append(documents, document)
	}

	return documents, nil
}
