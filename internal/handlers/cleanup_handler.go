package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type CleanupHandler struct {
	alertRepository *repositories.AlertRepository
}

func NewCleanupHandler(alertRepository *repositories.AlertRepository) *CleanupHandler {
	return &CleanupHandler{
		alertRepository: alertRepository,
	}
}

// CleanupExpiredExternalAlerts godoc
// @Summary Cleanup expired external alerts
// @Description Deactivates expired external alerts from configured alert sources.
// @Description
// @Description This is an admin maintenance operation used to keep the alert feed clean by marking expired external alerts as inactive.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description alerts:update
// @Tags Admin Maintenance
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Expired external alerts cleaned up successfully"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 500 {object} map[string]interface{} "Failed to cleanup expired external alerts"
// @Router /admin/maintenance/cleanup-expired-alerts [post]
func (h *CleanupHandler) CleanupExpiredExternalAlerts(c *gin.Context) {
	count, err := h.alertRepository.DeactivateExpiredExternalAlerts()
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to cleanup expired external alerts: "+err.Error(),
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Expired external alerts cleaned up successfully",
		gin.H{
			"deactivatedCount": count,
		},
	)
}
