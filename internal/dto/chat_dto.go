package dto

type CreateConversationRequest struct {
	Title     string `json:"title" binding:"required"`
	CaseType  string `json:"caseType"`
	CaseID    string `json:"caseId"`
	ContactID string `json:"contactId"`
}

type SendChatMessageRequest struct {
	Message string `json:"message" binding:"required"`
}

type MarkConversationReadRequest struct {
	MessageIDs []string `json:"messageIds"`
}