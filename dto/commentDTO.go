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
