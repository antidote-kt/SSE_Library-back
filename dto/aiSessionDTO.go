package dto

type CreateAISessionDTO struct {
	UserID uint64 `json:"userId" binding:"required"`
}

type UpdateAISessionDTO struct {
	UserID uint64 `json:"userId" binding:"required"`
	ID     uint64 `json:"id" binding:"required"`
	Title  string `json:"title,omitempty"`
}
