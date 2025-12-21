package dao

import (
	"errors"
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dto"
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

// GetDocumentsByUploaderID 获取指定用户上传的文档列表
func GetDocumentsByUploaderID(userID uint64) ([]models.Document, error) {
	db := config.GetDB()
	var documents []models.Document
	err := db.Where("uploader_id = ?", userID).Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

// GetDocumentsByPostID 获取指定帖子关联的文档列表
func GetDocumentsByPostID(postID uint64) ([]models.Document, error) {
	db := config.GetDB()
	var documents []models.Document
	// 使用子查询，更清晰和高效
	err := db.Where("id IN (?)",
		db.Table("post_documents").Select("document_id").Where("post_id = ?", postID)).
		Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}

func CreateDocumentWithTx(tx *gorm.DB, document models.Document, tagNames []string) (models.Document, error) {
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
					return models.Document{}, fmt.Errorf("%v: %s", constant.TagCreateFailed, tagName)
				}
			} else {
				// 其他数据库错误
				return models.Document{}, fmt.Errorf("%vv: %s", constant.TagGetFailed, tagName)
			}
		}
		tags = append(tags, tag)
	}
	// 创建文档
	if err := tx.Create(&document).Error; err != nil {
		return models.Document{}, fmt.Errorf(constant.DocumentCreateFail)
	}

	// 创建文档与标签的关联
	for _, tag := range tags {
		// 中间表模型 DocumentTag
		docTag := models.DocumentTag{
			DocumentID: document.ID,
			TagID:      tag.ID,
		}
		if err := tx.Create(&docTag).Error; err != nil {
			return models.Document{}, fmt.Errorf(constant.DocumentTagCreateFailed)
		}
	}
	return document, nil

}

func UpdateDocumentWithTx(tx *gorm.DB, document models.Document) error {
	if document.ID == 0 {
		return errors.New(constant.DocumentIDLack)
	}
	return tx.Save(&document).Error
}

func UpdateDocument(document models.Document) error {
	db := config.GetDB()
	if document.ID == 0 {
		return errors.New(constant.DocumentIDLack)
	}
	return db.Save(&document).Error
}

func DeleteDocumentWithTx(tx *gorm.DB, document models.Document) error {
	if document.ID == 0 {
		return errors.New(constant.DocumentIDLack)
	}
	err := tx.Select("Tags").Delete(&document).Error
	if err != nil {
		return errors.New(constant.DocumentDeletedFailed)
	}
	return nil
}

// SearchDocumentsByParams 根据参数搜索文档，先用key搜索，再进行其他参数过滤
func SearchDocumentsByParams(request dto.SearchDocumentDTO) ([]models.Document, error) {
	db := config.GetDB()
	query := db.Model(&models.Document{})
	// 首先根据key搜索 - 如果TypeOfKey参数传了，则只搜索指定字段，否则搜索全部字段（包括标签）
	if request.Key != nil && *request.Key != "" {
		key := *request.Key
		if request.TypeOfKey != nil {
			// 根据TypeOfKey参数确定搜索字段
			switch *request.TypeOfKey {
			case constant.TypeOfKeyName:
				query = query.Where("name LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyAuthor:
				query = query.Where("author LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyBookISBN:
				query = query.Where("book_isbn LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyIntroduction:
				query = query.Where("introduction LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyTag:
				// 需要JOIN标签表来搜索标签
				query = query.Joins("LEFT JOIN document_tag ON documents.id = document_tag.document_id").
					Joins("LEFT JOIN tags ON document_tag.tag_id = tags.id").
					Where("tags.tag_name LIKE ?", "%"+key+"%").
					Group("documents.id") // 避免因为JOIN导致的重复记录
			default:
				// 如果TypeOfKey不是预设值，默认搜索全部字段（包括标签）
				query = query.Joins("LEFT JOIN document_tag ON documents.id = document_tag.document_id").
					Joins("LEFT JOIN tags ON document_tag.tag_id = tags.id").
					Where("documents.name LIKE ? OR documents.author LIKE ? OR documents.book_isbn LIKE ? OR documents.introduction LIKE ? OR tags.tag_name LIKE ?",
								"%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%").
					Group("documents.id") // 避免因为JOIN导致的重复记录
			}
		} else {
			// 如果没有TypeOfKey参数，按原来的方式搜索全部字段（包括标签）
			query = query.Joins("LEFT JOIN document_tag ON documents.id = document_tag.document_id").
				Joins("LEFT JOIN tags ON document_tag.tag_id = tags.id").
				Where("documents.name LIKE ? OR documents.author LIKE ? OR documents.book_isbn LIKE ? OR documents.introduction LIKE ? OR tags.tag_name LIKE ?",
							"%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%").
				Group("documents.id") // 避免因为JOIN导致的重复记录
		}
	}

	// 根据其他参数进行过滤
	if request.CategoryID != nil {
		query = query.Where("category_id = ?", *request.CategoryID)
	}

	if request.Type != nil && *request.Type != "" {
		query = query.Where("type = ?", *request.Type)
	}

	if request.Year != nil && *request.Year != "" {
		query = query.Where("create_year = ?", *request.Year)
	}

	// 执行查询
	var documents []models.Document
	err := query.Find(&documents).Error
	if err != nil {
		return nil, err
	}

	return documents, nil
}

// GetDocumentList 获取文档列表
// isSuggest: 是否为推荐模式 (true: 返回阅读量前10的文档)
// categoryID: 分类ID查找特定分类的所有文档 (nil: 默认推荐模式)
func GetDocumentList(isSuggest bool, categoryID *uint64) ([]models.Document, error) {
	db := config.GetDB()
	var documents []models.Document
	query := db.Model(&models.Document{})

	// 基础条件：只返回状态为Open的文档，且未删除
	query = query.Where("status = ? AND deleted_at IS NULL", constant.DocumentStatusOpen)

	// 1. 处理分类筛选 (无论是推荐模式还是普通模式，分类筛选如果传了都应该生效)
	// 如果不希望在推荐模式下筛选分类，可以将这段移到 else 分支里
	if categoryID != nil && *categoryID != 0 {
		query = query.Where("category_id = ?", *categoryID)
	} else if isSuggest {
		// 2. 如果没传分类，那么看是否为推荐模式
		// 推荐模式：返回阅读量 (read_counts) 前 10 的文档
		query = query.Order("read_counts DESC").Limit(10)
	} else {
		// 既没传分类id也不是推荐模式，则返回全部文档
		query = query.Order("created_at DESC")
	}

	err := query.Find(&documents).Error
	if err != nil {
		return nil, err
	}
	return documents, nil
}
