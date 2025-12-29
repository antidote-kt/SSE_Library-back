package response

import (
	"github.com/antidote-kt/SSE_Library-back/models"
)

type NotificationResponse struct {
	Content    string `json:"content"`
	IsRead     bool   `json:"isRead"`
	ReceiverID uint64 `json:"receiverId"`
	ReminderID uint64 `json:"reminderId"`
	SendTime   string `json:"sendTime"`
	Type       string `json:"type"`       // comment, like, favorite
	SourceID   uint64 `json:"sourceId"`   // documentId 或 postId
	SourceType string `json:"sourceType"` // "document" 或 "post"
}

func BuildNotificationResponse(notification models.Notification) NotificationResponse {
	return NotificationResponse{
		Content:    notification.Content,
		IsRead:     notification.IsRead,
		ReceiverID: notification.ReceiverID,
		ReminderID: notification.ID,
		SendTime:   notification.CreatedAt.Format("2006-01-02 15:04:05"),
		Type:       notification.Type,
		SourceID:   notification.SourceID,
		SourceType: notification.SourceType,
	}
}

func BuildNotificationResponseList(notifications []models.Notification) []NotificationResponse {
	var responses []NotificationResponse
	for _, notification := range notifications {
		responses = append(responses, BuildNotificationResponse(notification))
	}
	return responses
}
