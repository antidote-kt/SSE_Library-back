package dao

import (
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/models"
	"gorm.io/gorm"
)

func CreateDocumentTagWithTx(tx *gorm.DB, documentID uint64, tags []string) error {
	for _, tagName := range tags {
		// 查找或创建标签
		var tag models.Tag
		if err := tx.Where("tag_name = ?", tagName).First(&tag).Error; err != nil {
			// 标签不存在，创建新标签
			tag = models.Tag{
				TagName: tagName,
			}
			if err := tx.Create(&tag).Error; err != nil {
				return fmt.Errorf("创建标签失败: %s", tagName)
			}
		}

		// 创建文档标签关联
		documentTag := models.DocumentTag{
			DocumentID: documentID,
			TagID:      tag.ID,
		}
		if err := tx.Create(&documentTag).Error; err != nil {
			return fmt.Errorf("创建标签关联失败")
		}
	}
	return nil
}

func DeleteDocumentTagByDocumentIDWithTx(tx *gorm.DB, documentID uint64) error {
	if err := tx.Where("document_id = ?", documentID).Delete(&models.DocumentTag{}).Error; err != nil {
		return err
	}
	return nil
}

func GetDocumentTagByDocumentID(documentID uint64) ([]models.Tag, error) {
	db := config.GetDB()
	var document models.Document
	err := db.Preload("Tags").First(&document, documentID).Error
	if err != nil {
		return nil, fmt.Errorf("获取文档失败: %v", err)
	}
	return document.Tags, nil
}
