package controllers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 构造文档信息的公共函数
func constructDocumentInfo(doc models.Document, category models.Category) gin.H {
	var fileURL string
	if doc.Type == constant.VideoType {
		fileURL = doc.URL // 视频类型直接返回URL
	} else {
		fileURL = utils.GetFileURL(doc.URL) // 其他类型需要从COS获取完整URL
	}

	return gin.H{
		"name":        doc.Name,
		"document_id": doc.ID,
		"type":        doc.Type,
		"uploadTime":  doc.CreatedAt.Format("2006-01-02 15:04:05"),
		"status":      doc.Status,
		"category":    category.Name,
		"collections": doc.Collections,
		"readCounts":  doc.ReadCounts,
		"URL":         fileURL,
	}
}

func CollectDocument(c *gin.Context) {
	var request dto.FavoriteDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数错误")
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(request.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "用户不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}

	// 检查是否已经收藏（防止重复收藏）
	exists, err := dao.CheckFavoriteExist(request.UserID, request.DocumentID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "检查收藏状态失败")
		return
	}
	if exists {
		response.Fail(c, http.StatusBadRequest, nil, "文档已收藏")
		return
	}

	// 使用事务处理收藏操作
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 创建收藏记录
		favorite := models.Favorite{
			UserID:     request.UserID,
			DocumentID: request.DocumentID,
		}
		if err := tx.Create(&favorite).Error; err != nil {
			return fmt.Errorf("创建收藏记录失败: %v", err)
		}

		// 更新文档的收藏数
		document.Collections++
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("更新文档收藏数失败: %v", err)
		}

		return nil
	})

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(request.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "获取收藏文档列表失败")
		return
	}

	// 构造返回数据数组
	var responseData []gin.H
	for _, favDoc := range favoriteDocuments {
		// 获取文档分类信息
		category, err := dao.GetCategoryByID(favDoc.CategoryID)
		if err != nil {
			continue // 跳过无法获取分类的文档
		}

		responseData = append(responseData, constructDocumentInfo(favDoc, category))
	}

	// 返回收藏的文档列表
	response.SuccessWithArray(c, responseData, "收藏成功")
}

func WithdrawCollection(c *gin.Context) {
	var request dto.FavoriteDTO
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数错误")
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(request.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "用户不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}

	// 验证文档是否存在
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}

	// 检查是否已收藏（必须已收藏才能取消收藏）
	exists, err := dao.CheckFavoriteExist(request.UserID, request.DocumentID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "检查收藏状态失败")
		return
	}
	if !exists {
		response.Fail(c, http.StatusBadRequest, nil, "文档未被收藏")
		return
	}

	// 使用事务处理取消收藏操作
	db := config.GetDB()
	err = db.Transaction(func(tx *gorm.DB) error {
		// 删除收藏记录
		if err := tx.Where("user_id = ? AND document_id = ?", request.UserID, request.DocumentID).Delete(&models.Favorite{}).Error; err != nil {
			return fmt.Errorf("删除收藏记录失败: %v", err)
		}

		// 更新文档的收藏数（减少1）
		document.Collections--
		if document.Collections < 0 {
			document.Collections = 0 // 确保收藏数不会小于0
		}
		if err := tx.Save(&document).Error; err != nil {
			return fmt.Errorf("更新文档收藏数失败: %v", err)
		}

		return nil
	})

	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 获取用户收藏的所有文档
	favoriteDocuments, err := dao.GetFavoriteDocumentsByUserID(request.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "获取收藏文档列表失败")
		return
	}

	// 构造返回数据数组
	var responseData []gin.H
	for _, favDoc := range favoriteDocuments {
		// 获取文档分类信息
		category, err := dao.GetCategoryByID(favDoc.CategoryID)
		if err != nil {
			continue // 跳过无法获取分类的文档
		}

		responseData = append(responseData, constructDocumentInfo(favDoc, category))
	}

	// 返回剩余的收藏文档列表
	response.SuccessWithArray(c, responseData, "取消收藏成功")
}
