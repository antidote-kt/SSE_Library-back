package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/response"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"github.com/gin-gonic/gin"
)

// AIChatRequest 表示 AI 聊天请求
type AIChatRequest struct {
	Messages []utils.Message `json:"messages"`
}

// TestStreamChat 测试 AI 聊天的 SSE 流式响应（直接构造请求示例）
func TestStreamChat(c *gin.Context) {
	// 直接构造请求示例
	req := AIChatRequest{
		Messages: []utils.Message{
			{
				Role:    "user",
				Content: "你好，我想了解一下图书馆的开放时间",
			},
		},
	}
	// 调用 StreamChatWithSessionID 函数处理流式响应，默认不启用思考内容推送
	// 使用临时 sessionId
	result, err := utils.StreamChatWithSessionID(c, "80000", req.Messages, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 虽然是流式响应，但可以在这里记录或处理完整内容
	fmt.Println("=== AI 完整回复内容 ===")
	if result != nil {
		fmt.Println("Content:", result.Content)
		fmt.Println("ThinkingContent:", result.ThinkingContent)
	} else {
		fmt.Println("result is nil")
	}
	fmt.Println("====================")
}

// TestGenerateTitle 测试会话标题生成
func TestGenerateTitle(c *gin.Context) {
	userInput := "你好，我想了解一下图书馆的开放时间"

	title, err := utils.GenerateSessionTitle(userInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"input": userInput,
		"title": title,
	})
}

// CancelAISessionStream 终止AI会话流式输出
func CancelAISessionStream(c *gin.Context) {
	sessionIdStr := c.Param("sessionId")
	if sessionIdStr == "" {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	_, exists := c.Get(constant.UserClaims)
	if !exists {
		response.Fail(c, http.StatusUnauthorized, nil, constant.GetUserInfoFailed)
		return
	}

	sessionId, err := strconv.ParseUint(sessionIdStr, 10, 64)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, nil, constant.ParamParseError)
		return
	}

	lastMessage, err := dao.GetLastAIMessageBySessionID(sessionId)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.CancelAISessionFailed)
		return
	}

	var messageId uint64 = 0
	var currentStatus string = ""
	var updatedMessage *models.AIMessage
	if lastMessage != nil {
		messageId = lastMessage.ID
		currentStatus = lastMessage.Status
		var err error
		updatedMessage, err = dao.UpdateAIMessageStatus(messageId, constant.AIMessageStatusInterrupted)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.CancelAISessionFailed)
			return
		}
	}

	if utils.CancelAISessionStreamTask(sessionIdStr) {
		var resp response.CancelAISessionResponse
		if updatedMessage != nil {
			resp = response.BuildCancelAISessionResponse(sessionId, updatedMessage.ID, updatedMessage.Status)
		} else {
			resp = response.BuildCancelAISessionResponse(sessionId, messageId, currentStatus)
		}
		response.SuccessWithData(c, resp, constant.CancelAISessionSuccess)
	} else {
		response.Fail(c, http.StatusNotFound, nil, constant.AISessionTaskNotExist)
	}
}
