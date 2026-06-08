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
	validator          *validator.Validate
}

func NewAssistanceHandler(
	assistanceService *services.AssistanceService,
) *AssistanceHandler {
	return &AssistanceHandler{
		assistanceService: assistanceService,
		validator:          validator.New(),
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
// @Summary Get all assistance requests
// @Description Admin dashboard fetches all user assistance requests.
// @Tags Admin Assistance
// @Security AdminApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
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
// @Summary Get assistance request by ID
// @Description Admin dashboard fetches one assistance request.
// @Tags Admin Assistance
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Assistance Request ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
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