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