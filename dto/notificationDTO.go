package dto

type MarkNotificationDTO struct {
	ReminderID uint64 `json:"reminderId"`
}

type GetUnreadMessageDTO struct {
	UserID uint64 `form:"id"`
	Type   string `form:"Type"`
}
