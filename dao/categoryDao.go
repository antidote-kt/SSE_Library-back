package dao

import (
	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
)

func GetCategoryByName(name string) ([]models.Category, error) {
	db := config.GetDB()
	var category []models.Category
	err := db.Where("name = ?", name).Find(&category).Error
	if err != nil {
		return nil, err
	}
	return category, nil
}
func GetCategoryByID(id uint64) (models.Category, error) {
	db := config.GetDB()
	var category models.Category
	err := db.Where("id = ?", id).First(&category).Error
	if err != nil {
		return models.Category{}, err
	}
	return category, nil
}

// GetCategoriesByIDs 根据一组ID批量获取分类
func GetCategoriesByIDs(ids []uint64) ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	// 如果传入的ID列表为空，直接返回空切片，避免无效的数据库查询
	if len(ids) == 0 {
		return categories, nil
	}
	// 使用 "id IN ?" 进行批量查询，并将结果填充到categories切片中
	err := db.Where("id IN ?", ids).Find(&categories).Error
	return categories, err
}
