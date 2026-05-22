package dto

type CreateReportRequest struct {
	Category         string   `json:"category" validate:"required"`
	Description      string   `json:"description" validate:"required,min=5"`
	TimeOfOccurrence string   `json:"timeOfOccurrence" validate:"required"`
	Latitude         float64  `json:"latitude" validate:"required"`
	Longitude        float64  `json:"longitude" validate:"required"`
	Address          string   `json:"address,omitempty"`
	MediaURLs        []string `json:"mediaUrls,omitempty"`
}