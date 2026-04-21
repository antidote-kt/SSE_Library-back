package controllers

import (
	"fmt"
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
	result, err := utils.StreamChat(c, req.Messages, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 虽然是流式响应，但可以在这里记录或处理完整内容
	fmt.Println("=== AI 完整回复内容 ===")
	fmt.Println("Content:", result.Content)
	fmt.Println("ThinkingContent:", result.ThinkingContent)
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
