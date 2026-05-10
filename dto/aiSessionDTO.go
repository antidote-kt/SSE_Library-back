package dto

type CreateAISessionDTO struct {
	UserID uint64 `json:"userId" binding:"required"`
}

type UpdateAISessionDTO struct {
	NewTitle  string `json:"newTitle" binding:"required"`
	UserID    uint64 `json:"userId" binding:"required"`
	SessionID uint64 `json:"sessionId" binding:"required"`
}

type DeleteAISessionDTO struct {
	UserID uint64 `json:"userId" binding:"required"`
}

type SendMessageRequest struct {
	UserID  uint64 `json:"userId" binding:"required"`
	Content string `json:"question" binding:"required"`
	IsThink bool   `json:"isThink"`
}
