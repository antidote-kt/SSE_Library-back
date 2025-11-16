package controllers

import (
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// GetChatMessages 获取聊天记录接口
func GetChatMessages(c *gin.Context) {
	// 从查询参数获取sessionId
	sessionIdStr := c.Query("sessionId")
	if sessionIdStr == "" {
		// 会话ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SessionIDLack)
		return
	}

	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 将字符串转换为uint64类型的会话ID
	sessionID, err := strconv.ParseUint(sessionIdStr, 10, 64)
	if err != nil {
		// 会话ID格式错误，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SessionIDFormatError)
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

	// 验证用户是否是会话的参与者
	session, err := dao.GetSessionByID(sessionID)
	if err != nil {
		// 如果会话不存在，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 验证当前用户是会话的参与者（User1ID或User2ID）
	if session.User1ID != userID && session.User2ID != userID {
		response.Fail(c, http.StatusUnauthorized, nil, constant.UserNotInSession)
		return
	}

	// 查询指定会话中的所有聊天记录
	messages, err := dao.GetAllMessagesBySession(sessionID)
	if err != nil {
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构造返回数据数组，存储聊天记录信息
	responseData, err := response.BuildChatRecordResponses(messages)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回用户的聊天记录列表
	response.SuccessWithData(c, responseData, constant.GetChatMessageSuccess)
}

// SearchChatMessages 搜索聊天记录接口
func SearchChatMessages(c *gin.Context) {
	// 从查询参数获取userId
	userIdStr := c.Query("userId")
	if userIdStr == "" {
		// 用户ID缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.UserIDLack)
		return
	}

	// 从查询参数获取searchKey
	searchKey := c.Query("searchKey")
	if searchKey == "" {
		// 搜索关键词缺失，返回错误响应
		response.Fail(c, http.StatusBadRequest, nil, constant.SearchKeyLack)
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

	// 搜索指定用户的聊天记录
	messages, err := dao.SearchChatMessagesByUser(userID, searchKey)
	if err != nil {
		// 数据库操作错误，返回错误响应
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 构造返回数据数组，存储聊天记录信息
	responseData, err := response.BuildChatRecordResponses(messages)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, err.Error())
		return
	}

	// 返回搜索到的聊天记录列表
	response.SuccessWithData(c, responseData, constant.GetChatMessageSuccess)
}
