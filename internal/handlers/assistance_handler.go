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
// @Summary List assistance requests for admin
// @Description Returns all assistance requests submitted by users.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:read
// @Description
// @Description SWAGGER TESTING:
// @Description 1. Click Authorize.
// @Description 2. Enter Sigtrack-Admin-API-Key under AdminApiKeyAuth.
// @Description 3. Enter the generated full UUID under PrivilegeCodeAuth.
// @Description 4. Execute this endpoint.
// @Tags Admin Assistance
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Assistance requests fetched successfully"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 500 {object} map[string]interface{} "Failed to fetch assistance requests"
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
// @Summary Get one assistance request for admin
// @Description Returns the full details of a single assistance request.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:read
// @Tags Admin Assistance
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Assistance Request ID"
// @Success 200 {object} map[string]interface{} "Assistance request fetched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid assistance request ID"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "Assistance request not found"
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
// @Description Use this endpoint when an admin accepts a request, dispatches help, marks responders as en route, confirms arrival, completes the request, cancels it, or rejects it.
// @Description The mobile app can use this status to update the user's assistance card in real time.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description assistance:update_status
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {"status":"en_route"}
// @Description
// @Description COMMON STATUS VALUES:
// @Description pending, accepted, en_route, arrived, completed, cancelled, rejected
// @Tags Admin Assistance
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "Assistance Request ID"
// @Param request body dto.UpdateAssistanceStatusRequest true "Status update payload"
// @Success 200 {object} map[string]interface{} "Assistance status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid status"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "Assistance request not found"
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
