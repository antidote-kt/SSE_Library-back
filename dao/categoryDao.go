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

func GetCategoriesByIDs(ids []uint64) ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	err := db.Where("id IN (?)", ids).Find(&categories).Error
	if err != nil {
		return []models.Category{}, err
	}
	return categories, nil
}
