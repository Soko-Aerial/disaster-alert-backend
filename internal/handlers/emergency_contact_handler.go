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

// CreateContact godoc
// @Summary Create emergency contact
// @Description Creates a new emergency contact for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user adds a family member, friend, responder, organization, or government contact that can be reached during emergencies.
// @Description
// @Description AUTH REQUIRED:
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - name: Contact name.
// @Description - phone: Contact phone number.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - email, relationship, type, organization, address, isPrimary, isGovernment.
// @Tags Emergency Contacts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateEmergencyContactRequest true "Emergency contact payload. name and phone are required."
// @Success 201 {object} map[string]interface{} "Emergency contact created successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or contact could not be created."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while creating emergency contact."
// @Router /emergency-contacts [post]
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

// GetContacts godoc
// @Summary List emergency contacts
// @Description Returns all emergency contacts belonging to the authenticated user.
// @Tags Emergency Contacts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Emergency contacts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch emergency contacts."
// @Router /emergency-contacts [get]
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

// GetContactByID godoc
// @Summary Get one emergency contact
// @Description Returns one emergency contact belonging to the authenticated user.
// @Tags Emergency Contacts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Emergency contact ID. This is the MongoDB ObjectID of the contact." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Emergency contact fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid contact ID."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "Emergency contact not found."
// @Router /emergency-contacts/{id} [get]
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

// UpdateContact godoc
// @Summary Update emergency contact
// @Description Updates an existing emergency contact belonging to the authenticated user.
// @Description
// @Description Send only the fields you want to change.
// @Tags Emergency Contacts
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Emergency contact ID. This is the MongoDB ObjectID of the contact." example(66e19b71c8f2a2b4d1234567)
// @Param request body dto.UpdateEmergencyContactRequest true "Emergency contact update payload. All fields are optional."
// @Success 200 {object} map[string]interface{} "Emergency contact updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid contact ID, invalid request body, or contact could not be updated."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "Emergency contact not found."
// @Router /emergency-contacts/{id} [put]
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

// DeleteContact godoc
// @Summary Delete emergency contact
// @Description Deletes an emergency contact belonging to the authenticated user.
// @Tags Emergency Contacts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Emergency contact ID. This is the MongoDB ObjectID of the contact." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Emergency contact deleted successfully."
// @Failure 400 {object} map[string]interface{} "Invalid contact ID."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 404 {object} map[string]interface{} "Emergency contact not found."
// @Failure 500 {object} map[string]interface{} "Failed to delete emergency contact."
// @Router /emergency-contacts/{id} [delete]
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
