package dto

// SendMessageDTO 发送消息接口的请求参数
type SendMessageDTO struct {
	SessionID  *uint64 `form:"sessionId"`                  // 会话ID (可选)
	ReceiverID *uint64 `form:"receiverId"`                 // 接收者ID (可选，如果没有SessionID，则必须有此字段)
	Content    string  `form:"content" binding:"required"` // 消息内容(不能发送空消息)
}

// CreateChatSessionDTO 创建聊天会话接口的请求参数
type CreateChatSessionDTO struct {
	SenderID   *uint64 `json:"myId"`       // 发送者ID (必须)
	ReceiverID *uint64 `json:"oppositeId"` // 接收者ID (必须)
}

// MarkSessionRead 标记会话为已读接口的请求参数
type MarkSessionRead struct {
	SessionID *uint64 `form:"sessionId"` // 会话ID (必须)
}
