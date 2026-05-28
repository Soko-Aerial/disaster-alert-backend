package dto

import "disaster_alert_backend/internal/models"

type CreateReportRequest struct {
	Category         string `json:"category" binding:"required"`
	Description      string `json:"description" binding:"required"`
	TimeOfOccurrence string `json:"timeOfOccurrence" binding:"required"`

	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Address   string  `json:"address"`
	Country string `json:"country"`
	Region  string `json:"region"`

	MediaURLs []string             `json:"mediaUrls,omitempty"`
	Media     []models.ReportMedia `json:"media,omitempty"`
}