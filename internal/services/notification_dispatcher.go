package services

import "disaster_alert_backend/internal/jobs"

type NotificationDispatcher interface {
	Dispatch(job jobs.NotificationJob) bool
}