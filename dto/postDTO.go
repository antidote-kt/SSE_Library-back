package dto

// CreatePostDTO 发帖接口请求参数
type CreatePostDTO struct {
	SenderID    uint64   `json:"senderId" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Content     string   `json:"content" binding:"required"`
	DocumentIDs []uint64 `json:"documents"` // 可选的关联文档ID列表
}

// GetPostListDTO 获取帖子列表的查询参数
type GetPostListDTO struct {
	Key   string `form:"key"`   // 搜索关键词
	Order string `form:"order"` // 排序方式: time, hot
}

// LikePostDTO 点赞帖子请求参数
type LikePostDTO struct {
	PostID uint64 `form:"postId" binding:"required"`
	UserID uint64 `form:"userId" binding:"required"`
}
