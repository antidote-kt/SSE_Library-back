package dto

// PostCommentDTO 发表评论请求DTO
type PostCommentDTO struct {
	Author     AuthorDTO `json:"author" binding:"required"`     // 作者信息
	Content    string    `json:"content" binding:"required"`    // 评论内容
	CreateTime string    `json:"createTime" binding:"required"` // 创建时间
	ParentID   *uint64   `json:"parent_id"`                     // 父评论ID（可选，为空则是主评论，有值则是回复）
}

// AuthorDTO 作者信息DTO
type AuthorDTO struct {
	UserID     uint   `json:"userId" binding:"required"`   // 用户ID
	UserName   string `json:"userName" binding:"required"` // 用户名
	UserAvatar string `json:"userAvatar"`                  // 用户头像（可选）
}

// CommentResponseDTO 评论响应DTO
type CommentResponseDTO struct {
	CommentID uint64           `json:"comment_id"`
	ParentID  *uint64          `json:"parent_id,omitempty"` // 父评论ID（null表示主评论）
	Commenter UserBriefDTO     `json:"commenter"`
	Book      DocumentBriefDTO `json:"book"`
	CreatedAt string           `json:"created_at"`
	Content   string           `json:"content"`
}

// UserBriefDTO 用户简要信息DTO
type UserBriefDTO struct {
	UserID     uint64 `json:"userId"`
	Username   string `json:"username"`
	UserAvatar string `json:"userAvatar"`
	Status     string `json:"status"`
	CreateTime string `json:"createTime"`
	Email      string `json:"email"`
	Role       string `json:"role"`
}

// DocumentBriefDTO 文档简要信息DTO
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
