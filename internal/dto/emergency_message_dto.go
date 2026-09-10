package dto

// EmergencyMessageLocationRequest contains optional location data attached to an emergency message.
type EmergencyMessageLocationRequest struct {
	// Latitude is the user's GPS latitude.
	Latitude float64 `json:"latitude" example:"5.6037"`

	// Longitude is the user's GPS longitude.
	Longitude float64 `json:"longitude" example:"-0.1870"`

	// Address is the readable location or landmark.
	Address string `json:"address" example:"Circle, Accra"`
}

// CreateEmergencyMessageRequest saves an emergency message draft/template.
//
// It does not necessarily send the message immediately.
type CreateEmergencyMessageRequest struct {
	// Title is the short title of the emergency message.
	Title string `json:"title" example:"Need urgent help"`

	// Message is the full emergency message text.
	Message string `json:"message" binding:"required" example:"I need urgent help at my current location."`

	// Type describes the message category.
	//
	// Examples:
	// free_text, sos, assistance, report, template
	Type string `json:"type" example:"free_text" enums:"free_text,sos,assistance,report,template"`

	// Location is optional location information attached to the message.
	Location *EmergencyMessageLocationRequest `json:"location"`
}

// SendEmergencyMessageRequest sends an emergency message to selected contacts or all contacts.
type SendEmergencyMessageRequest struct {
	// Title is the short title of the emergency message.
	Title string `json:"title" example:"Emergency help needed"`

	// Message is the full emergency message text.
	Message string `json:"message" binding:"required" example:"I need urgent help at my current location."`

	// ContactIDs is the list of emergency contact IDs to send the message to.
	//
	// Leave empty if sendToAllContacts is true.
	ContactIDs []string `json:"contactIds" example:"66e19b71c8f2a2b4d1234567"`

	// SendToAllContacts sends the message to all active emergency contacts when true.
	SendToAllContacts bool `json:"sendToAllContacts" example:"true"`

	// Location is optional location information attached to the message.
	Location *EmergencyMessageLocationRequest `json:"location"`
}
