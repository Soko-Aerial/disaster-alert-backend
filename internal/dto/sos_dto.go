package dto

type CreateSOSRequest struct {
	EmergencyType  string  `json:"emergencyType" validate:"required"`
	Message        string  `json:"message,omitempty"`
	Latitude       float64 `json:"latitude" validate:"required"`
	Longitude      float64 `json:"longitude" validate:"required"`
	Address        string  `json:"address,omitempty"`
	Accuracy       float64 `json:"accuracy,omitempty"`
	IsLiveTracking bool    `json:"isLiveTracking"`
}

type UpdateSOSStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active assigned resolved cancelled"`
}