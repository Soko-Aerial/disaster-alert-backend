package dto

// CreateAlertRequest is used by an admin to create a manual public alert.
//
// This endpoint is for verified/admin-created alerts.
// User-submitted reports should normally go through the report approval flow instead.
type CreateAlertRequest struct {
	// Title is the short headline users will see first.
	Title string `json:"title" validate:"required" example:"Heavy rainfall warning"`

	// Description gives the full details of the alert.
	// Minimum length: 5 characters.
	Description string `json:"description" validate:"required,min=5" example:"Heavy rainfall is expected in Accra with possible flooding in low-lying areas."`

	// Summary is a shorter version of the description.
	// It is useful for cards, notifications, and alert previews.
	Summary string `json:"summary,omitempty" example:"Heavy rainfall expected in Accra with possible flooding."`

	// Category describes the type of alert.
	//
	// Allowed values:
	// flood, fire, weather, health, security, earthquake, conflict, other
	Category string `json:"category" validate:"required" example:"flood" enums:"flood,fire,weather,health,security,earthquake,conflict,other"`

	// Severity describes how serious the alert is.
	//
	// Allowed values:
	// low, medium, high, critical
	Severity string `json:"severity" validate:"required,oneof=low medium high critical" example:"high" enums:"low,medium,high,critical"`

	// Status controls whether the alert is visible/active.
	//
	// If omitted, your service may default it to active or draft depending on your service logic.
	Status string `json:"status,omitempty" validate:"omitempty,oneof=draft active resolved expired cancelled" example:"active" enums:"draft,active,resolved,expired,cancelled"`

	// Latitude is the GPS latitude of the alert location.
	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	// Longitude is the GPS longitude of the alert location.
	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	// Address is a readable location, landmark, town, street, or area name.
	Address string `json:"address,omitempty" example:"Circle, Accra"`

	// Country is the affected country.
	Country string `json:"country,omitempty" example:"Ghana"`

	// Region is the affected region, state, province, or administrative area.
	Region string `json:"region,omitempty" example:"Greater Accra"`

	// RadiusKm is the approximate affected area around the latitude/longitude.
	//
	// Example: 10 means users within about 10km may be affected.
	RadiusKm float64 `json:"radiusKm,omitempty" example:"10"`

	// SafetyInstructions are clear actions users should take.
	SafetyInstructions []string `json:"safetyInstructions,omitempty" example:"Move to higher ground,Follow official instructions,Avoid flooded roads"`

	// SourceType identifies where the alert came from.
	//
	// Allowed values:
	// internal, external, system, user_report, weather, health, news, admin
	SourceType string `json:"sourceType,omitempty" validate:"omitempty,oneof=internal external system user_report weather health news admin" example:"admin" enums:"internal,external,system,user_report,weather,health,news,admin"`

	// SourceName is the human-readable source of the alert.
	SourceName string `json:"sourceName,omitempty" example:"Admin Dashboard"`

	// ExternalID is used when the alert comes from an external source such as GDACS, ReliefWeb, NASA FIRMS, or GDELT.
	ExternalID string `json:"externalId,omitempty" example:"GDACS-12345"`

	// SourceURL is an optional link to the original source.
	SourceURL string `json:"sourceUrl,omitempty" example:"https://example.com/source-alert"`

	// ImageURLs contains image evidence or images related to the alert.
	ImageURLs []string `json:"imageUrls,omitempty" example:"https://example.com/flood-image.jpg"`

	// VideoURLs contains video evidence or videos related to the alert.
	VideoURLs []string `json:"videoUrls,omitempty" example:"https://example.com/flood-video.mp4"`

	// Tags help classify and search alerts.
	Tags []string `json:"tags,omitempty" example:"flood,ghana,accra"`

	// PriorityScore is a numeric urgency score.
	//
	// Example:
	// 85 means serious or breaking depending on your service logic.
	PriorityScore int `json:"priorityScore,omitempty" example:"85"`

	// PriorityLabel is a readable priority level.
	//
	// Allowed values:
	// breaking, serious, watch, low
	PriorityLabel string `json:"priorityLabel,omitempty" validate:"omitempty,oneof=breaking serious watch low" example:"serious" enums:"breaking,serious,watch,low"`

	// EventTime is when the emergency happened or is expected to happen.
	//
	// Recommended format:
	// 2026-09-10T08:30:00Z
	EventTime string `json:"eventTime,omitempty" example:"2026-09-10T08:30:00Z"`

	// ExpiresAt is when the alert should stop being treated as active.
	//
	// Recommended format:
	// 2026-09-11T18:00:00Z
	ExpiresAt string `json:"expiresAt,omitempty" example:"2026-09-11T18:00:00Z"`

	// Confidence is how reliable the alert is.
	//
	// Example:
	// 0.95 means 95% confidence.
	Confidence float64 `json:"confidence,omitempty" example:"0.95"`
}

// UpdateAlertStatusRequest is used by an admin to change an alert status.
type UpdateAlertStatusRequest struct {
	// Status is the new alert status.
	//
	// Allowed values:
	// draft, active, resolved, expired, cancelled
	Status string `json:"status" validate:"required,oneof=draft active resolved expired cancelled" example:"resolved" enums:"draft,active,resolved,expired,cancelled"`
}
