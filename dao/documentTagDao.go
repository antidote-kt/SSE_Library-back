package dao

import (
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
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
				return fmt.Errorf("%v: %s", constant.TagCreateFailed, tagName)
			}
		}

		// 创建文档标签关联
		documentTag := models.DocumentTag{
			DocumentID: documentID,
			TagID:      tag.ID,
		}
		if err := tx.Create(&documentTag).Error; err != nil {
			return fmt.Errorf(constant.DocumentTagCreateFailed)
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

	// 通过中间表查询关联的标签ID
	var documentTags []models.DocumentTag
	err := db.Where("document_id = ?", documentID).Find(&documentTags).Error
	if err != nil {
		return nil, fmt.Errorf(constant.DocumentTagGetFailed)
	}

	// 提取标签ID
	var tagIDs []uint64
	for _, docTag := range documentTags {
		tagIDs = append(tagIDs, docTag.TagID)
	}

	// 如果没有关联标签，返回空数组
	if len(tagIDs) == 0 {
		return []models.Tag{}, nil
	}

	// 查询标签信息，使用DISTINCT确保不重复
	var tags []models.Tag
	err = db.Where("id IN ?", tagIDs).Find(&tags).Error
	if err != nil {
		return nil, fmt.Errorf(constant.TagGetFailed)
	}

	return tags, nil
}
