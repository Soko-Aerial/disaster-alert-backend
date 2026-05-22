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