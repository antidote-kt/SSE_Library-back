package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminModifyDocument(c *gin.Context) {
	var request dto.AdminModifyDocumentDTO
	var category models.Category
	db := config.GetDB()
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "参数错误")
		return
	}

	// 查找要修改的文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, gin.H{}, "文档不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, gin.H{}, "数据库查询失败")
		return
	}

	// 更新文档字段（如果提供了相应字段）
	if request.Author != nil {
		document.Author = *request.Author
	}
	if request.Category != nil {
		categories, err := dao.GetCategoryByName(*request.Category)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		}
		if len(categories) == 0 {
			response.Fail(c, http.StatusNotFound, nil, "分类不存在")
		}
		document.CategoryID = categories[0].ID
		category = categories[0]
	} else {
		category, err = dao.GetCategoryByID(document.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, nil, "分类不存在")
			}
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		}
	}

	if request.CreateYear != nil {
		document.CreateYear = *request.CreateYear
	}
	if request.ISBN != nil {
		document.BookISBN = *request.ISBN
	}
	if request.Name != nil {
		document.Name = *request.Name
	}
	if request.Type != nil {
		document.Type = *request.Type
	}
	if request.UploadTime != nil {
		document.CreatedAt, err = time.Parse("2006-01-02 15:04:05", *request.UploadTime)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "时间格式错误")
		}
	}
	if request.Cover != nil {
		err := utils.DeleteFile(document.Cover)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "旧封面删除失败")
		}
		document.Cover, err = utils.UploadCoverImage(request.Cover, category.Name)
	}
	if request.File != nil {
		err := utils.DeleteFile(document.URL)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "旧文件删除失败")
		}
		document.URL, err = utils.UploadMainFile(request.File, category.Name)
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := dao.UpdateDocumentWithTx(tx, document); err != nil {
			return err
		}
		if len(request.Tags) > 0 {
			// 删除原有的标签映射
			if err := dao.DeleteDocumentTagByDocumentIDWithTx(tx, document.ID); err != nil {
				return fmt.Errorf("标签更新失败")
			}
			// 创建新标签映射
			if err := dao.CreateDocumentTagWithTx(tx, document.ID, request.Tags); err != nil {
				return err
			}

		}
		return nil
	})
	// 检查事务执行结果
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 构造返回数据
	responseData := gin.H{
		"infoBrief": gin.H{
			"name":        document.Name,
			"document_id": document.ID,
			"type":        document.Type,
			"uploadTime":  document.CreatedAt.Format("2006-01-02 15:04:05"),
			"status":      document.Status,
			"category":    category.Name,
			"collections": document.Collections,
			"readCounts":  document.ReadCounts,
			"URL":         utils.GetFileURL(document.URL),
		},
		"bookISBN":     document.BookISBN,
		"author":       document.Author,
		"Cover":        utils.GetFileURL(document.Cover),
		"introduction": document.Introduction,
		"createYear":   document.CreateYear,
	}

	// 获取文档标签列表

	tags, err := dao.GetDocumentTagByDocumentID(document.ID)
	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.TagName)
	}
	responseData["tags"] = tagNames

	response.Success(c, responseData, "文档修改成功")
}

func AdminModifyDocumentStatus(c *gin.Context) {
	var request dto.AdminModifyDocumentStatusRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, gin.H{}, "参数错误")
		return
	}
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
	}
	document.Type = request.Type
	document.Status = request.NewStatus
	if request.Name != nil {
		document.Name = *request.Name
	}
	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "文档更新失败")
	}
}
