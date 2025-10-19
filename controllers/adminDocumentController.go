package controllers

import (
	"errors"
	"net/http"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func AdminModifyDocumentStatus(c *gin.Context) {
	var request dto.AdminModifyDocumentStatusRequest
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}
	//查询对应document
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	if request.Status != nil {
		document.Status = *request.Status
	}
	if request.Name != nil {
		document.Name = *request.Name
	}
	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentStatusUpdateFailed)
		return
	}

	response.Success(c, nil, constant.DocumentStatusUpdateSuccess)
}
