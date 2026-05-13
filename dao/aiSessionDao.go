package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

func CreateAISession(aiSession *models.AISession) error {
	db := config.GetDB()
	return db.Create(aiSession).Error
}

func GetAISessionByID(id uint64) (models.AISession, error) {
	var aiSession models.AISession
	db := config.GetDB()
	err := db.First(&aiSession, id).Error
	return aiSession, err
}

func GetAISessionsByUserID(userID uint64) ([]models.AISession, error) {
	var aiSessions []models.AISession
	db := config.GetDB()
	err := db.Where("user_id = ?", userID).
		Order("updated_at DESC").
		Find(&aiSessions).Error
	return aiSessions, err
}

func UpdateAISession(aiSession *models.AISession) error {
	db := config.GetDB()
	return db.Save(aiSession).Error
}

func DeleteAISession(aiSession *models.AISession) error {
	db := config.GetDB()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ai_sessions_id = ?", aiSession.ID).Delete(&models.AIMessage{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(aiSession).Error; err != nil {
			return err
		}
		return nil
	})
}
