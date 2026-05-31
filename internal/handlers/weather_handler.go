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
	lat, lng, ok := parseWeatherCoordinates(c)
	if !ok {
		return
	}

	name := c.DefaultQuery("name", "Selected Location")
	country := c.DefaultQuery("country", "")

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

func (h *WeatherHandler) GetCurrentWeather(c *gin.Context) {
	lat, lng, ok := parseWeatherCoordinates(c)
	if !ok {
		return
	}

	country := c.DefaultQuery("country", "")

	weather, err := h.weatherService.FetchCurrentWeather(
		lat,
		lng,
		country,
	)
	if err != nil {
		log.Println("Current weather fetch error:", err)

		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch current weather: "+err.Error(),
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Current weather fetched successfully",
		weather,
	)
}

func (h *WeatherHandler) GetWeatherForecast(c *gin.Context) {
	lat, lng, ok := parseWeatherCoordinates(c)
	if !ok {
		return
	}

	country := c.DefaultQuery("country", "")

	forecast, err := h.weatherService.FetchWeatherForecast(
		lat,
		lng,
		country,
	)
	if err != nil {
		log.Println("Weather forecast fetch error:", err)

		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch weather forecast: "+err.Error(),
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Weather forecast fetched successfully",
		forecast,
	)
}

func parseWeatherCoordinates(c *gin.Context) (float64, float64, bool) {
	latQuery := c.Query("lat")
	lngQuery := c.Query("lng")

	lat, err := strconv.ParseFloat(latQuery, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid latitude value", err)
		return 0, 0, false
	}

	lng, err := strconv.ParseFloat(lngQuery, 64)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid longitude value", err)
		return 0, 0, false
	}

	return lat, lng, true
}