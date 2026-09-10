package dto

// UpdateAlertPreferenceRequest is used by a mobile user to choose which alert categories they want.
//
// All fields are optional.
// Use true to enable a category and false to disable it.
// Pointer booleans are used so the backend can tell whether a field was omitted or intentionally set to false.
type UpdateAlertPreferenceRequest struct {
	// Fire alerts.
	Fire *bool `json:"fire" example:"true"`

	// Flood alerts.
	Flood *bool `json:"flood" example:"true"`

	// Weather alerts such as storms, heavy rainfall, heat, wind, or severe weather.
	Weather *bool `json:"weather" example:"true"`

	// Earthquake alerts.
	Earthquake *bool `json:"earthquake" example:"false"`

	// Health alerts such as outbreaks, contamination, or public health warnings.
	Health *bool `json:"health" example:"true"`

	// Conflict alerts.
	Conflict *bool `json:"conflict" example:"true"`

	// Drought alerts.
	Drought *bool `json:"drought" example:"false"`

	// Protest or civil unrest alerts.
	Protests *bool `json:"protests" example:"true"`

	// Robbery or crime-related alerts.
	Robbery *bool `json:"robbery" example:"true"`

	// Munitions or explosive-related alerts.
	Munitions *bool `json:"munitions" example:"true"`

	// Galamsey or illegal mining-related alerts.
	Galamsey *bool `json:"galamsey" example:"false"`

	// Unverified activity alerts.
	UnverifiedActivity *bool `json:"unverifiedActivity" example:"false"`

	// CriticalAlerts controls whether the user receives critical/highest priority alerts.
	CriticalAlerts *bool `json:"criticalAlerts" example:"true"`
}
