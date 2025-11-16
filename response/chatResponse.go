package response

import (
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/utils"
)

type ChatRecordResponse struct {
	Content      string `json:"content"`
	SenderAvatar string `json:"senderAvatar"`
	SenderID     uint64 `json:"senderId"`
	SenderName   string `json:"senderName"`
	SendTime     string `json:"sendTime"`
	SessionID    uint64 `json:"sessionId"`
	Status       string `json:"status"`
}

func BuildChatRecordResponse(message models.Message) (ChatRecordResponse, error) {
	// 获取发送者用户信息
	sender, err := dao.GetUserByID(message.SenderID)
	if err != nil {
		return ChatRecordResponse{}, fmt.Errorf(constant.GetSenderFailed)
	}

	return ChatRecordResponse{
		Content:      message.Content,
		SenderAvatar: utils.GetFileURL(sender.Avatar),
		SenderID:     message.SenderID,
		SenderName:   sender.Username,
		SendTime:     message.CreatedAt.Format("2006-01-02 15:04:05"), // 格式化时间
		SessionID:    message.SessionID,
		Status:       message.Status,
	}, nil
}

// BuildChatRecordResponses 批量构建聊天记录响应
func BuildChatRecordResponses(messages []models.Message) ([]ChatRecordResponse, error) {
	var responses []ChatRecordResponse
	for _, message := range messages {
		response, err := BuildChatRecordResponse(message)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}
