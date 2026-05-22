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