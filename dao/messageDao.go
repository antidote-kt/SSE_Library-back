package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
)

// GetAllMessagesBySession 获取特定会话中的所有聊天记录
func GetAllMessagesBySession(sessionID uint64) ([]models.Message, error) {
	db := config.GetDB()
	var messages []models.Message
	err := db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// 将该会话中的所有消息标记为已读
	db.Model(&models.Message{}).Where("session_id = ?", sessionID).Update("status", "read")

	return messages, nil
}

// CreateMessage 创建一条新消息
func CreateMessage(message *models.Message) error {
	db := config.GetDB()
	return db.Create(message).Error
}

// GetLastMessageBySessionID 获取会话的最后一条消息
func GetLastMessageBySessionID(sessionID uint64) (models.Message, error) {
	db := config.GetDB()
	var message models.Message
	// 按时间倒序取第一条（查询单条记录如果没有结果，会返回非nil错误：RecordNotFound）
	err := db.Where("session_id = ?", sessionID).Order("created_at DESC").First(&message).Error
	if err != nil {
		return message, err
	}
	return message, nil
}

// CountUnreadMessages 统计用户在某会话中的未读消息数
// 逻辑：统计该会话中，发送者不是我(receiverID)，且状态不是'read'的消息
func CountUnreadMessages(sessionID, receiverID uint64) (uint64, error) {
	db := config.GetDB()
	var count int64
	err := db.Model(&models.Message{}).
		Where("session_id = ? AND sender_id != ? AND status != ?", sessionID, receiverID, "read").
		Count(&count).Error
	return uint64(count), err
}

// SearchMessagesByUserAndKeyword 根据用户ID和关键词搜索消息
// 返回值：包含关键词的所有消息列表
// （按时间倒序排列，这样能保证在调用该函数的controller函数中，对返回的消息列表进行循环处理并分组聚合后得到的会话列表中，第一条消息就是最新的）
func SearchMessagesByUserAndKeyword(userID uint64, keyword string) ([]models.Message, error) {
	db := config.GetDB()
	var messages []models.Message

	// SQL 逻辑：
	// 1. 从 messages 表查
	// 2. 关联 sessions 表
	// 3. 条件：消息内容包含 keyword
	// 4. 条件：会话必须包含当前 userID (user1_id = uid OR user2_id = uid)
	// 5. 排序：按创建时间倒序（这样第一条遇到的就是“最新消息”）
	err := db.Model(&models.Message{}).
		Joins("JOIN sessions ON messages.session_id = sessions.id").
		Where("messages.content LIKE ? AND messages.deleted_at IS NULL", "%"+keyword+"%").
		Where("(sessions.user1_id = ? OR sessions.user2_id = ?)", userID, userID).
		Order("messages.created_at DESC").
		Find(&messages).Error

	return messages, err
}
