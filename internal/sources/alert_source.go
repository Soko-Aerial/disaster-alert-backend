package sources

import "disaster_alert_backend/internal/models"

type AlertSource interface {
	Name() string
	FetchAlerts() ([]models.Alert, error)
}