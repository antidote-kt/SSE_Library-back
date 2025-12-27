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

	// 使用 JOIN 查询，只获取 documents 表的数据
	// 同时也过滤了 post_documents 表的软删除记录 (如果存在 deleted_at)
	// 以及 documents 表本身必须是 Open 状态且未删除
	err := db.Model(&models.Document{}).
		Select("documents.*"). // 显式指定只查询 documents 表的字段，防止 ID 被 post_documents.id 覆盖
		Joins("JOIN post_documents ON post_documents.document_id = documents.id").
		Where("post_documents.post_id = ? AND post_documents.deleted_at IS NULL", postID).
		Where("documents.status = ? AND documents.deleted_at IS NULL", constant.DocumentStatusOpen).
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

// IncrementDocumentViewCount 增加文档浏览量
func IncrementDocumentViewCount(id uint64) error {
	db := config.GetDB()
	// 使用 UpdateColumn 进行原子更新，避免并发冲突，且不更新 UpdatedAt 时间
	return db.Model(&models.Document{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// SearchDocumentsByParams 根据参数搜索文档，先用key搜索，再进行其他参数过滤
func SearchDocumentsByParams(request dto.SearchDocumentDTO) ([]models.Document, error) {
	db := config.GetDB()
	// 使用表别名简化查询
	query := db.Table("documents AS d")
	// 基础条件：只查询状态为Open的文档，且未删除
	query = query.Where("d.status = ? AND d.deleted_at IS NULL", constant.DocumentStatusOpen)
	// 首先根据key搜索 - 如果TypeOfKey参数传了，则只搜索指定字段，否则搜索全部字段（包括标签）
	if request.Key != nil && *request.Key != "" {
		key := *request.Key
		if request.TypeOfKey != nil {
			// 根据TypeOfKey参数确定搜索字段
			switch *request.TypeOfKey {
			case constant.TypeOfKeyName:
				query = query.Where("d.name LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyAuthor:
				query = query.Where("d.author LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyBookISBN:
				query = query.Where("d.book_isbn LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyIntroduction:
				query = query.Where("d.introduction LIKE ?", "%"+key+"%")
			case constant.TypeOfKeyTag:
				// 需要JOIN标签表来搜索标签
				query = query.Joins("LEFT JOIN document_tag AS dt ON d.id = dt.document_id AND dt.deleted_at IS NULL").
					Joins("LEFT JOIN tags AS t ON dt.tag_id = t.id AND t.deleted_at IS NULL").
					Where("t.tag_name LIKE ?", "%"+key+"%").
					Group("d.id") // 避免因为JOIN导致的重复记录
			default:
				// 如果TypeOfKey不是预设值，默认搜索全部字段（包括标签）
				query = query.Joins("LEFT JOIN document_tag AS dt ON d.id = dt.document_id AND dt.deleted_at IS NULL").
					Joins("LEFT JOIN tags AS t ON dt.tag_id = t.id AND t.deleted_at IS NULL").
					Where("d.name LIKE ? OR d.author LIKE ? OR d.book_isbn LIKE ? OR d.introduction LIKE ? OR t.tag_name LIKE ?",
							"%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%").
					Group("d.id") // 避免因为JOIN导致的重复记录
			}
		} else {
			// 如果没有TypeOfKey参数，按原来的方式搜索全部字段（包括标签）
			query = query.Joins("LEFT JOIN document_tag AS dt ON d.id = dt.document_id AND dt.deleted_at IS NULL").
				Joins("LEFT JOIN tags AS t ON dt.tag_id = t.id AND t.deleted_at IS NULL").
				Where("d.name LIKE ? OR d.author LIKE ? OR d.book_isbn LIKE ? OR d.introduction LIKE ? OR t.tag_name LIKE ?",
						"%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%", "%"+key+"%").
				Group("d.id") // 避免因为JOIN导致的重复记录
		}
	}

	// 根据其他参数进行过滤
	if request.CategoryID != nil {
		query = query.Where("d.category_id = ?", *request.CategoryID)
	}

	if request.Type != nil && *request.Type != "" && *request.Type != "null" {
		query = query.Where("d.type = ?", *request.Type)
	}

	if request.Year != nil && *request.Year != "" {
		query = query.Where("d.create_year = ?", *request.Year)
	}

	// 执行查询
	var documents []models.Document
	err := query.Find(&documents).Error
	if err != nil {
		return nil, err
	}

	return documents, nil
}

// GetAllDocuments 获取所有未删除的文档（管理员使用，不过滤状态）
func GetAllDocuments() ([]models.Document, error) {
	db := config.GetDB()
	var documents []models.Document

	// 只过滤软删除的文档，不过滤状态
	err := db.Model(&models.Document{}).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Find(&documents).Error

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
	query = query.Where("deleted_at IS NULL")

	// 1. 处理分类筛选 (无论是推荐模式还是普通模式，分类筛选如果传了都应该生效)
	// 如果不希望在推荐模式下筛选分类，可以将这段移到 else 分支里
	if categoryID != nil && *categoryID != 0 {
		// 获取该分类下的所有子分类（课程）
		subCategories, err := GetCategoriesByParentID(*categoryID)
		if err != nil {
			return nil, err
		}

		// 构建分类ID列表：包括当前分类和所有子分类（课程）
		categoryIDs := []uint64{*categoryID}
		for _, subCategory := range subCategories {
			categoryIDs = append(categoryIDs, subCategory.ID)
		}

		// 查询该分类及其所有子分类（课程）下的文档
		query = query.Where("category_id IN ?", categoryIDs)
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
