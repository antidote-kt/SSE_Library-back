package controllers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/dto"
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

	// RAG知识增强检索
	// 1. 提取用户最新的问题
	userQuery := req.Messages[len(req.Messages)-1].Content

	// 2. 将用户问题向量化
	queryVec, err := utils.GetEmbeddings([]string{userQuery})
	if err == nil && len(queryVec) > 0 {

		// 3. 去 Milvus 中检索最相关的 3 段内容
		relatedTexts, err := utils.SearchKnowledge(queryVec[0], 3)
		if err != nil {
			log.Printf("RAG检索出错: %v", err)
		}

		// 4. 将检索到的内容拼接到 Prompt 中
		if len(relatedTexts) > 0 {
			contextStr := ""
			for i, txt := range relatedTexts {
				contextStr += fmt.Sprintf("\n[片段%d]: %s", i+1, txt)
			}

			augmentedPrompt := fmt.Sprintf(`请结合以下参考资料回答我的问题。如果参考资料中没有相关信息，请明确告知，不要编造。
			参考资料：%s
			我的问题：%s`, contextStr, userQuery)

			// 替换最新一条消息的内容为增强后的 Prompt
			req.Messages[len(req.Messages)-1].Content = augmentedPrompt
		}
	}

	// 调用 StreamChatWithSessionID 函数处理流式响应，默认不启用思考内容推送
	// 使用临时 sessionId : 80000
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
	// 持久化 AI 的回复
	aiMsg := &models.AIMessage{
		AISessionsID: 80000,
		Role:         "assistant",
		Content:      result.Content,
	}
	err = dao.CreateAIMessage(aiMsg)
	if err != nil {
		fmt.Printf("保存 AI 回复失败: %v\n", err)
		return
	}
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

// SendAISessionMessages 发送问题并获取流式输出
func SendAISessionMessages(c *gin.Context) {
	// 1. 获取 URL 路径中的 sessionId
	sessionIdStr := c.Param("sessionId")
	sessionId, err := strconv.ParseUint(sessionIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "无效的会话ID"})
		return
	}

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "消息内容不能为空"})
		return
	}

	// 2. 验证用户身份一致性
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
	if userClaims.UserID != req.UserID {
		// 不是本人操作，返回错误响应
		response.Fail(c, http.StatusUnauthorized, nil, constant.NonSelf)
		return
	}

	// 补充逻辑
	// 在将当前消息存入数据库之前，查询该会话是否已经有历史消息，若没有说明是第一条信息，自动根据用户输入智能更新标题
	isFirstMessage := false
	count, _ := dao.GetMessageCountBySessionId(uint(sessionId))
	if count == 0 {
		isFirstMessage = true
	}
	// 补充逻辑：如果是当前对话的第一条消息，则根据用户输入智能修改标题
	if isFirstMessage {
		session, err := dao.GetAISessionByID(sessionId)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
			return
		}
		if session.Title == "新对话" {
			// 如果标题默认，则使用智能生成
			title, err := utils.GenerateSessionTitle(req.Content)
			if err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.GenerateSessionTitleFailed)
				return
			}
			session.Title = title
			if err := dao.UpdateAISession(&session); err != nil {
				response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
				return
			}
		}
	}

	// 3. 将用户发送的消息持久化到数据库
	userMsg := &models.AIMessage{
		AISessionsID: sessionId,
		Role:         "user",
		Content:      req.Content,
	}
	if err := dao.CreateAIMessage(userMsg); err != nil {
		response.Fail(c, http.StatusInternalServerError, nil, constant.DatabaseError)
		return
	}

	// 4. 构建大模型上下文 (获取所有历史对话，用户问题和ai回答交替排列)
	var messages []utils.Message
	// 传入 -1 取消 Limit 限制，获取会话的全部历史记录
	historyList, _ := dao.GetMessagesBySessionId(uint(sessionId), -1)

	for _, msg := range historyList {
		// 跳过刚刚插入的最新用户提问，因为稍后要单独用 RAG 增强
		if msg.ID == userMsg.ID {
			continue
		}

		if len(messages) == 0 {
			// 规范：确保上下文的第一条是用户提问
			if msg.Role == "user" {
				messages = append(messages, utils.Message{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		} else {
			lastIdx := len(messages) - 1
			if messages[lastIdx].Role == msg.Role {
				// 遇到连续相同角色（例如用户连发两句），将其合并为一条，保证严格交替
				messages[lastIdx].Content += "\n" + msg.Content
			} else {
				// 角色交替，正常追加
				messages = append(messages, utils.Message{
					Role:    msg.Role,
					Content: msg.Content,
				})
			}
		}
	}

	// 5. 知识库检索增强
	enhancedContent := req.Content // 默认使用原问题

	// 对用户当前问题进行向量化
	queryVec, err := utils.GetEmbeddings([]string{req.Content})
	if err == nil && len(queryVec) > 0 {
		// 在 Milvus 中检索 Top-3 的知识片段
		relatedChunks, searchErr := utils.SearchKnowledge(queryVec[0], 3)
		if searchErr == nil && len(relatedChunks) > 0 {
			// 将检索到的片段拼接
			contextInfo := strings.Join(relatedChunks, "\n\n---\n\n")

			// 替换为增强型 Prompt
			enhancedContent = fmt.Sprintf(
				"你是一个智能图书助手。请根据以下[已知知识库信息]回答用户的[问题]。\n如果已知信息中没有相关内容，请明确告知，不要自行编造。\n\n[已知知识库信息]：\n%s\n\n[用户问题]：%s",
				contextInfo,
				req.Content,
			)
		}
	} else {
		fmt.Printf("[RAG Warning] 问题向量化失败: %v\n", err)
	}

	// 6. 将增强后的问题作为最后一条 message 发给 AI
	// 保证最后加入的 enhancedContent (Role: user) 不会破坏交替排列
	if len(messages) > 0 && messages[len(messages)-1].Role == "user" {
		// 如果历史记录的最后一条也是 user (比如上一次 AI 网络超时没回复)，将其与本次提问合并
		messages[len(messages)-1].Content += "\n\n" + enhancedContent
	} else {
		// 正常交替，追加作为最后一条消息
		messages = append(messages, utils.Message{
			Role:    "user",
			Content: enhancedContent,
		})
	}

	// 7. 调用流式推流工具
	streamResult, err := utils.StreamChatWithSessionID(c, sessionIdStr, messages, req.IsThink)
	if err != nil {
		fmt.Printf("流式回复推送失败: %v\n", err)
		// 如果推送失败，也建议存入一条错误提示到数据库，保证会话连贯性
	}

	// 8. 保存 AI 完整回复
	aiMsg := &models.AIMessage{
		AISessionsID: sessionId,
		Role:         "assistant",
		Content:      streamResult.Content,
	}
	// 持久化 AI 的回复
	err = dao.CreateAIMessage(aiMsg)
	if err != nil {
		fmt.Printf("保存 AI 回复失败: %v\n", err)
		return
	}
}
