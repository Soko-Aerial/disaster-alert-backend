package dto

type CreateAlertRequest struct {
	Title              string   `json:"title" validate:"required"`
	Description        string   `json:"description" validate:"required,min=5"`
	Category           string   `json:"category" validate:"required"`
	Severity           string   `json:"severity" validate:"required,oneof=low medium high critical"`
	Status             string   `json:"status,omitempty" validate:"omitempty,oneof=draft active resolved expired cancelled"`
	Latitude           float64  `json:"latitude" validate:"required"`
	Longitude          float64  `json:"longitude" validate:"required"`
	Address            string   `json:"address,omitempty"`
	Country            string   `json:"country,omitempty"`
	Region             string   `json:"region,omitempty"`
	RadiusKm           float64  `json:"radiusKm,omitempty"`
	SafetyInstructions []string `json:"safetyInstructions,omitempty"`
	SourceType 		   string   `json:"sourceType,omitempty" validate:"omitempty,oneof=internal external system user_report weather health news admin"`
	SourceName         string   `json:"sourceName,omitempty"`
	ExternalID         string   `json:"externalId,omitempty"`
	SourceURL          string   `json:"sourceUrl,omitempty"`
	EventTime          string   `json:"eventTime,omitempty"`
	ExpiresAt          string   `json:"expiresAt,omitempty"`
	Confidence         float64  `json:"confidence,omitempty"`
}

type UpdateAlertStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active resolved expired cancelled"`
}