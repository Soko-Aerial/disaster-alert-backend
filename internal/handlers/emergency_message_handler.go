package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmergencyMessageHandler struct {
	messageService *services.EmergencyMessageService
}

func NewEmergencyMessageHandler(
	messageService *services.EmergencyMessageService,
) *EmergencyMessageHandler {
	return &EmergencyMessageHandler{
		messageService: messageService,
	}
}

// CreateMessage godoc
// @Summary Create emergency message
// @Description Saves an emergency message for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user wants to save a reusable emergency message or draft.
// @Tags Emergency Messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateEmergencyMessageRequest true "Emergency message payload. message is required."
// @Success 201 {object} map[string]interface{} "Emergency message created successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or message could not be created."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Router /emergency-messages [post]
func (h *EmergencyMessageHandler) CreateMessage(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.CreateEmergencyMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	message, err := h.messageService.CreateMessage(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Emergency message created successfully",
		message,
	)
}

// GetMessages godoc
// @Summary List emergency messages
// @Description Returns emergency messages saved or sent by the authenticated user.
// @Tags Emergency Messages
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Emergency messages fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch emergency messages."
// @Router /emergency-messages [get]
func (h *EmergencyMessageHandler) GetMessages(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	messages, err := h.messageService.GetUserMessages(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch emergency messages", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency messages fetched successfully",
		messages,
	)
}

// GetMessageByID godoc
// @Summary Get one emergency message
// @Description Returns one emergency message belonging to the authenticated user.
// @Tags Emergency Messages
// @Security BearerAuth
// @Produce json
// @Param id path string true "Emergency message ID. This is the MongoDB ObjectID of the message." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Emergency message fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid message ID."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "Emergency message not found."
// @Router /emergency-messages/{id} [get]
func (h *EmergencyMessageHandler) GetMessageByID(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	messageID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid message ID", err)
		return
	}

	message, err := h.messageService.GetMessageByID(messageID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Emergency message not found", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency message fetched successfully",
		message,
	)
}

// DeleteMessage godoc
// @Summary Delete emergency message
// @Description Deletes an emergency message belonging to the authenticated user.
// @Tags Emergency Messages
// @Security BearerAuth
// @Produce json
// @Param id path string true "Emergency message ID. This is the MongoDB ObjectID of the message." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Emergency message deleted successfully."
// @Failure 400 {object} map[string]interface{} "Invalid message ID."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "Emergency message not found."
// @Failure 500 {object} map[string]interface{} "Failed to delete emergency message."
// @Router /emergency-messages/{id} [delete]
func (h *EmergencyMessageHandler) DeleteMessage(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	messageID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid message ID", err)
		return
	}

	if err := h.messageService.DeleteMessage(messageID, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete emergency message", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency message deleted successfully",
		nil,
	)
}

// SendMessage godoc
// @Summary Send emergency message
// @Description Sends an emergency message to selected emergency contacts or to all active contacts.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the user wants to notify emergency contacts with a free-text message and optional location.
// @Description
// @Description REQUIRED FIELDS:
// @Description - message: The message to send.
// @Description
// @Description CONTACT OPTIONS:
// @Description - To send to selected contacts, provide contactIds.
// @Description - To send to all contacts, set sendToAllContacts to true.
// @Tags Emergency Messages
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SendEmergencyMessageRequest true "Send emergency message payload. message is required."
// @Success 201 {object} map[string]interface{} "Emergency message queued successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or message could not be sent."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Router /emergency-messages/send [post]
func (h *EmergencyMessageHandler) SendMessage(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.SendEmergencyMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	message, err := h.messageService.SendFreeMessage(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Emergency message queued successfully",
		message,
	)
}
