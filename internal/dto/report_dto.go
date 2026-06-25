package dto

import "disaster_alert_backend/internal/models"

type CreateReportRequest struct {
	Category string `json:"category" binding:"required" example:"flood" enums:"flood,fire,security,health,weather,accident,other"`

	Description string `json:"description" binding:"required" example:"Flooding has started around the roadside and vehicles cannot pass."`

	TimeOfOccurrence string `json:"timeOfOccurrence" binding:"required" example:"2026-06-25T10:30:00Z"`

	Latitude float64 `json:"latitude" binding:"required" example:"5.6037"`

	Longitude float64 `json:"longitude" binding:"required" example:"-0.1870"`

	Address string `json:"address" example:"Circle, Accra"`

	Country string `json:"country" example:"Ghana"`

	Region string `json:"region" example:"Greater Accra"`

	MediaURLs []string `json:"mediaUrls,omitempty" example:"https://res.cloudinary.com/example/image/upload/report.jpg"`

	Media []models.ReportMedia `json:"media,omitempty"`
}