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
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != request.UserID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 查找对应文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}
	// 判断请求的用户是否是文档的拥有者
	if userClaims.UserID != document.UploaderID {
		response.Fail(c, http.StatusForbidden, nil, constant.NotAllowWithdrawOthers)
		return
	}
	if document.Status != constant.DocumentStatusAudit {
		response.Fail(c, http.StatusBadRequest, nil, constant.NotAllowWithdraw)
		return
	}
	document.Status = constant.DocumentStatusWithdraw
	// 查找对应分类
	_, err = dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
			return
		} else {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}

	if err := dao.UpdateDocument(document); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentUpdateFail)
		return
	}

	response.Success(c, nil, constant.WithdrawUploadSuccessMsg)

}

// GetUserUploadDocument 获取用户上传的文档列表
func GetUserUploadDocument(c *gin.Context) {
	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 将字符串转换为uint64
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	userClaims := claims.(*utils.MyClaims)

	if userClaims.UserID != userID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err = dao.GetUserByID(userClaims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 查找用户上传的所有文档
	documents, err := dao.GetDocumentsByUploaderID(userClaims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构造返回数据数组
	var responseData []response.InfoBriefResponse
	for _, doc := range documents {
		// 使用BuildInfoBriefResponse构建文档信息
		infoBrief, err := response.BuildInfoBriefResponse(doc)
		if err != nil {
			continue // 跳过构建失败的文档
		}
		responseData = append(responseData, infoBrief)
	}

	// 返回用户上传的文档列表
	response.SuccessWithData(c, responseData, constant.GetUserUploadSuccessMsg)
}
