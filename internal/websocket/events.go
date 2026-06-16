package websocket

type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

const (
	EventSOSCreated           = "SOS_CREATED"
	EventAssistanceCreated    = "ASSISTANCE_CREATED"
	EventAssistanceStatusUpdated = "ASSISTANCE_STATUS_UPDATED"
	EventReportCreated        = "REPORT_CREATED"
	EventAlertCreated         = "ALERT_CREATED"
	EventAlertApproved        = "ALERT_APPROVED"
	EventSOSStatusUpdated     = "SOS_STATUS_UPDATED"
	EventChatMessageCreated   = "CHAT_MESSAGE_CREATED"
	EventNotificationCreated  = "NOTIFICATION_CREATED"
)