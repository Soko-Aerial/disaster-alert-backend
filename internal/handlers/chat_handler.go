package handlers

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var systemAdminObjectID = primitive.ObjectID([12]byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1,
})

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
	}
}

// CreateConversation godoc
// @Summary Create or open a chat conversation
// @Description Creates or opens a conversation for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user wants to start a chat related to SOS, assistance, report follow-up, emergency contact support, or general help.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - title: Conversation title.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - caseType: sos, assistance, report, or general.
// @Description - caseId: Related case ID.
// @Description - contactId: Related contact ID.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "title": "Flood assistance conversation",
// @Description   "caseType": "assistance",
// @Description   "caseId": "667c9b2f12ab34cd56ef7890",
// @Description   "contactId": "667c9b2f12ab34cd56ef7891"
// @Description }
// @Tags User Chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateConversationRequest true "Conversation payload. title is required."
// @Success 201 {object} map[string]interface{} "Conversation created or opened successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or conversation could not be created."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while creating conversation."
// @Router /chats/conversations [post]
func (h *ChatHandler) CreateConversation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	conversation, err := h.chatService.CreateConversation(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Conversation created successfully",
		conversation,
	)
}

// GetConversations godoc
// @Summary List chat conversations
// @Description Returns chat conversations visible to the current actor.
// @Description
// @Description MOBILE USER USE:
// @Description Users call /chats/conversations to see their own conversations.
// @Description
// @Description ADMIN USE:
// @Description Admins call /admin/chats/conversations to see conversations available to the admin dashboard.
// @Description Admin route requires AdminApiKeyAuth and PrivilegeCodeAuth with `chats:read` permission.
// @Tags Chats
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Conversations fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid authentication."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code for admin route."
// @Failure 500 {object} map[string]interface{} "Failed to fetch conversations."
// @Router /chats/conversations [get]
func (h *ChatHandler) GetConversations(c *gin.Context) {
	userID, role, ok := getChatActorFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	conversations, err := h.chatService.GetUserConversations(userID, role)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch conversations", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Conversations fetched successfully",
		conversations,
	)
}

// GetConversationMessages godoc
// @Summary Get conversation messages
// @Description Returns messages inside a selected chat conversation.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the mobile app or admin dashboard opens a conversation thread.
// @Description
// @Description PATH PARAMETER:
// @Description - id: Conversation ID.
// @Tags Chats
// @Security BearerAuth
// @Produce json
// @Param id path string true "Conversation ID. This is the MongoDB ObjectID of the conversation." example(667c9b2f12ab34cd56ef7890)
// @Success 200 {object} map[string]interface{} "Messages fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid conversation ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid authentication."
// @Failure 403 {object} map[string]interface{} "User is not allowed to access this conversation."
// @Failure 404 {object} map[string]interface{} "Conversation not found."
// @Failure 500 {object} map[string]interface{} "Server error while fetching messages."
// @Router /chats/conversations/{id}/messages [get]
func (h *ChatHandler) GetConversationMessages(c *gin.Context) {
	userID, role, ok := getChatActorFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err)
		return
	}

	messages, err := h.chatService.GetConversationMessages(
		userID,
		conversationID,
		role,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Messages fetched successfully",
		messages,
	)
}

// SendMessage godoc
// @Summary Send chat message
// @Description Sends a message into an existing conversation.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user, admin, responder, or support team needs to reply inside an SOS, assistance, report, or general support conversation.
// @Description
// @Description REAL-TIME EFFECT:
// @Description The message can be delivered through WebSocket updates and may also trigger push/in-app notifications depending on your service logic.
// @Description
// @Description REQUIRED BODY:
// @Description - message: Text message to send.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "message": "Help is on the way. Please remain in a safe location."
// @Description }
// @Tags Chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID. This is the MongoDB ObjectID of the conversation." example(667c9b2f12ab34cd56ef7890)
// @Param request body dto.SendChatMessageRequest true "Chat message payload. message is required."
// @Success 201 {object} map[string]interface{} "Message sent successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, invalid conversation ID, or message could not be sent."
// @Failure 401 {object} map[string]interface{} "Missing or invalid authentication."
// @Failure 403 {object} map[string]interface{} "User is not allowed to send messages in this conversation."
// @Failure 404 {object} map[string]interface{} "Conversation not found."
// @Failure 500 {object} map[string]interface{} "Server error while sending message."
// @Router /chats/conversations/{id}/messages [post]
func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, role, ok := getChatActorFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err)
		return
	}

	var req dto.SendChatMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	message, err := h.chatService.SendMessage(
		userID,
		conversationID,
		role,
		req,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Message sent successfully",
		message,
	)
}

// MarkConversationRead godoc
// @Summary Mark conversation messages as read
// @Description Marks selected messages in a conversation as read for the current actor.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the user/admin opens a conversation and messages should no longer appear as unread.
// @Description
// @Description REQUIRED BODY:
// @Description Send messageIds as a list of message IDs.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "messageIds": [
// @Description     "667c9b2f12ab34cd56ef7892"
// @Description   ]
// @Description }
// @Tags Chats
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID. This is the MongoDB ObjectID of the conversation." example(667c9b2f12ab34cd56ef7890)
// @Param request body dto.MarkConversationReadRequest true "Conversation read payload. messageIds is the list of message IDs to mark as read."
// @Success 200 {object} map[string]interface{} "Messages marked as read."
// @Failure 400 {object} map[string]interface{} "Invalid request body, invalid conversation ID, or messages could not be marked as read."
// @Failure 401 {object} map[string]interface{} "Missing or invalid authentication."
// @Failure 403 {object} map[string]interface{} "User is not allowed to update this conversation."
// @Failure 404 {object} map[string]interface{} "Conversation not found."
// @Failure 500 {object} map[string]interface{} "Server error while marking conversation as read."
// @Router /chats/conversations/{id}/read [put]
func (h *ChatHandler) MarkConversationRead(c *gin.Context) {
	userID, role, ok := getChatActorFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	conversationID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid conversation ID", err)
		return
	}

	var req dto.MarkConversationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.chatService.MarkConversationRead(
		userID,
		conversationID,
		role,
		req,
	); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Messages marked as read",
		nil,
	)
}

func getChatActorFromContext(c *gin.Context) (primitive.ObjectID, string, bool) {
	role := getRoleFromContext(c)

	if isAdminRole(role) {
		return systemAdminObjectID, role, true
	}

	userID, ok := getUserIDFromContext(c)
	if !ok {
		return primitive.NilObjectID, "", false
	}

	if role == "" {
		role = "user"
	}

	return userID, role, true
}

func getRoleFromContext(c *gin.Context) string {
	roleValue, exists := c.Get("role")
	if !exists {
		return "user"
	}

	role, ok := roleValue.(string)
	if !ok {
		return "user"
	}

	role = strings.ToLower(strings.TrimSpace(role))
	if role == "" {
		return "user"
	}

	return role
}

func isAdminRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "super_admin", "responder":
		return true
	default:
		return false
	}
}
