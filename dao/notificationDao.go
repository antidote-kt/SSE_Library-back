package dao

import (
	"errors"
	"log"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// CreateNotification 创建通知
func CreateNotification(notification *models.Notification) error {
	db := config.GetDB()
	result := db.Create(notification)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

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

	// 先查询当前通知的状态
	var notification models.Notification
	err := db.Where("id = ? AND receiver_id = ?", notificationID, userID).First(&notification).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gorm.ErrRecordNotFound
		}
		return err
	}

	// 检查是否已经是已读状态
	if notification.IsRead {
		// 如果已经是已读状态，不需要再更新
		return nil
	}

	// 如果不是已读状态，才进行更新
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

// GetUnreadMessageCount 根据用户ID和消息类型获取总的未读消息数量
func GetUnreadMessageCount(userId uint64, messageType string) (int64, error) {
	db := config.GetDB()

	// 检查消息类型是否有效
	if messageType != "message" && messageType != "reminder" {
		return 0, errors.New(constant.InvalidMessageType)
	}

	var count int64
	// 根据消息类型查询未读消息数量
	if messageType == "message" { // 如果要获取总的未读聊天信息数
		//以下是N+1查询低性能做法，仅供参考，采取后续联表查询的做法
		//// 先查找跟用户相关的会话
		//var unreadSession []models.Session
		//err := db.Model(models.Session{}).Where("user1_id = ? OR user2_id = ?", userId, userId).Find(&unreadSession)
		//if err != nil {
		//	log.Println(err.Error)
		//	return 0, errors.New(constant.DatabaseError)
		//}
		//
		//// 再查找每个会话的未读消息数量（限制只统计sender_id不是用户本人的message,这样保证了只查用户接收的信息）
		//var unreadMessageCount int64
		//for _, session := range unreadSession {
		//	err = db.Model(models.Message{}).Where("session_id = ? AND status = ? AND sender_id != ?", session.ID, "unread", userId).Count(&unreadMessageCount)
		//	if err != nil {
		//		log.Println(err.Error)
		//		return 0, errors.New(constant.DatabaseError)
		//	}
		//	count += unreadMessageCount // 循环累加
		//}

		// 【优化】使用 JOIN 一次性查询所有相关会话的未读消息
		// 逻辑：统计消息表，条件是：
		// 1. 消息属于该用户参与的会话 (联表 sessions)
		// 2. 消息状态是 unread
		// 3. 发送者不是自己 (sender_id != userId)
		err := db.Model(&models.Message{}).
			Joins("JOIN sessions ON messages.session_id = sessions.id").
			Where("sessions.user1_id = ? OR sessions.user2_id = ?", userId, userId). // 确保是该用户的会话
			Where("messages.status = ?", "unread").                                  // 状态未读
			Where("messages.sender_id != ?", userId).                                // 发送者不是自己
			Count(&count).Error                                                      // 统计数量
		// GORM 在使用 db.Model(&models.Message{}) 配合 Joins 和 Count 时会自动生成：
		// SELECT count(*)
		// FROM `messages`
		// JOIN sessions ON messages.session_id = sessions.id
		// WHERE `messages`.`deleted_at` IS NULL

		if err != nil {
			log.Println("查询聊天未读数失败:", err)
			return 0, errors.New(constant.DatabaseError)
		}

		return count, nil
	} else { // 如果要获取总的未读提醒数（如果能执行到这一步说明messageType一定为reminder，无需if检查）
		err := db.Model(models.Notification{}).Where("receiver_id = ? AND is_read = ?", userId, false).Count(&count).Error
		if err != nil {
			log.Println("查询通知未读数失败:", err)
			return 0, errors.New(constant.DatabaseError)
		}

		return count, nil
	}

}
