package dto

// CreateConversationRequest is used by a mobile user to create or open a conversation.
//
// Conversations can be linked to SOS cases, assistance requests, reports,
// emergency contacts, or general support.
type CreateConversationRequest struct {
	// Title is the conversation title shown in the chat list.
	Title string `json:"title" binding:"required" example:"Flood assistance conversation"`

	// CaseType describes what this conversation is connected to.
	//
	// Allowed values:
	// sos, assistance, report, general
	CaseType string `json:"caseType" example:"assistance" enums:"sos,assistance,report,general"`

	// CaseID is the MongoDB ObjectID of the related SOS, assistance request, or report.
	CaseID string `json:"caseId" example:"667c9b2f12ab34cd56ef7890"`

	// ContactID is the MongoDB ObjectID of the emergency contact or participant, if applicable.
	ContactID string `json:"contactId" example:"667c9b2f12ab34cd56ef7891"`
}

// SendChatMessageRequest is used to send a chat message in a conversation.
type SendChatMessageRequest struct {
	Message string `json:"message" binding:"required" example:"Help is on the way. Please remain in a safe location."`
}

// MarkConversationReadRequest is used to mark messages in a conversation as read.
type MarkConversationReadRequest struct {
	// MessageIDs is the list of chat message IDs to mark as read.
	MessageIDs []*string `json:"messageIds" example:"667c9b2f12ab34cd56ef7892"`
}
