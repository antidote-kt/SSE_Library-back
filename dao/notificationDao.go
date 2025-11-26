package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// GetNotificationsByUserId 根据用户ID获取通知列表
func GetNotificationsByUserId(userId uint64) ([]models.Notification, error) {
	db := config.GetDB()
	var notifications []models.Notification
	err := db.Where("receiver_id = ?", userId).Order("created_at DESC").Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkNotificationAsRead 将指定通知标记为已读
func MarkNotificationAsRead(notificationID uint64, userID uint64) error {
	db := config.GetDB()

	// 更新指定的通知为已读状态，确保该通知属于指定用户
	result := db.Model(&models.Notification{}).
		Where("id = ? AND receiver_id = ?", notificationID, userID).
		Update("is_read", true)

	if result.Error != nil {
		return result.Error
	}

	// 检查是否更新了记录（即通知是否存在且属于该用户）
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
