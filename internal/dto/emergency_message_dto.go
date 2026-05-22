package dto

type EmergencyMessageLocationRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}

type CreateEmergencyMessageRequest struct {
	Title   string `json:"title"`
	Message string `json:"message" binding:"required"`

	Type string `json:"type"`

	Location *EmergencyMessageLocationRequest `json:"location"`
}

type SendEmergencyMessageRequest struct {
	Title   string `json:"title"`
	Message string `json:"message" binding:"required"`

	ContactIDs []string `json:"contactIds"`

	SendToAllContacts bool `json:"sendToAllContacts"`

	Location *EmergencyMessageLocationRequest `json:"location"`
}