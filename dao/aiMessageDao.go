package dao

import (
	"errors"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

func CreateAIMessage(aiMessage *models.AIMessage) error {
	db := config.GetDB()
	return db.Create(aiMessage).Error
}

func UpdateAIMessage(aiMessage *models.AIMessage) error {
	db := config.GetDB()
	return db.Save(aiMessage).Error
}

func UpdateAIMessageStatus(messageID uint64, status string) (*models.AIMessage, error) {
	db := config.GetDB()

	var message models.AIMessage
	if err := db.First(&message, messageID).Error; err != nil {
		return nil, err
	}

	message.Status = status
	if err := db.Save(&message).Error; err != nil {
		return nil, err
	}

	return &message, nil
}

func GetLastAIMessageBySessionID(sessionID uint64) (*models.AIMessage, error) {
	db := config.GetDB()
	var aiMessage models.AIMessage
	err := db.Where("ai_sessions_id = ?", sessionID).Order("created_at DESC").First(&aiMessage).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &aiMessage, nil
}

// GetMessagesBySessionId 获取某个会话的所有历史消息（按时间正序排列）
// limit: 获取最近的 N 条记录，防止上下文超出 Token 限制（传-1则取消限制）
func GetMessagesBySessionId(sessionId uint, limit int) ([]models.AIMessage, error) {
	var messages []models.AIMessage
	// 先按时间倒序查出最近的N条，再在内存中或者通过子查询反转为正序（传入 Limit(-1) 会自动忽略 LIMIT SQL 语句，从而查询全部匹配数据）
	err := config.DB.Where("session_id = ?", sessionId).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error

	if err != nil {
		return nil, err
	}

	// 将倒序（最新在前）反转为正序（最旧在前），以符合大模型阅读习惯
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// GetMessageCountBySessionId 获取特定会话的消息总数
func GetMessageCountBySessionId(sessionId uint) (int64, error) {
	var count int64
	err := config.DB.Model(&models.AIMessage{}).Where("session_id = ?", sessionId).Count(&count).Error
	return count, err
}
