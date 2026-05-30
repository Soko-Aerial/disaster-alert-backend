package services

import (
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	conversationRepo         *repositories.ConversationRepository
	messageRepo              *repositories.ChatMessageRepository
	eventNotificationService *EventNotificationService
}

func NewChatService(
	conversationRepo *repositories.ConversationRepository,
	messageRepo *repositories.ChatMessageRepository,
	eventNotificationService *EventNotificationService,
) *ChatService {
	return &ChatService{
		conversationRepo:         conversationRepo,
		messageRepo:              messageRepo,
		eventNotificationService: eventNotificationService,
	}
}

func (s *ChatService) CreateConversation(
	userID primitive.ObjectID,
	req dto.CreateConversationRequest,
) (*models.Conversation, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, errors.New("conversation title is required")
	}

	caseType := strings.TrimSpace(req.CaseType)
	if caseType == "" {
		caseType = "general"
	}

	if caseType == "response_center" {
		existingConversation, err := s.conversationRepo.FindUserConversationByType(
			userID,
			"response_center",
		)

		if err == nil && existingConversation != nil {
			return existingConversation, nil
		}
	}

	var caseID *primitive.ObjectID
	var contactID *primitive.ObjectID

	if strings.TrimSpace(req.CaseID) != "" {
		parsedCaseID, err := primitive.ObjectIDFromHex(req.CaseID)
		if err != nil {
			return nil, errors.New("invalid case id")
		}

		caseID = &parsedCaseID
	}

	if strings.TrimSpace(req.ContactID) != "" {
		parsedContactID, err := primitive.ObjectIDFromHex(req.ContactID)
		if err != nil {
			return nil, errors.New("invalid contact id")
		}

		contactID = &parsedContactID
	}

	if caseType == "contact" {
		if contactID == nil {
			return nil, errors.New("contact id is required for contact conversation")
		}

		existingConversation, err := s.conversationRepo.FindContactConversation(
			userID,
			*contactID,
		)

		if err == nil && existingConversation != nil {
			return existingConversation, nil
		}
	}

	conversation := models.Conversation{
		UserID:         userID,
		Title:          title,
		CaseType:       caseType,
		CaseID:         caseID,
		ContactID:      contactID,
		ParticipantIDs: []primitive.ObjectID{userID},
		Status:         "open",
	}

	return s.conversationRepo.Create(conversation)
}

func (s *ChatService) GetUserConversations(
	userID primitive.ObjectID,
	role string,
) ([]models.Conversation, error) {
	if isStaffRole(role) {
		return s.conversationRepo.FindAll()
	}

	return s.conversationRepo.FindByUserID(userID)
}

func (s *ChatService) GetConversationMessages(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	role string,
) ([]models.ChatMessage, error) {
	_, err := s.conversationRepo.FindByIDWithAccess(
		conversationID,
		userID,
		role,
	)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	return s.messageRepo.FindByConversationID(conversationID)
}

func (s *ChatService) SendMessage(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	role string,
	req dto.SendChatMessageRequest,
) (*models.ChatMessage, error) {
	messageText := strings.TrimSpace(req.Message)
	if messageText == "" {
		return nil, errors.New("message is required")
	}

	cleanRole := strings.TrimSpace(role)
	if cleanRole == "" {
		cleanRole = "user"
	}

	conversation, err := s.conversationRepo.FindByIDWithAccess(
		conversationID,
		userID,
		cleanRole,
	)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	now := time.Now().UTC()

	message := models.ChatMessage{
		ConversationID: conversationID,
		SenderID:       userID,
		SenderRole:     cleanRole,
		Message:        messageText,
		Status:         "sent",
	}

	createdMessage, err := s.messageRepo.Create(message)
	if err != nil {
		return nil, err
	}

	_ = s.conversationRepo.UpdateLastMessage(
		conversationID,
		messageText,
		now,
	)

	if s.eventNotificationService != nil {
		s.notifyChatReceiverAsync(
			conversation,
			userID,
			cleanRole,
			messageText,
		)
	}

	return createdMessage, nil
}

func (s *ChatService) MarkConversationRead(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	role string,
	req dto.MarkConversationReadRequest,
) error {
	_, err := s.conversationRepo.FindByIDWithAccess(
		conversationID,
		userID,
		role,
	)
	if err != nil {
		return errors.New("conversation not found")
	}

	messageIDs := make([]primitive.ObjectID, 0)

	for _, idString := range req.MessageIDs {
		id, err := primitive.ObjectIDFromHex(idString)
		if err != nil {
			return errors.New("invalid message id: " + idString)
		}

		messageIDs = append(messageIDs, id)
	}

	if len(messageIDs) == 0 {
		return nil
	}

	return s.messageRepo.MarkMessagesRead(
		conversationID,
		userID,
		messageIDs,
	)
}

func (s *ChatService) notifyChatReceiverAsync(
	conversation *models.Conversation,
	senderID primitive.ObjectID,
	senderRole string,
	messageText string,
) {
	if conversation == nil {
		return
	}

	role := strings.ToLower(strings.TrimSpace(senderRole))

	// Admin/responder reply should notify the user who owns the conversation.
	if isStaffRole(role) {
		if conversation.UserID.IsZero() || conversation.UserID == senderID {
			return
		}

		go s.eventNotificationService.NotifyUserForChatResponse(
			conversation.UserID,
			conversation.ID,
			"Emergency Response",
			shortMessagePreview(messageText),
		)

		return
	}

	// Fallback for future multi-participant conversations.
	for _, participantID := range conversation.ParticipantIDs {
		if participantID.IsZero() || participantID == senderID {
			continue
		}

		go s.eventNotificationService.NotifyUserForChatResponse(
			participantID,
			conversation.ID,
			"Emergency Chat",
			shortMessagePreview(messageText),
		)
	}
}

func shortMessagePreview(message string) string {
	message = strings.TrimSpace(message)

	if len(message) <= 120 {
		return message
	}

	return message[:120] + "..."
}

func isStaffRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "super_admin", "responder":
		return true
	default:
		return false
	}
}