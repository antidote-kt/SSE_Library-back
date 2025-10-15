package dto

type PostCommentDTO struct {
	Author     AuthorDTO `json:"author" binding:"required"`
	Content    string    `json:"content" binding:"required"`
	CreateTime string    `json:"createTime" binding:"required"`
	ParentID   *uint64   `json:"parent_id"`
}

type AuthorDTO struct {
	UserID     uint   `json:"userId" binding:"required"`
	UserName   string `json:"userName" binding:"required"`
	UserAvatar string `json:"userAvatar"`
}

type CommentResponseDTO struct {
	CommentID uint64           `json:"comment_id"`
	ParentID  *uint64          `json:"parent_id,omitempty"`
	Commenter UserBriefDTO     `json:"commenter"`
	Document  DocumentBriefDTO `json:"document"`
	CreatedAt string           `json:"created_at"`
	Content   string           `json:"content"`
}

type UserBriefDTO struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

type DocumentBriefDTO struct {
	Name        string `json:"name"`
	DocumentID  uint64 `json:"document_id"`
	Type        string `json:"type"`
	UploadTime  string `json:"uploadTime"`
	Status      string `json:"status"`
	Category    string `json:"category"`
	Course      string `json:"course"`
	Collections int    `json:"collections"`
	ReadCounts  int    `json:"readCounts"`
	URL         string `json:"URL"`
	Content     string `json:"content"`
	CreateTime  string `json:"createTime"`
}
