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

// AdminModifyDocumentStatus 管理员修改文档状态控制器
// 该函数允许管理员修改文档的状态和名称
func AdminModifyDocumentStatus(c *gin.Context) {
	// 初始化管理员修改文档状态请求结构体
	var request dto.AdminModifyDocumentStatusRequest

	// 绑定并验证请求参数
	if err := c.ShouldBind(&request); err != nil {
		// 参数解析失败，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 查询对应文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		// 检查文档是否不存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文档不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		}
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 如果请求中包含状态更新信息，则更新文档状态
	if request.Status != nil {
		document.Status = *request.Status
	}

	// 更新数据库中的文档信息
	if err := dao.UpdateDocument(document); err != nil {
		// 文档更新失败，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentStatusUpdateFailed)
		return
	}

	// 返回成功响应
	response.Success(c, nil, constant.DocumentStatusUpdateSuccess)
}
