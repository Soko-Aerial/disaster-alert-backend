package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmergencyContactHandler struct {
	contactService *services.EmergencyContactService
}

func NewEmergencyContactHandler(
	contactService *services.EmergencyContactService,
) *EmergencyContactHandler {
	return &EmergencyContactHandler{
		contactService: contactService,
	}
}

func (h *EmergencyContactHandler) CreateContact(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.CreateEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contact, err := h.contactService.CreateContact(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Emergency contact created successfully",
		contact,
	)
}

func (h *EmergencyContactHandler) GetContacts(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	contacts, err := h.contactService.GetUserContacts(userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch emergency contacts", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency contacts fetched successfully",
		contacts,
	)
}

func (h *EmergencyContactHandler) GetContactByID(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	contactID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	contact, err := h.contactService.GetContactByID(contactID, userID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Emergency contact not found", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency contact fetched successfully",
		contact,
	)
}

func (h *EmergencyContactHandler) UpdateContact(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	contactID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	var req dto.UpdateEmergencyContactRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	contact, err := h.contactService.UpdateContact(contactID, userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency contact updated successfully",
		contact,
	)
}

func (h *EmergencyContactHandler) DeleteContact(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	contactID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid contact ID", err)
		return
	}

	if err := h.contactService.DeleteContact(contactID, userID); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete emergency contact", err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Emergency contact deleted successfully",
		nil,
	)
}
