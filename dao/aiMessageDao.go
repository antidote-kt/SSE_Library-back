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
