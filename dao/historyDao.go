package dao

import (
	"errors"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// AddViewHistory 添加或更新浏览历史
// sourceType: "document" 或 "post"
func AddViewHistory(userID, sourceID uint64, sourceType string) error {
	db := config.GetDB()
	var history models.ViewHistory

	// 1. 查询是否存在记录
	// 注意：唯一索引是 (UserID, SourceID, SourceType) 的组合
	err := db.Where("user_id = ? AND source_id = ? AND source_type = ?", userID, sourceID, sourceType).
		First(&history).Error

	if err == nil {
		// 2. 如果存在，更新 UpdatedAt 为当前时间
		history.UpdatedAt = time.Now()
		return db.Save(&history).Error
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 3. 如果不存在，创建新记录
		newHistory := models.ViewHistory{
			UserID:     userID,
			SourceID:   sourceID,
			SourceType: sourceType,
		}
		return db.Create(&newHistory).Error
	} else {
		return err
	}
}

// GetUserViewHistory 获取用户的浏览历史列表（通用）
// 可以通过 sourceType 过滤，传 "" 则查所有
func GetUserViewHistory(userID uint64, sourceType string, page, pageSize int) ([]models.ViewHistory, int64, error) {
	db := config.GetDB()
	var histories []models.ViewHistory
	var total int64

	offset := (page - 1) * pageSize

	query := db.Model(&models.ViewHistory{}).Where("user_id = ?", userID)

	if sourceType != "" {
		query = query.Where("source_type = ?", sourceType)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// 按 UpdatedAt (最近浏览时间) 倒序排列
	err = query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&histories).Error

	return histories, total, err
}
