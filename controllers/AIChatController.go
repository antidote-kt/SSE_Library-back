package controllers

import (
	"net/http"

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

	// 调用 StreamChat 函数处理流式响应，默认不启用思考内容推送
	if err := utils.StreamChat(c, req.Messages, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}
