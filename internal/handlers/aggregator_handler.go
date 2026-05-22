package handlers

import (
	"log"
	"net/http"
	"sync/atomic"

	"disaster_alert_backend/internal/aggregator"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AggregatorHandler struct {
	alertAggregator *aggregator.AlertAggregator
	isSyncing       atomic.Bool
}

func NewAggregatorHandler(alertAggregator *aggregator.AlertAggregator) *AggregatorHandler {
	return &AggregatorHandler{
		alertAggregator: alertAggregator,
	}
}

func (h *AggregatorHandler) SyncExternalAlerts(c *gin.Context) {
	if !h.isSyncing.CompareAndSwap(false, true) {
		utils.SuccessResponse(
			c,
			http.StatusAccepted,
			"External alert sync is already running",
			gin.H{
				"status": "already_running",
			},
		)
		return
	}

	go func() {
		defer h.isSyncing.Store(false)

		log.Println("Manual external alert sync started in background")

		if err := h.alertAggregator.SyncExternalAlerts(); err != nil {
			log.Println("Manual external alert sync failed:", err)
			return
		}

		log.Println("Manual external alert sync completed successfully")
	}()

	utils.SuccessResponse(
		c,
		http.StatusAccepted,
		"External alert sync started in background",
		gin.H{
			"status": "started",
		},
	)
}