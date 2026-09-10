package dto

// CreateSOSRequest is used by a mobile user to trigger an emergency SOS.
//
// SOS is for urgent emergency situations where the user needs immediate help.
type CreateSOSRequest struct {
	// EmergencyType describes the kind of emergency.
	//
	// Allowed values:
	// medical, fire, security, flood, accident, other
	EmergencyType string `json:"emergencyType" validate:"required" example:"medical" enums:"medical,fire,security,flood,accident,other"`

	// Message is an optional emergency message from the user.
	Message string `json:"message,omitempty" example:"I need urgent help at my location."`

	// Latitude is the user's current GPS latitude.
	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	// Longitude is the user's current GPS longitude.
	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	// Address is the readable location, landmark, street, or area name.
	Address string `json:"address,omitempty" example:"Circle, Accra"`

	// Accuracy is the GPS accuracy in meters, if available from the mobile device.
	Accuracy float64 `json:"accuracy,omitempty" example:"8.5"`

	// IsLiveTracking tells the backend whether the user wants location updates
	// to continue while the emergency is active.
	IsLiveTracking bool `json:"isLiveTracking" example:"true"`
}

// UpdateSOSStatusRequest is used by an admin to update the state of an SOS request.
type UpdateSOSStatusRequest struct {
	// Status is the new SOS status.
	//
	// Allowed values:
	// active, assigned, resolved, cancelled
	Status string `json:"status" validate:"required,oneof=active assigned resolved cancelled" example:"resolved" enums:"active,assigned,resolved,cancelled"`
}
