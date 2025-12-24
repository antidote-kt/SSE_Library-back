package dao

import (
	"errors"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// CreatePostWithTx 使用事务创建帖子及其关联文档
func CreatePostWithTx(post *models.Post, documentIDs []uint64) error {
	db := config.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建帖子主体
		if err := tx.Create(post).Error; err != nil {
			return errors.New(constant.CreatePostFailed)
		}

		// 2. 如果有关联文档，创建关联关系
		if len(documentIDs) > 0 {
			var postDocuments []models.PostDocument
			for _, docID := range documentIDs {
				// 检查文档是否存在
				var count int64
				tx.Model(&models.Document{}).Where("id = ?", docID).Count(&count)
				if count == 0 {
					return errors.New(constant.DocumentNotExist)
				}

				postDocuments = append(postDocuments, models.PostDocument{
					PostID:     post.ID,
					DocumentID: docID,
				})
			}

			if err := tx.Create(&postDocuments).Error; err != nil {
				return errors.New(constant.CreatePostDocumentFailed)
			}
		}

		return nil
	})
}

// GetPostByID 根据ID获取帖子
func GetPostByID(postID uint64) (models.Post, error) {
	db := config.GetDB()
	var post models.Post
	err := db.Where("id = ?", postID).First(&post).Error
	return post, err
}

// GetPostList 获取帖子列表
func GetPostList(key string, order string) ([]models.Post, error) {
	db := config.GetDB()
	var posts []models.Post
	query := db.Model(&models.Post{})

	// 1. 处理关键词搜索 (标题或内容)
	if key != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+key+"%", "%"+key+"%")
	}

	// 2. 处理排序
	// order: "time" -> 时间倒序 (默认)
	// order: "hot"  -> 收藏量倒序 (也可以按阅读量或点赞量)
	if order == "hot" {
		query = query.Order("collect_count DESC")
	} else {
		// 默认按时间排序
		query = query.Order("created_at DESC")
	}

	err := query.Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}
