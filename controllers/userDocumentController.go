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

// WithdrawUpload 用户撤回上传文档控制器
// 该函数允许用户撤回自己上传的文档（仅限状态为审核中的文档）
func WithdrawUpload(c *gin.Context) {
	// 初始化撤回上传请求结构体
	var request dto.WithdrawUploadDTO

	// 从查询参数绑定请求数据
	if err := c.ShouldBindQuery(&request); err != nil {
		// 参数解析失败，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从上下文中获取用户声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 获取用户信息失败，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	// 类型转换用户声明信息
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与当前登录用户ID一致
	if userClaims.UserID != request.UserID {
		// 不是本人操作，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 查找对应文档
	document, err := dao.GetDocumentByID(request.DocumentID)
	if err != nil {
		// 检查文档是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 文档不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.DocumentNotExist)
			return
		} else {
			// 数据库操作错误，返回错误响应
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}

	// 判断请求的用户是否是文档的拥有者
	if userClaims.UserID != document.UploaderID {
		// 不允许撤回他人上传的文档，返回错误响应
		response.Fail(c, http.StatusForbidden, nil, constant.NotAllowWithdrawOthers)
		return
	}

	//// 检查文档状态是否为审核中（只有审核中的文档才能撤回）
	//if document.Status != constant.DocumentStatusPending {
	//	// 不允许撤回，返回错误响应
	//	response.Fail(c, http.StatusBadRequest, nil, constant.NotAllowWithdraw)
	//	return
	//}

	// 更新文档状态为撤回状态
	document.Status = constant.DocumentStatusWithdrawn

	// 查找对应分类（验证分类是否存在）
	_, err = dao.GetCategoryByID(document.CategoryID)
	if err != nil {
		// 检查分类是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 分类不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.CategoryNotExist)
			return
		} else {
			// 数据库操作错误，返回错误响应
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
	}

	// 更新数据库中的文档信息
	if err := dao.UpdateDocument(document); err != nil {
		// 文档更新失败，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DocumentUpdateFail)
		return
	}

	// 返回撤回上传成功消息
	response.Success(c, nil, constant.WithdrawUploadSuccessMsg)
}

// GetUserUploadDocument 获取用户上传的文档列表控制器
// 该函数返回指定用户上传的所有文档的简要信息列表
func GetUserUploadDocument(c *gin.Context) {
	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 将字符串转换为uint64类型的用户ID
	userID, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		// 用户ID格式错误，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}

	// 从上下文中获取用户声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 获取用户信息失败，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	// 类型转换用户声明信息
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与当前登录用户ID一致
	if userClaims.UserID != userID {
		// 不是本人操作，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err = dao.GetUserByID(userClaims.UserID)
	if err != nil {
		// 检查用户是否存在
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，返回错误响应
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 查找用户上传的所有文档
	documents, err := dao.GetDocumentsByUploaderID(userClaims.UserID)
	if err != nil {
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构造返回数据数组，存储文档简要信息
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
