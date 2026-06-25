package dto

type CreateAssistanceRequest struct {
	AssistanceType string `json:"assistanceType" validate:"required" example:"medical" enums:"medical,rescue,food,shelter,security,evacuation,other"`

	UrgencyLevel string `json:"urgencyLevel" validate:"required" example:"high" enums:"low,medium,high,critical"`

	AffectedIndividuals int `json:"affectedIndividuals" validate:"required,min=1" example:"3"`

	OtherInformation string `json:"otherInformation,omitempty" example:"One person is injured and needs urgent medical support."`

	Latitude float64 `json:"latitude" validate:"required" example:"5.6037"`

	Longitude float64 `json:"longitude" validate:"required" example:"-0.1870"`

	Address string `json:"address,omitempty" example:"Circle, Accra"`
}

type UpdateAssistanceStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending accepted en_route arrived completed cancelled rejected" example:"en_route" enums:"pending,accepted,en_route,arrived,completed,cancelled,rejected"`
}