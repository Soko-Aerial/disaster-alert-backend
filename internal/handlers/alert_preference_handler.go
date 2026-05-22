package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AlertPreferenceHandler struct {
	preferenceService *services.AlertPreferenceService
}

func NewAlertPreferenceHandler(
	preferenceService *services.AlertPreferenceService,
) *AlertPreferenceHandler {
	return &AlertPreferenceHandler{
		preferenceService: preferenceService,
	}
}

func (h *AlertPreferenceHandler) GetPreferences(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	preferences, err := h.preferenceService.GetUserPreferences(userID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch alert preferences",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alert preferences fetched successfully",
		preferences,
	)
}

func (h *AlertPreferenceHandler) UpdatePreferences(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.UpdateAlertPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	preferences, err := h.preferenceService.UpdateUserPreferences(userID, req)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to update alert preferences",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alert preferences updated successfully",
		preferences,
	)
}