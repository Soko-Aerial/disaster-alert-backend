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
		"Conversation ready successfully",
		conversation,
	)
}

// GetConversations godoc
// @Summary Get chat conversations
// @Description Admin dashboard fetches all user chat conversations.
// @Tags Admin Chats
// @Security AdminApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/chats/conversations [get]
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
// @Description Admin dashboard fetches messages in a conversation.
// @Tags Admin Chats
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/chats/conversations/{id}/messages [get]
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
// @Description Admin dashboard sends a response message to a user conversation.
// @Tags Admin Chats
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID"
// @Param request body map[string]string true "Message body"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/chats/conversations/{id}/messages [post]
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