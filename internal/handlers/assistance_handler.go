package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AssistanceHandler struct {
	assistanceService *services.AssistanceService
	validator         *validator.Validate
}

func NewAssistanceHandler(
	assistanceService *services.AssistanceService,
) *AssistanceHandler {
	return &AssistanceHandler{
		assistanceService: assistanceService,
		validator:         validator.New(),
	}
}

// CreateAssistanceRequest godoc
// @Summary Submit an assistance request
// @Description Allows an authenticated mobile user to request help during or after an emergency.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a user needs medical help, rescue, food, shelter, security support, evacuation, or other assistance.
// @Description
// @Description DIFFERENCE BETWEEN SOS AND ASSISTANCE:
// @Description - SOS is for immediate emergency panic situations.
// @Description - Assistance is for structured help requests that responders/admins can review and manage.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - assistanceType: Type of help needed. Example: medical.
// @Description - urgencyLevel: low, medium, high, or critical.
// @Description - affectedIndividuals: Number of people needing help. Minimum is 1.
// @Description - latitude: User/request latitude.
// @Description - longitude: User/request longitude.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - otherInformation: Extra details about the situation.
// @Description - address: Human-readable location or landmark.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "assistanceType": "medical",
// @Description   "urgencyLevel": "high",
// @Description   "affectedIndividuals": 3,
// @Description   "otherInformation": "One person is injured and needs urgent medical support.",
// @Description   "latitude": 5.6037,
// @Description   "longitude": -0.1870,
// @Description   "address": "Circle, Accra"
// @Description }
// @Tags User Assistance
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAssistanceRequest true "Assistance request payload. assistanceType, urgencyLevel, affectedIndividuals, latitude, and longitude are required."
// @Success 201 {object} map[string]interface{} "Assistance request submitted successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, or request could not be created."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while submitting assistance request."
// @Router /assistance [post]
func (h *AssistanceHandler) CreateAssistanceRequest(c *gin.Context) {
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

	var req dto.CreateAssistanceRequest

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

	assistanceRequest, err := h.assistanceService.CreateAssistanceRequest(userID, req)
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
		"Assistance request submitted successfully",
		assistanceRequest,
	)
}

// GetAssistanceRequests godoc
// @Summary List assistance requests
// @Description Returns assistance requests submitted by users.
// @Description
// @Description ADMIN USE:
// @Description Admins use this endpoint to view requests that need review, assignment, dispatch, or completion.
// @Description
// @Description REQUIRED HEADERS FOR ADMIN ROUTE:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `assistance:read`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:read
// @Tags Admin Assistance
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Assistance requests fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch assistance requests."
// @Router /admin/assistance [get]
func (h *AssistanceHandler) GetAssistanceRequests(c *gin.Context) {
	requests, err := h.assistanceService.GetAssistanceRequests()
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch assistance requests",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Assistance requests fetched successfully",
		requests,
	)
}

// GetAssistanceRequestByID godoc
// @Summary Get one assistance request
// @Description Returns the full details of a single assistance request.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin wants to inspect one assistance request before accepting, dispatching, updating, completing, cancelling, or rejecting it.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `assistance:read`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:read
// @Tags Admin Assistance
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Assistance request ID. This is the MongoDB ObjectID of the assistance request." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Assistance request fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid assistance request ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Assistance request not found."
// @Failure 500 {object} map[string]interface{} "Server error while fetching assistance request."
// @Router /admin/assistance/{id} [get]
func (h *AssistanceHandler) GetAssistanceRequestByID(c *gin.Context) {
	requestID := c.Param("id")

	request, err := h.assistanceService.GetAssistanceRequestByID(requestID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusNotFound,
			"Assistance request not found",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Assistance request fetched successfully",
		request,
	)
}

// UpdateAssistanceStatus godoc
// @Summary Update assistance request status
// @Description Updates the status of a user assistance request from the admin dashboard.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin accepts a request, dispatches help, marks responders as en route, confirms arrival, completes the request, cancels it, or rejects it.
// @Description
// @Description MOBILE APP EFFECT:
// @Description The mobile app can use this status to update the user's assistance card in real time.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `assistance:update_status`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:update_status
// @Description
// @Description ALLOWED STATUS VALUES:
// @Description - pending: Request has been submitted but not accepted yet.
// @Description - accepted: Admin/responder has accepted the request.
// @Description - en_route: Help is on the way.
// @Description - arrived: Responder has arrived.
// @Description - completed: Assistance has been completed.
// @Description - cancelled: Request was cancelled.
// @Description - rejected: Request was rejected.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "status": "en_route"
// @Description }
// @Tags Admin Assistance
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "Assistance request ID. This is the MongoDB ObjectID of the assistance request." example(66e19b71c8f2a2b4d1234567)
// @Param request body dto.UpdateAssistanceStatusRequest true "Status update payload. status is required."
// @Success 200 {object} map[string]interface{} "Assistance status updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid status."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Assistance request not found."
// @Failure 500 {object} map[string]interface{} "Server error while updating assistance status."
// @Router /admin/assistance/{id}/status [put]
func (h *AssistanceHandler) UpdateAssistanceStatus(c *gin.Context) {
	requestID := c.Param("id")

	var req dto.UpdateAssistanceStatusRequest

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

	request, err := h.assistanceService.UpdateAssistanceStatus(
		requestID,
		req.Status,
	)
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
		"Assistance status updated successfully",
		request,
	)
}
