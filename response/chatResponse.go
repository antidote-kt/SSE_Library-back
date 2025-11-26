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

// SessionResponse 会话列表响应(单个会话)结构
// 包含session数据模型的相关信息（ID、user1、user2）以及每个会话具体信息（avatar、username、LastMessage、LastTime、UnreadCount）
// 因为接口返回数据不仅仅是models.session包含的信息，还有其他特定需要的信息，因此要额外为响应数据定义结构体
type SessionResponse struct {
	SessionID   uint64 `json:"sessionId"`
	UserID1     uint64 `json:"userId1"`
	Avatar1     string `json:"avatar1"`
	Username1   string `json:"username1"`
	UserID2     uint64 `json:"userId2"`
	Avatar2     string `json:"avatar2"`
	Username2   string `json:"username2"`
	LastMessage string `json:"lastMessage"` // 最后一条消息内容
	LastTime    string `json:"lastTime"`    // 最后一条消息时间
	UnreadCount uint64 `json:"unreadCount"` // 未读消息数
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
