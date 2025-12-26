package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

// CreateUser 创建新用户
func CreateUser(user models.User) (models.User, error) {
	db := config.GetDB()
	err := db.Create(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func GetUserByID(id uint64) (models.User, error) {
	db := config.GetDB()
	var user models.User
	err := db.Where("id = ? and status = ?", id, "active").First(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func AdminGetUserByID(id uint64) (models.User, error) {
	db := config.GetDB()
	var user models.User
	err := db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// GetUserByUsername 根据用户名查询用户
func GetUserByUsername(username string) (models.User, error) {
	db := config.GetDB()
	var user models.User
	// 直接执行查询，并将原始的 err 返回
	err := db.Where("username = ? and status = ?", username, "active").First(&user).Error
	return user, err // 直接返回 user 和可能发生的任何错误（包括 gorm.ErrRecordNotFound）
}

// GetUserByEmail 根据邮箱获取用户
func GetUserByEmail(email string) (models.User, error) {
	db := config.GetDB()
	var user models.User
	err := db.Where("email = ? and status = ?", email, "active").First(&user).Error
	return user, err
}

// UpdateUser 更新用户信息
func UpdateUser(user models.User) error {
	db := config.GetDB()
	return db.Save(&user).Error
}

// GetUsers 管理员获取/搜索用户列表
func GetUsers(username *string, userID *uint64, role string) ([]models.User, error) {
	db := config.GetDB()
	var users []models.User
	query := db.Model(&models.User{})

	// 普通用户搜索其他用户，只能搜索状态为“active”的用户
	if role == "user" {
		// 如果提供了username作为搜索条件
		if username != nil && *username != "" {
			query = query.Where("username LIKE ? and status = ?", "%"+*username+"%", "active") // 模糊搜索
		}

		// 如果提供了userID作为搜索条件
		if userID != nil && *userID != 0 {
			query = query.Where("id = ? and status = ?", *userID, "active")
		}
	}
	// 管理员搜索用户，可以搜索所有用户
	if role == "admin" {
		// 如果提供了username作为搜索条件
		if username != nil && *username != "" {
			query = query.Where("username LIKE ?", "%"+*username+"%") // 模糊搜索
		}

		// 如果提供了userID作为搜索条件
		if userID != nil && *userID != 0 {
			query = query.Where("id = ?", *userID)
		}
	}

	err := query.Find(&users).Error // 如果两者都没提供（即获取用户列表接口），则会查询所有数据，也就是获取所有用户。

	// 如果发生数据库错误，直接返回
	if err != nil {
		return nil, err
	}

	// 如果查询成功但没有找到任何记录，手动返回 gorm.ErrRecordNotFound
	if len(users) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return users, nil
}
