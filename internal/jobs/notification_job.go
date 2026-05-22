package jobs

type NotificationTargetType string

const (
	TargetToken NotificationTargetType = "token"
	TargetUser  NotificationTargetType = "user"
	TargetAll   NotificationTargetType = "all"
)

type NotificationJob struct {
	TargetType NotificationTargetType
	UserID     string
	Token      string
	Title      string
	Body       string
	Data       map[string]string
}