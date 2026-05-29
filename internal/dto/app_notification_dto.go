package dto

type CreateAppNotificationRequest struct {
	RecipientID   string            `json:"recipientId" binding:"required"`
	RecipientRole string            `json:"recipientRole"`
	Title         string            `json:"title" binding:"required"`
	Body          string            `json:"body" binding:"required"`
	Type          string            `json:"type" binding:"required"`
	ReferenceID   string            `json:"referenceId"`
	Data          map[string]string `json:"data"`
}

type MarkAppNotificationReadRequest struct {
	NotificationIDs []string `json:"notificationIds"`
}