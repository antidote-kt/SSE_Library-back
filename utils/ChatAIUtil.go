package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type RequestBody struct {
	Model          string    `json:"model"`
	Messages       []Message `json:"messages"`
	Stream         bool      `json:"stream"`
	EnableThinking bool      `json:"enable_thinking"`
}

type Delta struct {
	Content          string `json:"content"`
	ReasoningContent string `json:"reasoning_content"`
}

type ChunkResponse struct {
	Choices []struct {
		Delta        Delta  `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
}

type StreamResult struct {
	Content         string
	ThinkingContent string
}

// processStreamResponse 处理流式响应并将数据发送到通道，同时收集完整内容
func processStreamResponse(resp *http.Response, dataChan chan string, enableThinking bool, resultChan chan *StreamResult) {
	defer close(dataChan)
	defer close(resultChan)

	var contentBuilder strings.Builder
	var thinkingBuilder strings.Builder

	// 使用最小缓冲区，确保实时读取
	reader := bufio.NewReaderSize(resp.Body, 128) // 更小的缓冲区

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}

		// 跳过空行
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 去除 "data: " 前缀
		if strings.HasPrefix(line, "data: ") {
			line = strings.TrimPrefix(line, "data: ")
		}

		// 跳过结束标记
		if line == "[DONE]" {
			dataChan <- "[END_OF_STREAM]"
			break
		}

		// 直接解析 JSON
		var chunk ChunkResponse
		if err := json.Unmarshal([]byte(line), &chunk); err != nil {
			continue // 跳过解析失败的行
		}

		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			if content != "" {
				dataChan <- content
				contentBuilder.WriteString(content)
			}

			// 处理思考内容
			if enableThinking {
				reasoningContent := chunk.Choices[0].Delta.ReasoningContent
				if reasoningContent != "" {
					dataChan <- "[thinking] " + reasoningContent
					thinkingBuilder.WriteString(reasoningContent)
				}
			}

			if chunk.Choices[0].FinishReason == "stop" {
				// 发送结束标记
				dataChan <- "[END_OF_STREAM]"
				break
			}
		}
	}

	resultChan <- &StreamResult{
		Content:         contentBuilder.String(),
		ThinkingContent: thinkingBuilder.String(),
	}
}

// StreamChat 处理 SSE 流式响应
// enableThinking 是否启用思考内容推送
// 返回值: StreamResult（包含完整内容和思考内容）, error
func StreamChat(c *gin.Context, messages []Message, enableThinking bool) (*StreamResult, error) {
	// 从配置中读取参数
	model := viper.GetString("dashscope.model")
	if model == "" {
		model = "qwen-plus"
	}
	systemPrompt := constant.AIChatSystemPrompt

	apiKey := viper.GetString("dashscope.api_key")
	if apiKey == "" {
		return nil, fmt.Errorf("api key not found")
	}
	endpoint := viper.GetString("dashscope.endpoint")
	if endpoint == "" {
		endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	}

	// 构建请求体
	requestBody := RequestBody{
		Model: model,
		Messages: append([]Message{{
			Role:    "system",
			Content: systemPrompt,
		}}, messages...),
		Stream:         true,
		EnableThinking: enableThinking,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	// 创建并发送请求
	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 使用默认的HTTP客户端
	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 创建无缓冲通道，确保实时传递
	dataChan := make(chan string)
	resultChan := make(chan *StreamResult)

	// 启动协程处理流式响应
	go processStreamResponse(resp, dataChan, enableThinking, resultChan)

	//  Gin Stream 实时推送版本
	c.Stream(func(w io.Writer) bool {
		msg, ok := <-dataChan
		if !ok {
			return false
		}

		if msg == "[END_OF_STREAM]" {
			// 发送结束事件
			c.SSEvent("end", "DONE")
			return false
		}

		// 发送消息
		if strings.HasPrefix(msg, "[thinking] ") {
			// 发送思考内容
			thinkingContent := strings.TrimPrefix(msg, "[thinking] ")
			c.SSEvent("thinking", thinkingContent)
		} else {
			// 发送普通内容
			c.SSEvent("message", msg)
		}

		return true
	})

	// 获取完整结果
	streamResult := <-resultChan
	return streamResult, nil
}

// Chat 非流式调用AI模型，返回完整内容
func Chat(messages []Message) (string, error) {
	model := viper.GetString("dashscope.model")
	if model == "" {
		model = "qwen-plus"
	}

	apiKey := viper.GetString("dashscope.api_key")
	if apiKey == "" {
		return "", fmt.Errorf("api key not found")
	}
	endpoint := viper.GetString("dashscope.endpoint")
	if endpoint == "" {
		endpoint = "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	}

	requestBody := RequestBody{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) > 0 {
		return chatResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from model")
}

// GenerateSessionTitle 根据用户输入生成会话标题
func GenerateSessionTitle(userInput string) (string, error) {
	messages := []Message{
		{
			Role:    "system",
			Content: constant.AISessionTitlePrompt,
		},
		{
			Role:    "user",
			Content: userInput,
		},
	}

	return Chat(messages)
}
