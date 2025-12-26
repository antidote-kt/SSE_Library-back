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

func GetNotification(c *gin.Context) {
	var userId uint64

	// 首先尝试从URL查询参数获取userId
	userIdStr := c.Query("userId")
	parsedUserId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.MsgUserIDFormatError)
		return
	}
	// 从上下文中获取用户JWT声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 如果无法获取用户信息，返回401未授权错误
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	// 将接口类型转换为具体的声明结构体
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的用户ID是否与JWT中的用户ID一致（防止越权操作）
	if userClaims.UserID != parsedUserId {
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 验证用户是否存在
	_, err = dao.GetUserByID(userClaims.UserID)
	if err != nil {
		// 如果用户不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}
	userId = userClaims.UserID

	// 调用dao层获取通知列表
	notifications, err := dao.GetNotificationsByUserId(userId)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构建响应
	notificationResponses := response.BuildNotificationResponseList(notifications)
	response.SuccessWithData(c, notificationResponses, constant.GetNotificationSuccess)
}

func MarkNotification(c *gin.Context) {
	// 声明标记通知已读请求参数结构体
	var request dto.MarkNotificationDTO

	// 解析JSON请求参数
	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果参数解析失败，返回400错误
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从上下文中获取用户JWT声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 如果无法获取用户信息，返回401未授权错误
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}
	// 将接口类型转换为具体的声明结构体
	userClaims := claims.(*utils.MyClaims)

	// 验证请求的通知ID是否有效
	notificationID := request.ReminderID
	if notificationID == 0 {
		response.Fail(c, http.StatusBadRequest, nil, constant.NotificationIDInvalid)
		return
	}

	// 验证用户是否存在
	_, err := dao.GetUserByID(userClaims.UserID)
	if err != nil {
		// 如果用户不存在，返回404错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.UserNotExist)
			return
		}
		// 其他数据库错误，返回500错误
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 在这里需要调用DAO层的函数来标记通知为已读
	// 验证通知是否存在且属于当前用户
	err = dao.MarkNotificationAsRead(notificationID, userClaims.UserID)
	if err != nil {
		// 如果更新失败，返回错误
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, nil, constant.NotificationNotExist)
			return
		}
		response.Fail(c, http.StatusInternalServerError, nil, constant.MarkReadFailed)
		return
	}

	// 构建响应
	response.Success(c, nil, constant.MarkReadSuccess)
}

func GetUnreadMessage(c *gin.Context) {
	// 声明请求参数结构体并解析参数
	var request dto.GetUnreadMessageDTO
	if err := c.ShouldBind(&request); err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	// 从上下文中获取用户JWT声明信息
	claims, exists := c.Get(constant.UserClaims)
	if !exists {
		// 如果无法获取用户信息，返回401未授权错误
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	// 将接口类型转换为具体的声明结构体
	userClaims := claims.(*utils.MyClaims)
	// 验证请求的用户ID是否为本人（这一步基本用不到，从始至终会直接用userClaims.UserID作为用户信息）
	if request.UserID != userClaims.UserID {
		response.Fail(c, http.StatusBadRequest, nil, constant.NonSelf)
		return
	}

	// 根据用户ID和消息类型查询未读消息数量
	count, err := dao.GetUnreadMessageCount(request.UserID, request.Type)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回成功响应
	response.SuccessWithData(c, count, constant.GetUnreadMessageSuccess)
}
