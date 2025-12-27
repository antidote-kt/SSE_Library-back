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

// IncrementPostViewCount 增加帖子浏览量
// 使用 UpdateColumn 进行原子更新 (view_count = view_count + 1)
func IncrementPostViewCount(id uint64) error {
	db := config.GetDB()
	return db.Model(&models.Post{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// LikePost 点赞帖子 (事务处理：插入记录 + 计数+1)
func LikePost(userID, postID uint64) error {
	db := config.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 检查是否已经点赞
		var count int64
		tx.Model(&models.PostLike{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count)
		if count > 0 {
			return errors.New(constant.AlreadyLiked) // 避免重复点赞
		}

		// 2. 创建点赞记录
		newLike := models.PostLike{
			UserID: userID,
			PostID: postID,
		}
		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}

		// 3. 帖子点赞数 +1 (原子操作)
		if err := tx.Model(&models.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

// UnlikePost 取消点赞 (事务处理：删除记录 + 计数-1)
func UnlikePost(userID, postID uint64) error {
	db := config.GetDB()

	return db.Transaction(func(tx *gorm.DB) error {
		// 1. 删除点赞记录 (Unscoped 硬删除，或者使用软删除均可，这里建议硬删除以节省空间)
		// 注意：GORM 的 Delete 需要 Where 条件
		result := tx.Where("user_id = ? AND post_id = ?", userID, postID).Unscoped().Delete(&models.PostLike{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("not liked yet") // 还没点赞过
		}

		// 2. 帖子点赞数 -1 (防止减为负数)
		if err := tx.Model(&models.Post{}).
			Where("id = ? AND like_count > 0", postID).
			UpdateColumn("like_count", gorm.Expr("like_count - ?", 1)).Error; err != nil {
			return err
		}

		return nil
	})
}

// IsPostLikedByUser 检查用户是否已点赞
func IsPostLikedByUser(userID, postID uint64) (bool, error) {
	db := config.GetDB()
	var count int64
	// 查询是否存在记录
	err := db.Model(&models.PostLike{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error

	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// IncrementPostCommentCount 增加帖子的评论数
func IncrementPostCommentCount(postID uint64) error {
	db := config.GetDB()
	return db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
}

// DecrementPostCommentCount 减少帖子的评论数
func DecrementPostCommentCount(postID uint64) error {
	db := config.GetDB()
	return db.Model(&models.Post{}).
		Where("id = ?", postID).
		Update("comment_count", gorm.Expr("CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END")).Error
}

// GetPostsByUserID 获取指定用户发布的帖子列表
func GetPostsByUserID(userID uint64) ([]models.Post, error) {
	db := config.GetDB()
	var posts []models.Post

	err := db.Model(&models.Post{}).
		Where("sender_id = ? AND deleted_at IS NULL", userID).
		Order("created_at DESC").
		Find(&posts).Error

	if err != nil {
		return nil, err
	}
	return posts, nil
}
