package response

import (
	"errors"
	"fmt"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
	"github.com/antidote-kt/SSE_Library-back/utils"
	"gorm.io/gorm"
)

// SearchChatResult 聊天记录搜索结果
type SearchChatResult struct {
	SessionID   uint64 `json:"sessionId"`
	User1ID     uint64 `json:"user1Id"`
	User1Avatar string `json:"user1Avatar"`
	User1Name   string `json:"user1Name"`
	User2ID     uint64 `json:"user2Id"`
	User2Avatar string `json:"user2Avatar"`
	User2Name   string `json:"user2Name"`
	MatchCount  uint64 `json:"matchedCount"`
	LatestMsg   string `json:"example"`
}

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

// BuildSessionResponse 构建单个会话响应
// （创建聊天接口直接调用以及获取会话列表接口复用，为了明确当前用户身份以统计会话未读消息数，这里还需要传入当前用户ID）
func BuildSessionResponse(session models.Session, currentUserID uint64) (SessionResponse, error) {
	// 1. 获取双方用户信息
	user1, err1 := dao.GetUserByID(session.User1ID)
	user2, err2 := dao.GetUserByID(session.User2ID)
	if err1 != nil || err2 != nil {
		return SessionResponse{}, fmt.Errorf(constant.GetSessionUserFailed)
	}

	// 2. 获取最后一条消息（新建聊天没有消息，因此要区分dao函数查不到记录和其他错误）
	// 这里GetLastMessageBySessionID内部使用的是First方法查找单条记录，如果没有找到会返回RecordNotFound错误
	lastMsg, errMsg := dao.GetLastMessageBySessionID(session.ID)
	lastContent := ""
	lastTime := ""
	if errMsg == nil {
		lastContent = lastMsg.Content
		lastTime = lastMsg.CreatedAt.Format("2006-01-02 15:04:05")
	} else if !errors.Is(errMsg, gorm.ErrRecordNotFound) {
		// 如果是 RecordNotFound 说明是刚创建的新聊天，还没有消息，这是正常的；其他错误则要响应
		return SessionResponse{}, fmt.Errorf(constant.GetChatMessageFailed)
	}

	// 3. 统计未读消息数 (我是接收者，所以统计别人发给我的未读)
	unreadCount, _ := dao.CountUnreadMessages(session.ID, currentUserID)

	// 4. 构建响应
	return SessionResponse{
		SessionID:   session.ID,
		UserID1:     session.User1ID,
		Avatar1:     utils.GetFileURL(user1.Avatar),
		Username1:   user1.Username,
		UserID2:     session.User2ID,
		Avatar2:     utils.GetFileURL(user2.Avatar),
		Username2:   user2.Username,
		LastMessage: lastContent,
		LastTime:    lastTime,
		UnreadCount: unreadCount,
	}, nil
}

// BuildSessionResponses 批量构建会话响应
func BuildSessionResponses(sessions []models.Session, currentUserID uint64) ([]SessionResponse, error) {
	var responses []SessionResponse
	for _, session := range sessions {
		response, err := BuildSessionResponse(session, currentUserID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, nil
}
