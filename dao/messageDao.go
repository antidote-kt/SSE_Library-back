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

// SearchChatMessagesByUser 模糊搜索用户的聊天记录（作为参与者参与的会话）
func SearchChatMessagesByUser(userID uint64, searchKey string) ([]models.Message, error) {
	db := config.GetDB()
	var messages []models.Message

	// 获取用户参与的所有会话（作为User1ID或User2ID）
	var sessions []models.Session
	err := db.Where("user1_id = ? OR user2_id = ?", userID, userID).Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	// 获取会话ID列表
	var sessionIDs []uint64
	for _, session := range sessions {
		sessionIDs = append(sessionIDs, session.ID)
	}

	if len(sessionIDs) == 0 {
		return []models.Message{}, nil
	}

	// 在用户参与的会话中搜索匹配的消息内容
	err = db.Where("session_id IN ? AND content LIKE ?", sessionIDs, "%"+searchKey+"%").Find(&messages).Error
	if err != nil {
		return nil, err
	}
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
	// 按时间倒序取第一条
	err := db.Where("session_id = ?", sessionID).Order("created_at DESC").First(&message).Error
	if err != nil {
		return message, err
	}
	return message, nil
}

// CountUnreadMessages 统计用户在某会话中的未读消息数
// 逻辑：统计该会话中，发送者不是我(receiverID)，且状态不是'read'的消息
func CountUnreadMessages(sessionID, receiverID uint64) (int64, error) {
	db := config.GetDB()
	var count int64
	err := db.Model(&models.Message{}).
		Where("session_id = ? AND sender_id != ? AND status != ?", sessionID, receiverID, "read").
		Count(&count).Error
	return count, err
}
