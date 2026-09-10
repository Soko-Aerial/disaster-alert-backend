package dto

// CreateAssistanceRequest is used by a mobile user to request help.
//
// This is for non-SOS support requests such as medical help, rescue,
// food, shelter, security support, evacuation, or other assistance.
type CreateAssistanceRequest struct {
	// AssistanceType describes the kind of help the user needs.
	//
	// Allowed values:
	// medical, rescue, food, shelter, security, evacuation, other
	AssistanceType string `json:"assistanceType" validate:"required" example:"medical" enums:"medical,rescue,food,shelter,security,evacuation,other"`

	// UrgencyLevel describes how urgent the assistance request is.
	//
	// Allowed values:
	// low, medium, high, critical
	UrgencyLevel string `json:"urgencyLevel" validate:"required" example:"high" enums:"low,medium,high,critical"`

	// AffectedIndividuals is the number of people who need help.
	//
	// Minimum value: 1
	AffectedIndividuals int `json:"affectedIndividuals" validate:"required,min=1" example:"3"`

	// OtherInformation gives extra details about the situation.
	//
	// Example: injuries, blocked access, trapped people, security concern,
	// food shortage, shelter need, or evacuation challenge.
	OtherInformation string `json:"otherInformation,omitempty" example:"One person is injured and needs urgent medical support."`

	// Latitude is the user's current GPS latitude.
	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	// Longitude is the user's current GPS longitude.
	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	// Address is the readable location, landmark, street, or area name.
	Address string `json:"address,omitempty" example:"Circle, Accra"`
}

// UpdateAssistanceStatusRequest is used by an admin to update the state
// of an assistance request.
type UpdateAssistanceStatusRequest struct {
	// Status is the new assistance request status.
	//
	// Allowed values:
	// pending, accepted, en_route, arrived, completed, cancelled, rejected
	Status string `json:"status" validate:"required,oneof=pending accepted en_route arrived completed cancelled rejected" example:"en_route" enums:"pending,accepted,en_route,arrived,completed,cancelled,rejected"`
}
