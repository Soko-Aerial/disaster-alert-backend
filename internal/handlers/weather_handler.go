package handlers

import (
	"log"
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type WeatherHandler struct {
	weatherService *services.WeatherService
}

func NewWeatherHandler(weatherService *services.WeatherService) *WeatherHandler {
	return &WeatherHandler{
		weatherService: weatherService,
	}
}

func (h *WeatherHandler) GetWeatherAlerts(c *gin.Context) {
	latQuery := c.Query("lat")
	lngQuery := c.Query("lng")
	name := c.DefaultQuery("name", "Selected Location")
	country := c.DefaultQuery("country", "")

	lat, err := strconv.ParseFloat(latQuery, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid latitude value", err)
		return
	}

	lng, err := strconv.ParseFloat(lngQuery, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid longitude value", err)
		return
	}

	alerts, err := h.weatherService.FetchWeatherAlertsForLocation(
		lat,
		lng,
		name,
		country,
	)
	if err != nil {
		log.Println("Weather alert fetch error:", err)

		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch weather alerts: "+err.Error(),
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Weather alerts fetched successfully",
		alerts,
	)
}