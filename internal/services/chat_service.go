package services

import (
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ChatService struct {
	conversationRepo         *repositories.ConversationRepository
	messageRepo              *repositories.ChatMessageRepository
	userRepo                 *repositories.UserRepository
	eventNotificationService *EventNotificationService
	broadcaster              *websocket.Broadcaster
}

func NewChatService(
	conversationRepo *repositories.ConversationRepository,
	messageRepo *repositories.ChatMessageRepository,
	userRepo *repositories.UserRepository,
	eventNotificationService *EventNotificationService,
	broadcaster *websocket.Broadcaster,
) *ChatService {
	return &ChatService{
		conversationRepo:         conversationRepo,
		messageRepo:              messageRepo,
		userRepo:                 userRepo,
		eventNotificationService: eventNotificationService,
		broadcaster:              broadcaster,
	}
}

func (s *ChatService) CreateConversation(
	userID primitive.ObjectID,
	req dto.CreateConversationRequest,
) (map[string]interface{}, error) {
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
			return s.buildConversationResponse(existingConversation), nil
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
			return s.buildConversationResponse(existingConversation), nil
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

	createdConversation, err := s.conversationRepo.Create(conversation)
	if err != nil {
		return nil, err
	}

	return s.buildConversationResponse(createdConversation), nil
}

func (s *ChatService) GetUserConversations(
	userID primitive.ObjectID,
	role string,
) ([]map[string]interface{}, error) {
	var conversations []models.Conversation
	var err error

	if isStaffRole(role) {
		conversations, err = s.conversationRepo.FindAll()
	} else {
		conversations, err = s.conversationRepo.FindByUserID(userID)
	}

	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(conversations))

	for i := range conversations {
		response = append(
			response,
			s.buildConversationResponse(&conversations[i]),
		)
	}

	return response, nil
}

func (s *ChatService) GetConversationMessages(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	role string,
) ([]map[string]interface{}, error) {
	_, err := s.conversationRepo.FindByIDWithAccess(
		conversationID,
		userID,
		role,
	)
	if err != nil {
		return nil, errors.New("conversation not found")
	}

	messages, err := s.messageRepo.FindByConversationID(conversationID)
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(messages))

	for i := range messages {
		response = append(
			response,
			s.buildChatMessageResponse(&messages[i]),
		)
	}

	return response, nil
}

func (s *ChatService) SendMessage(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	role string,
	req dto.SendChatMessageRequest,
) (map[string]interface{}, error) {
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

	response := s.buildChatMessageResponse(createdMessage)

	if s.broadcaster != nil {
		if isStaffRole(cleanRole) {
			s.broadcaster.BroadcastChatMessage(
				conversation.UserID.Hex(),
				response,
			)
		} else {
			s.broadcaster.BroadcastChatMessageToAdmins(response)
		}
	}

	return response, nil
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

func (s *ChatService) buildConversationResponse(
	conversation *models.Conversation,
) map[string]interface{} {
	if conversation == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":             conversation.ID.Hex(),
		"userId":         conversation.UserID.Hex(),
		"user":           s.buildUserSummary(conversation.UserID),
		"title":          conversation.Title,
		"caseType":       conversation.CaseType,
		"caseId":         objectIDPointerToString(conversation.CaseID),
		"contactId":      objectIDPointerToString(conversation.ContactID),
		"participantIds": objectIDsToStrings(conversation.ParticipantIDs),
		"status":         conversation.Status,
	}
}

func (s *ChatService) buildChatMessageResponse(
	message *models.ChatMessage,
) map[string]interface{} {
	if message == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":             message.ID.Hex(),
		"conversationId": message.ConversationID.Hex(),
		"senderId":       message.SenderID.Hex(),
		"senderRole":     message.SenderRole,
		"sender":         s.buildUserSummary(message.SenderID),
		"message":        message.Message,
		"status":         message.Status,
		"createdAt":      message.CreatedAt,
	}
}

func (s *ChatService) buildUserSummary(
	userID primitive.ObjectID,
) map[string]interface{} {
	userData := map[string]interface{}{
		"id":       userID.Hex(),
		"name":     "User",
		"email":    "",
		"phone":    "",
		"gender":   "",
		"role":     "",
		"location": nil,
	}

	if s.userRepo == nil {
		return userData
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil || user == nil {
		return userData
	}

	userData["name"] = user.Name
	userData["email"] = user.Email
	userData["phone"] = user.Phone
	userData["gender"] = user.Gender
	userData["role"] = user.Role

	if user.Location != nil {
		userData["location"] = map[string]interface{}{
			"name":      user.Location.Name,
			"country":   user.Location.Country,
			"region":    user.Location.Region,
			"address":   user.Location.Address,
			"latitude":  user.Location.Latitude,
			"longitude": user.Location.Longitude,
			"source":    user.Location.Source,
			"isDefault": user.Location.IsDefault,
		}
	}

	return userData
}

func objectIDPointerToString(id *primitive.ObjectID) string {
	if id == nil || id.IsZero() {
		return ""
	}

	return id.Hex()
}

func objectIDsToStrings(ids []primitive.ObjectID) []string {
	values := make([]string, 0, len(ids))

	for _, id := range ids {
		if id.IsZero() {
			continue
		}

		values = append(values, id.Hex())
	}

	return values
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