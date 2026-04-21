package response

import (
	"github.com/antidote-kt/SSE_Library-back/dao"
	"github.com/antidote-kt/SSE_Library-back/models"
)

type CreateAISessionResponse struct {
	AISessionID uint64 `json:"aiSessionId"`
	CreateTime  string `json:"createTime"`
}

type AISessionListItemResponse struct {
	AISessionID   uint64  `json:"aiSessionId"`
	UserID        uint64  `json:"userId"`
	AISessionName string  `json:"aiSessionName"`
	LastTime      *string `json:"lasttime"`
}

func BuildCreateAISessionResponse(session models.AISession) CreateAISessionResponse {
	return CreateAISessionResponse{
		AISessionID: session.ID,
		CreateTime:  session.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func BuildAISessionListItemResponse(session models.AISession) AISessionListItemResponse {
	var lastTime *string

	lastMessage, _ := dao.GetLastAIMessageBySessionID(session.ID)
	if lastMessage != nil {
		timeStr := lastMessage.CreatedAt.Format("2006-01-02 15:04:05")
		lastTime = &timeStr
	}

	return AISessionListItemResponse{
		AISessionID:   session.ID,
		UserID:        session.UserID,
		AISessionName: session.Title,
		LastTime:      lastTime,
	}
}

func BuildAISessionListResponses(sessions []models.AISession) []AISessionListItemResponse {
	responses := make([]AISessionListItemResponse, len(sessions))
	for i, session := range sessions {
		responses[i] = BuildAISessionListItemResponse(session)
	}
	return responses
}
