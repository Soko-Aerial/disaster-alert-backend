package dto

type CreateConversationRequest struct {
	Title string `json:"title" binding:"required" example:"Flood assistance conversation"`

	CaseType string `json:"caseType" example:"assistance" enums:"sos,assistance,report,general"`

	CaseID string `json:"caseId" example:"667c9b2f12ab34cd56ef7890"`

	ContactID string `json:"contactId" example:"667c9b2f12ab34cd56ef7891"`
}

type SendChatMessageRequest struct {
	Message string `json:"message" binding:"required" example:"Help is on the way. Please remain in a safe location."`
}

type MarkConversationReadRequest struct {
	MessageIDs []string `json:"messageIds" example:"667c9b2f12ab34cd56ef7892"`
}