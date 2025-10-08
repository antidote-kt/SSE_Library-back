package dao

import (
	"errors"
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

func GetDocumentByID(id uint64) (models.Document, error) {
	db := config.GetDB()
	var document models.Document
	err := db.Where("id = ?", id).First(&document).Error
	if err != nil {
		return models.Document{}, err
	}
	return document, nil
}

func CreateDocumentWithTx(tx *gorm.DB, document models.Document, tagNames []string) error {
	var tags []models.Tag
	for _, tagName := range tagNames {
		var tag models.Tag
		if err := tx.Where("tag_name = ?", tagName).First(&tag).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 标签不存在，创建新标签
				tag = models.Tag{
					TagName: tagName,
				}
				if err := tx.Create(&tag).Error; err != nil {
					return fmt.Errorf("创建标签失败: %s, 错误: %v", tagName, err)
				}
			} else {
				// 其他数据库错误
				return fmt.Errorf("查询标签失败: %s, 错误: %v", tagName, err)
			}
		}
		tags = append(tags, tag)
	}
	// 创建文档
	if err := tx.Create(&document).Error; err != nil {
		return fmt.Errorf("创建文档失败: %v", err)
	}

	// 创建文档与标签的关联
	for _, tag := range tags {
		// 中间表模型 DocumentTag
		docTag := models.DocumentTag{
			DocumentID: document.ID,
			TagID:      tag.ID,
		}
		if err := tx.Create(&docTag).Error; err != nil {
			return fmt.Errorf("创建文档标签关联失败: %v", err)
		}
	}
	return nil

}

func UpdateDocumentWithTx(tx *gorm.DB, document models.Document) error {
	if document.ID == 0 {
		return errors.New("文档ID不能为空")
	}
	return tx.Save(&document).Error
}

func DeleteDocumentWithTx(tx *gorm.DB, document models.Document) error {
	if document.ID == 0 {
		return errors.New("文档ID不能为空")
	}
	err := tx.Select("Tags").Delete(&document).Error
	if err != nil {
		return errors.New("文档删除失败")
	}
	return nil
}
