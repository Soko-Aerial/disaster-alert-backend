package dto

import "disaster_alert_backend/internal/models"

// CreateReportRequest is used by a mobile user to submit a new incident report.
//
// The report will be saved as "pending" first.
// An admin must approve it before it becomes a public alert.
type CreateReportRequest struct {
	// Category describes the type of incident being reported.
	//
	// Allowed examples:
	// flood, fire, security, health, weather, accident, other
	Category string `json:"category" binding:"required" example:"flood" enums:"flood,fire,security,health,weather,accident,other"`

	// Description explains what the user saw or experienced.
	//
	// Write a clear short description of the incident.
	Description string `json:"description" binding:"required" example:"Flooding has started around the roadside and vehicles cannot pass."`

	// TimeOfOccurrence is when the incident happened.
	//
	// Use ISO date/time format where possible.
	// Example: 2026-09-10T08:30:00Z
	TimeOfOccurrence string `json:"timeOfOccurrence" binding:"required" example:"2026-09-10T08:30:00Z"`

	// Latitude is the GPS latitude of the incident location.
	Latitude float64 `json:"latitude" binding:"required" example:"5.6037"`

	// Longitude is the GPS longitude of the incident location.
	Longitude float64 `json:"longitude" binding:"required" example:"-0.1870"`

	// Address is the readable location, landmark, street, or area name.
	Address string `json:"address" example:"Circle, Accra"`

	// Country is the country where the incident happened.
	Country string `json:"country" example:"Ghana"`

	// Region is the region, state, or province where the incident happened.
	Region string `json:"region" example:"Greater Accra"`

	// MediaURLs contains already-uploaded image or video links.
	//
	// For normal JSON reports, this can be empty.
	// For uploaded files, the backend fills this after Cloudinary upload.
	MediaURLs []string `json:"mediaUrls,omitempty" example:"https://res.cloudinary.com/example/video/upload/report.mp4"`

	// Media contains detailed uploaded media information.
	//
	// For normal JSON reports, this can be empty.
	// For multipart upload, the backend fills URL, type, and Cloudinary publicId.
	Media []models.ReportMedia `json:"media,omitempty"`
}
