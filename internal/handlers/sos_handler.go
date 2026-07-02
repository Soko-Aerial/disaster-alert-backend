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
// @Summary List SOS requests for admin
// @Description Returns all SOS emergency requests submitted by users.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:read
// @Tags Admin SOS
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "SOS requests fetched successfully"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 500 {object} map[string]interface{} "Failed to fetch SOS requests"
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
// @Summary Get one SOS request for admin
// @Description Returns details of a single SOS emergency request.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:read
// @Tags Admin SOS
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Param id path string true "SOS ID"
// @Success 200 {object} map[string]interface{} "SOS request fetched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid SOS ID"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "SOS request not found"
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
// @Summary Update SOS emergency status
// @Description Updates the status of an SOS emergency request from the admin dashboard.
// @Description
// @Description Use assigned when a responder/admin has taken the case, resolved when the emergency is handled, and cancelled when the request is cancelled.
// @Description The mobile app can stop SOS alarm/live tracking when status becomes resolved or cancelled.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description sos:update_status
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {"status":"assigned"}
// @Description
// @Description COMMON STATUS VALUES:
// @Description active, assigned, resolved, cancelled
// @Tags Admin SOS
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "SOS ID"
// @Param request body dto.UpdateSOSStatusRequest true "SOS status update payload"
// @Success 200 {object} map[string]interface{} "SOS status updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid status"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "SOS request not found"
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
