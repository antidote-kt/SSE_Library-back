package dto

// PostDocumentItemDTO 发帖时关联的文档信息
type PostDocumentItemDTO struct {
	DocumentID uint64 `json:"documentId" binding:"required"`
	Cover      string `json:"cover" binding:"required"`
}

// CreatePostDTO 发帖接口请求参数
type CreatePostDTO struct {
	SenderID     uint64                `json:"senderId" binding:"required"`
	SenderName   string                `json:"senderName" binding:"required"`
	SenderAvatar string                `json:"senderAvatar" binding:"required"`
	Title        string                `json:"title" binding:"required"`
	Content      string                `json:"content" binding:"required"`
	SendTime     string                `json:"sendTime" binding:"required"` // 使用前端传来的时间字符串作为 create_time
	Documents    []PostDocumentItemDTO `json:"documents"`                   // 可选的关联文档列表
}

// GetPostListDTO 获取帖子列表的查询参数
type GetPostListDTO struct {
	Key   string `form:"key"`   // 搜索关键词
	Order string `form:"order"` // 排序方式: time, hot
}
