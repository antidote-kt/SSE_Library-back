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

// AdminModifyDocumentStatus 管理员修改文档状态控制器
// 该函数允许管理员修改文档的状态和名称
func AdminModifyDocumentStatus(c *gin.Context) {
	// 验证管理员身份
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	if userClaims.Role != "admin" {
		response.Fail(c, http.StatusForbidden, nil, constant.NoPermission)
		return
	}

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

// AdminGetDocumentList 管理员获取文档列表
// 返回所有未删除的文档详情列表（包括非open状态的文档）
func AdminGetDocumentList(c *gin.Context) {
	// 验证管理员身份
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)
	if userClaims.Role != "admin" {
		response.Fail(c, http.StatusForbidden, nil, constant.NoPermission)
		return
	}

	// 获取所有未删除的文档（管理员可以看到所有状态的文档）
	documents, err := dao.GetAllDocuments()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建文档详情响应列表
	var documentDetailResponses []response.DocumentDetailResponse
	for _, document := range documents {
		docDetailResponse, err := response.BuildDocumentDetailResponse(document)
		if err != nil {
			// 如果构建某个文档详情失败，记录错误但继续处理其他文档
			continue
		}
		documentDetailResponses = append(documentDetailResponses, docDetailResponse)
	}

	// 返回成功响应
	response.SuccessWithData(c, documentDetailResponses, constant.DocumentsObtain)
}
