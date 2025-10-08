package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/config"
	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func WithdrawUpload(c *gin.Context) {
	db := config.GetDB()
	var request dto.WithdrawUploadDTO
	if err := c.ShouldBindQuery(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数解析失败")
	}

	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		}
	}
	if document.Status != constant.DocumentStatusAudit {
		response.Fail(c, http.StatusBadRequest, nil, "文档已审核，不允许撤回")
	}
	category, err := dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "分类不存在")
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		}
	}
	if document.URL != "" {
		err := utils.DeleteFile(document.URL)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "删除文档失败")
		}
	}
	if document.Cover != "" {
		err := utils.DeleteFile(document.Cover)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, "删除封面失败")
		}
	}
	err = db.Transaction(func(tx *gorm.DB) error {
		err := dao.DeleteDocumentWithTx(tx, document)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
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
