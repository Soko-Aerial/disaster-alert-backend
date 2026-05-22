package handlers

import (
	"net/http"

	"disaster_alert_backend/config"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ExternalSourceHandler struct {
	cfg *config.Config
}

func NewExternalSourceHandler(cfg *config.Config) *ExternalSourceHandler {
	return &ExternalSourceHandler{
		cfg: cfg,
	}
}

func (h *ExternalSourceHandler) GetExternalSourceStatus(c *gin.Context) {
	data := gin.H{
		"gdacs": gin.H{
			"enabled": isSourceEnabled(h.cfg.GDACSEnabled),
			"type":    "background_sync",
		},
		"nasa_firms": gin.H{
			"enabled":  isSourceEnabled(h.cfg.NASAFIRMSEnabled),
			"type":     "background_sync",
			"source":   h.cfg.NASAFIRMSSource,
			"area":     h.cfg.NASAFIRMSArea,
			"dayRange": h.cfg.NASAFIRMSDayRange,
			"limit":    h.cfg.NASAFIRMSLimit,
		},
		"gdelt": gin.H{
			"enabled":  isSourceEnabled(h.cfg.GDELTEnabled),
			"type":     "background_sync",
			"limit":    h.cfg.GDELTLimit,
			"timespan": h.cfg.GDELTTimespan,
		},
		"reliefweb": gin.H{
			"enabled": isSourceEnabled(h.cfg.ReliefWebEnabled),
			"type":    "background_sync",
			"limit":   h.cfg.ReliefWebLimit,
			"reason":  "Requires approved ReliefWeb appname",
		},
		"openweather": gin.H{
			"enabled": isSourceEnabled(h.cfg.OpenWeatherEnabled),
			"type":    "user_selected_location",
			"mode":    "current_weather_advisory",
		},
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"External source status fetched successfully",
		data,
	)
}

func isSourceEnabled(value string) bool {
	switch value {
	case "true", "TRUE", "True", "1", "yes", "YES", "enabled", "ENABLED":
		return true
	default:
		return false
	}
}