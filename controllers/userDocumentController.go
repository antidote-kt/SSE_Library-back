package controllers

import (
	"errors"
	"net/http"

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
