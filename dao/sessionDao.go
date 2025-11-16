package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// GetSessionByID 根据ID获取会话信息
func GetSessionByID(sessionID uint64) (models.Session, error) {
	var session models.Session
	db := config.GetDB()
	err := db.First(&session, sessionID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return session, err
		}
		return session, err
	}
	return session, nil
}
