package dto

type CreateSOSRequest struct {
	EmergencyType string `json:"emergencyType" validate:"required" example:"medical" enums:"medical,fire,security,flood,accident,other"`

	Message string `json:"message,omitempty" example:"I need urgent help at my location."`

	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	Address string `json:"address,omitempty" example:"Circle, Accra"`

	Accuracy float64 `json:"accuracy,omitempty" example:"8.5"`

	IsLiveTracking bool `json:"isLiveTracking" example:"true"`
}

type UpdateSOSStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active assigned resolved cancelled" example:"resolved" enums:"active,assigned,resolved,cancelled"`
}