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

// GetPreferences godoc
// @Summary Get alert preferences
// @Description Returns the authenticated user's alert category preferences.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the app opens the user's alert settings/preferences screen.
// @Tags User Settings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Alert preferences fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch alert preferences."
// @Router /user/alert-preferences [get]
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

// UpdatePreferences godoc
// @Summary Update alert preferences
// @Description Updates the authenticated user's alert category preferences.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the user enables or disables alert categories in the app settings.
// @Description
// @Description IMPORTANT:
// @Description All fields are optional. Send only the categories you want to change.
// @Description Use true to enable a category and false to disable it.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "fire": true,
// @Description   "flood": true,
// @Description   "weather": true,
// @Description   "health": true,
// @Description   "criticalAlerts": true
// @Description }
// @Tags User Settings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateAlertPreferenceRequest true "Alert preference update payload. All fields are optional."
// @Success 200 {object} map[string]interface{} "Alert preferences updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to update alert preferences."
// @Router /user/alert-preferences [put]
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
