package dao

import (
	"errors"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// CheckFavoriteDocumentExist 检查是否已收藏文档
func CheckFavoriteDocumentExist(userID, sourceID uint64) (bool, error) {
	db := config.GetDB()
	var favorite models.Favorite

	err := db.Where("user_id = ? AND source_id = ? AND source_type = ?", userID, sourceID, constant.DocumentType).First(&favorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New(constant.DatabaseError)
	}

	return true, nil
}

// CheckFavoritePostExist 检查是否已收藏帖子
func CheckFavoritePostExist(userID, postID uint64) (bool, error) {
	db := config.GetDB()
	var favorite models.Favorite

	err := db.Where("user_id = ? AND source_id = ? AND source_type = ?", userID, postID, constant.PostType).First(&favorite).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, errors.New(constant.DatabaseError)
	}

	return true, nil
}

// GetFavoriteDocumentsByUserID 获取用户收藏的文档列表
func GetFavoriteDocumentsByUserID(userID uint64) ([]models.Document, error) {
	db := config.GetDB()
	var favorites []models.Favorite

	// 获取用户收藏的文档记录
	err := db.Where("user_id = ? AND source_type = ?", userID, constant.DocumentType).Find(&favorites).Error
	if err != nil {
		return nil, errors.New(constant.FavoriteGetFailed)
	}

	var documents []models.Document
	for _, favorite := range favorites {
		var document models.Document
		// 获取收藏的文档
		err := db.Where("id = ?", favorite.SourceID).First(&document).Error
		if err != nil {
			continue // 跳过不存在的文档
		}
		documents = append(documents, document)
	}

	return documents, nil
}

// GetFavoritePostsByUserID 获取用户收藏的帖子列表
func GetFavoritePostsByUserID(userID uint64) ([]models.Post, error) {
	db := config.GetDB()
	var favorites []models.Favorite

	// 获取用户收藏的帖子记录
	err := db.Where("user_id = ? AND source_type = ?", userID, constant.PostType).Find(&favorites).Error
	if err != nil {
		return nil, errors.New(constant.FavoriteGetFailed)
	}

	var posts []models.Post
	for _, favorite := range favorites {
		var post models.Post
		// 获取收藏的帖子
		err := db.Where("id = ?", favorite.SourceID).First(&post).Error
		if err != nil {
			continue // 跳过不存在的帖子
		}
		posts = append(posts, post)
	}

	return posts, nil
}
