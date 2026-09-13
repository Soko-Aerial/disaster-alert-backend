package dto

// AlertEscalationRequest is used when an existing alert becomes more serious
// or its affected area expands.
type AlertEscalationRequest struct {
	// Severity is optional.
	// If changed to critical, the system sends an update to previous and new affected users.
	Severity string `json:"severity,omitempty" validate:"omitempty,oneof=low medium high critical" example:"critical"`

	// RadiusKm is optional shortcut for expanding a radius alert.
	// If targeting.radiusKm is also provided, targeting.radiusKm takes priority.
	RadiusKm float64 `json:"radiusKm,omitempty" example:"15"`

	// Targeting contains the new alert targeting area.
	// Use this to expand radius, change polygon, change region, or send national alert.
	Targeting AlertTargetingRequest `json:"targeting,omitempty"`

	// Message is an optional update message for the escalation.
	Message string `json:"message,omitempty" example:"The flood area has expanded. Move away from low-lying areas immediately."`

	// Reason explains why the alert is being escalated.
	Reason string `json:"reason,omitempty" example:"Water level has increased and affected area expanded."`
}
