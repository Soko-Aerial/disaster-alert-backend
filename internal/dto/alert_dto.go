package dto

type CreateAlertRequest struct {
	Title string `json:"title" validate:"required" example:"Heavy rainfall warning"`

	Description string `json:"description" validate:"required,min=5" example:"Heavy rainfall is expected in Accra with possible flooding in low-lying areas."`

	Category string `json:"category" validate:"required" example:"flood" enums:"flood,fire,weather,health,security,earthquake,conflict,other"`

	Severity string `json:"severity" validate:"required,oneof=low medium high critical" example:"high" enums:"low,medium,high,critical"`

	Status string `json:"status,omitempty" validate:"omitempty,oneof=draft active resolved expired cancelled" example:"active" enums:"draft,active,resolved,expired,cancelled"`

	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	Address string `json:"address,omitempty" example:"Circle, Accra"`

	Country string `json:"country,omitempty" example:"Ghana"`

	Region string `json:"region,omitempty" example:"Greater Accra"`

	RadiusKm float64 `json:"radiusKm,omitempty" example:"10"`

	SafetyInstructions []string `json:"safetyInstructions,omitempty" example:"Move to higher ground"`

	SourceType string `json:"sourceType,omitempty" validate:"omitempty,oneof=internal external system user_report weather health news admin" example:"admin" enums:"internal,external,system,user_report,weather,health,news,admin"`

	SourceName string `json:"sourceName,omitempty" example:"Admin Dashboard"`

	ExternalID string `json:"externalId,omitempty" example:"GDACS-12345"`

	SourceURL string `json:"sourceUrl,omitempty" example:"https://example.com/source-alert"`

	EventTime string `json:"eventTime,omitempty" example:"2026-06-25T10:00:00Z"`

	ExpiresAt string `json:"expiresAt,omitempty" example:"2026-06-26T18:00:00Z"`

	Confidence float64 `json:"confidence,omitempty" example:"0.95"`
}

type UpdateAlertStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active resolved expired cancelled" example:"resolved" enums:"draft,active,resolved,expired,cancelled"`
}