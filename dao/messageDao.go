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
