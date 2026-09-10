package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SOSHandler struct {
	sosService *services.SOSService
	validator  *validator.Validate
}

func NewSOSHandler(sosService *services.SOSService) *SOSHandler {
	return &SOSHandler{
		sosService: sosService,
		validator:  validator.New(),
	}
}

// CreateSOSRequest godoc
// @Summary Trigger an emergency SOS
// @Description Allows an authenticated mobile user to trigger an urgent SOS emergency request.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user is in immediate danger and needs urgent help.
// @Description Examples: medical emergency, fire, security threat, flood danger, road accident, or other emergency.
// @Description
// @Description DIFFERENCE BETWEEN SOS AND ASSISTANCE:
// @Description - SOS is for immediate emergency panic situations.
// @Description - Assistance is for structured help requests such as food, shelter, evacuation, or medical support.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - emergencyType: Type of emergency. Example: medical.
// @Description - latitude: User's current latitude.
// @Description - longitude: User's current longitude.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - message: Extra message from the user.
// @Description - address: Human-readable location or landmark.
// @Description - accuracy: GPS accuracy in meters.
// @Description - isLiveTracking: true if location tracking should continue while SOS is active.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "emergencyType": "medical",
// @Description   "message": "I need urgent help at my location.",
// @Description   "latitude": 5.6037,
// @Description   "longitude": -0.1870,
// @Description   "address": "Circle, Accra",
// @Description   "accuracy": 8.5,
// @Description   "isLiveTracking": true
// @Description }
// @Tags User SOS
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateSOSRequest true "SOS request payload. emergencyType, latitude, and longitude are required."
// @Success 201 {object} map[string]interface{} "SOS request triggered successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, or SOS could not be created."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while triggering SOS request."
// @Router /sos [post]
func (h *SOSHandler) CreateSOSRequest(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"User not found in request context",
			nil,
		)
		return
	}

	userID, ok := userIDValue.(string)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"Invalid user id in request context",
			nil,
		)
		return
	}

	var req dto.CreateSOSRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	sos, err := h.sosService.CreateSOSRequest(userID, req)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"SOS request triggered successfully",
		sos,
	)
}

// GetSOSRequests godoc
// @Summary List SOS requests
// @Description Returns SOS emergency requests submitted by users.
// @Description
// @Description ADMIN USE:
// @Description Admins use this endpoint to view urgent SOS emergencies that need response, assignment, or resolution.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `sos:read`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:read
// @Tags Admin SOS
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "SOS requests fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch SOS requests."
// @Router /admin/sos [get]
func (h *SOSHandler) GetSOSRequests(c *gin.Context) {
	requests, err := h.sosService.GetSOSRequests()
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch SOS requests",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"SOS requests fetched successfully",
		requests,
	)
}

// GetSOSByID godoc
// @Summary Get one SOS request
// @Description Returns details of a single SOS emergency request.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin wants to inspect one SOS request, view its emergency type, message, location, live tracking flag, and response status.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `sos:read`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:read
// @Tags Admin SOS
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "SOS ID. This is the MongoDB ObjectID of the SOS request." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "SOS request fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid SOS ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "SOS request not found."
// @Failure 500 {object} map[string]interface{} "Server error while fetching SOS request."
// @Router /admin/sos/{id} [get]
func (h *SOSHandler) GetSOSByID(c *gin.Context) {
	sosID := c.Param("id")

	sos, err := h.sosService.GetSOSByID(sosID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusNotFound,
			"SOS request not found",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"SOS request fetched successfully",
		sos,
	)
}

// UpdateSOSStatus godoc
// @Summary Update SOS status
// @Description Updates the status of an SOS emergency request from the admin dashboard.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin or responder takes over an SOS case, assigns it, resolves it, or cancels it.
// @Description
// @Description MOBILE APP EFFECT:
// @Description The mobile app can stop SOS alarm/live tracking when status becomes resolved or cancelled, depending on your frontend logic.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `sos:update_status`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:update_status
// @Description
// @Description ALLOWED STATUS VALUES:
// @Description - active: SOS is still active.
// @Description - assigned: A responder/admin has taken the case.
// @Description - resolved: Emergency has been handled.
// @Description - cancelled: SOS was cancelled.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "status": "assigned"
// @Description }
// @Tags Admin SOS
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "SOS ID. This is the MongoDB ObjectID of the SOS request." example(66e19b71c8f2a2b4d1234567)
// @Param request body dto.UpdateSOSStatusRequest true "SOS status update payload. status is required."
// @Success 200 {object} map[string]interface{} "SOS status updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid status."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "SOS request not found."
// @Failure 500 {object} map[string]interface{} "Server error while updating SOS status."
// @Router /admin/sos/{id}/status [put]
func (h *SOSHandler) UpdateSOSStatus(c *gin.Context) {
	sosID := c.Param("id")

	var req dto.UpdateSOSStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	sos, err := h.sosService.UpdateSOSStatus(sosID, req.Status)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"SOS status updated successfully",
		sos,
	)
}
