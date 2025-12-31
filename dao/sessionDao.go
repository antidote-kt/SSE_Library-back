package dao

import (
	"errors"
	"time"

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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return session, err
		}
		return session, err
	}
	return session, nil
}

// GetSessionByUsers 根据两个用户ID查找他们之间的会话
func GetSessionByUsers(user1ID, user2ID uint64) (*models.Session, error) {
	db := config.GetDB()
	var session models.Session
	// 查询 user1=u1 AND user2=u2 或者 user1=u2 AND user2=u1
	err := db.Where("(user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)",
		user1ID, user2ID, user2ID, user1ID).First(&session).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // 未找到会话（这是允许的，不属于查询错误，因此返回nil）
		}
		return nil, err
	}
	return &session, nil
}

// CreateSession 创建一个新的会话
func CreateSession(session *models.Session) error {
	db := config.GetDB()
	var existsSession models.Session
	// 检查会话是否已经存在，已存在则直接返回niu不创建新会话
	db.Where("user1_id = ? AND user2_id = ?", session.User1ID, session.User2ID).
		Or("user1_id = ? AND user2_id = ?", session.User2ID, session.User1ID).
		Find(&existsSession)
	if existsSession.ID != 0 {
		return nil
	}
	// 否则创建新会话再返回
	return db.Create(session).Error
}

// GetUserSessions 获取用户参与的所有会话
func GetUserSessions(userID uint64) ([]models.Session, error) {
	db := config.GetDB()
	var sessions []models.Session
	// 查询作为 user1 或 user2 参与的所有会话
	err := db.Where("user1_id = ? OR user2_id = ?", userID, userID).
		Order("updated_at DESC"). // 按最后更新时间排序
		Find(&sessions).Error
	return sessions, err
}

// UpdateSessionTime 更新会话的最后活动时间
func UpdateSessionTime(sessionID uint64) error {
	db := config.GetDB()
	// 只更新 updated_at 字段为当前时间
	return db.Model(&models.Session{}).Where("id = ?", sessionID).Update("updated_at", time.Now()).Error
}
