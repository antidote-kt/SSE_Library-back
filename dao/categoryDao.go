package dao

import (
	"time"

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

// GetAllCategories 获取所有分类和课程
func GetAllCategories() ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	err := db.Where("deleted_at IS NULL").Find(&categories).Error
	return categories, err
}

// CountDocumentsByCategory 统计分类下的文档数量
func CountDocumentsByCategory(categoryID uint64) (int64, error) {
	db := config.GetDB()
	var count int64
	err := db.Model(&models.Document{}).Where("category_id = ? AND deleted_at IS NULL", categoryID).Count(&count).Error
	return count, err
}

// GetDocumentReadCountsByCategory 获取分类下所有文档的总浏览量
func GetDocumentReadCountsByCategory(categoryID uint64) (int64, error) {
	db := config.GetDB()
	var totalReads int64
	err := db.Model(&models.Document{}).
		Where("category_id = ? AND deleted_at IS NULL", categoryID).
		Select("COALESCE(SUM(read_counts), 0)").
		Scan(&totalReads).Error
	return totalReads, err
}

// SearchCategoriesByName 根据名称搜索分类和课程
func SearchCategoriesByName(name string) ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	err := db.Where("name LIKE ? AND deleted_at IS NULL", "%"+name+"%").Find(&categories).Error
	return categories, err
}

// GetRecentCategories 获取最近一段时间内有更新的分类（用于热度计算）
func GetRecentCategories(days int) ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	cutoffTime := time.Now().AddDate(0, 0, -days)
	err := db.Where("updated_at >= ? AND deleted_at IS NULL", cutoffTime).Find(&categories).Error
	return categories, err
}

// CreateCategory 创建新的分类/课程
func CreateCategory(category *models.Category) error {
	db := config.GetDB()
	return db.Create(category).Error
}

// DeleteCategoryByName 根据名称删除分类（软删除）
func DeleteCategoryByName(name string) error {
	db := config.GetDB()
	err := db.Model(&models.Category{}).Where("name = ? AND deleted_at IS NULL", name).Update("deleted_at", time.Now()).Error
	return err
}

// UpdateCategory 更新分类信息
func UpdateCategory(category *models.Category) error {
	db := config.GetDB()
	err := db.Save(category).Error
	return err

}

// GetCategoriesByParentID 根据父分类ID获取所有子分类
func GetCategoriesByParentID(parentID uint64) ([]models.Category, error) {
	db := config.GetDB()
	var categories []models.Category
	err := db.Where("parent_id = ? AND deleted_at IS NULL", parentID).Find(&categories).Error
	return categories, err
}
