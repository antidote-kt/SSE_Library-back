package dao

import (
	"errors"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

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
