package dto

type SaveFCMTokenRequest struct {
	Token      string `json:"token" validate:"required"`
	DeviceType string `json:"deviceType,omitempty"`
}

type SendTestNotificationRequest struct {
	Title string `json:"title" validate:"required"`
	Body  string `json:"body" validate:"required"`
}