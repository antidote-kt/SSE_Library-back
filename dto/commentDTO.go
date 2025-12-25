package dto

type PostCommentDTO struct {
	ParentID   *uint64        `json:"parentId"`
	Commenter  UserBriefDTO   `json:"commenter" binding:"required"`
	SourceData *SourceDataDTO `json:"sourceData" binding:"required"`
	Content    string         `json:"content" binding:"required"`
}

type AuthorDTO struct {
	UserID     uint   `json:"userId" binding:"required"`
	UserName   string `json:"userName" binding:"required"`
	UserAvatar string `json:"userAvatar"`
}

// SourceDataDTO 定义了所属位置的信息
type SourceDataDTO struct {
	SourceID   uint64 `json:"sourceId"`
	Name       string `json:"name"`
	SourceType string `json:"sourceType"` // document 或 post
}

type CommentResponseDTO struct {
	CommentID  uint64         `json:"commentId"`
	ParentID   *uint64        `json:"parentId"`
	SourceData *SourceDataDTO `json:"sourceData"`
	Commenter  UserBriefDTO   `json:"commenter"`
	CreatedAt  string         `json:"createdAt"`
	Content    string         `json:"content"`
}
