package dto

type CreateAssistanceRequest struct {
	AssistanceType      string  `json:"assistanceType" validate:"required"`
	UrgencyLevel        string  `json:"urgencyLevel" validate:"required"`
	AffectedIndividuals int     `json:"affectedIndividuals" validate:"required,min=1"`
	OtherInformation    string  `json:"otherInformation,omitempty"`
	Latitude            float64 `json:"latitude" validate:"required"`
	Longitude           float64 `json:"longitude" validate:"required"`
	Address             string  `json:"address,omitempty"`
}

type UpdateAssistanceStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending accepted en_route arrived completed cancelled rejected"`
}