package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminModifyDocumentStatus(c *gin.Context) {
	var request dto.AdminModifyDocumentStatusRequest
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, "参数错误")
		return
	}
	//查询对应document
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, "文档不存在")
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, "数据库错误")
		return
	}
	if request.Status != nil {
		document.Status = *request.Status
	}
	if request.Name != nil {
		document.Name = *request.Name
	}
	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, "文档更新失败")
		return
	}

	response.Success(c, nil, "文档状态更新成功")
}
