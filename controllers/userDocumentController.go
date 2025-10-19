package controllers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WithdrawUpload(c *gin.Context) {
	var request dto.WithdrawUploadDTO
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数解析失败")
		return
	}

	// 查找对应文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
			return
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
			return
		}
	}
	// 判断请求的用户是否是文档的拥有者
	if request.UserID != document.UploaderID {
		response.Fail(c, http.StatusForbidden, nil, "不允许撤回其他人的文档")
		return
	}
	if document.Status != constant.DocumentStatusAudit {
		response.Fail(c, http.StatusBadRequest, nil, "文档不在审核中，不允许撤回")
		return
	}
	document.Status = constant.DocumentStatusWithdraw
	// 查找对应分类
	category, err := dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "分类不存在")
			return
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
			return
		}
	}

	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "文档更新失败")
		return
	}

	responseData := gin.H{
		"name":        document.Name,
		"document_id": document.ID,
		"type":        document.Type,
		"uploadTime":  document.CreatedAt,
		"status":      document.Status,
		"category":    category.Name,
		"collections": document.Collections,
		"readCounts":  document.ReadCounts,
		"URL":         utils.GetFileURL(document.URL),
	}
	response.Success(c, responseData, "撤回成功")

}

// GetUserUploadDocument 获取用户上传的文档列表
func GetUserUploadDocument(c *gin.Context) {
	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		response.Fail(c, http.StatusBadRequest, nil, "缺少userId参数")
		return
	}

	// 将字符串转换为uint64
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "userId参数格式错误")
		return
	}

	// 验证用户是否存在
	_, err = dao.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "用户不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}

	// 查找用户上传的所有文档
	documents, err := dao.GetDocumentsByUploaderID(userID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "数据库查询失败")
		return
	}

	// 构造返回数据数组
	var responseData []gin.H
	for _, doc := range documents {
		// 获取文档分类信息
		category, err := dao.GetCategoryByID(doc.CategoryID)
		if err != nil {
			continue // 跳过无法获取分类的文档
		}

		var fileURL string
		if doc.Type == constant.VideoType {
			fileURL = doc.URL // 视频类型直接返回URL
		} else {
			fileURL = utils.GetFileURL(doc.URL) // 其他类型需要从COS获取完整URL
		}

		responseData = append(responseData, gin.H{
			"name":        doc.Name,
			"document_id": doc.ID,
			"type":        doc.Type,
			"uploadTime":  doc.CreatedAt.Format("2006-01-02 15:04:05"),
			"status":      doc.Status,
			"category":    category.Name,
			"collections": doc.Collections,
			"readCounts":  doc.ReadCounts,
			"URL":         fileURL,
		})
	}

	// 返回用户上传的文档列表
	response.SuccessWithArray(c, responseData, "获取用户上传文档成功")
}
