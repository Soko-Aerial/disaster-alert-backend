package handlers

import (
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"
	"disaster_alert_backend/internal/repositories"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AlertHandler struct {
	alertService *services.AlertService
	validator    *validator.Validate
}

func NewAlertHandler(alertService *services.AlertService) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		validator:    validator.New(),
	}
}

func (h *AlertHandler) CreateAlert(c *gin.Context) {
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

	var req dto.CreateAlertRequest

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

	alert, err := h.alertService.CreateAlert(userID, req)
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
		"Alert created successfully",
		alert,
	)
}

func (h *AlertHandler) GetAlerts(c *gin.Context) {
	alerts, err := h.alertService.GetAlerts()
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch alerts",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alerts fetched successfully",
		alerts,
	)
}

func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	limit := 50

	limitQuery := c.Query("limit")
	if limitQuery != "" {
		parsedLimit, err := strconv.Atoi(limitQuery)
		if err != nil || parsedLimit <= 0 {
			utils.ErrorResponse(
				c,
				http.StatusBadRequest,
				"Invalid limit value",
				err,
			)
			return
		}

		if parsedLimit > 200 {
			parsedLimit = 200
		}

		limit = parsedLimit
	}

	filter := repositories.AlertFilter{
		Category:   c.Query("category"),
		Severity:   c.Query("severity"),
		SourceName: c.Query("sourceName"),
		SourceType: c.Query("sourceType"),
		Country:    c.Query("country"),
		Limit:      limit,
	}

	alerts, err := h.alertService.GetActiveAlertsWithFilters(filter)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch active alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Active alerts fetched successfully",
		alerts,
	)
}

func (h *AlertHandler) GetAlertByID(c *gin.Context) {
	alertID := c.Param("id")

	alert, err := h.alertService.GetAlertByID(alertID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusNotFound,
			"Alert not found",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alert fetched successfully",
		alert,
	)
}

func (h *AlertHandler) UpdateAlertStatus(c *gin.Context) {
	alertID := c.Param("id")

	var req dto.UpdateAlertStatusRequest

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

	alert, err := h.alertService.UpdateAlertStatus(alertID, req.Status)
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
		"Alert status updated successfully",
		alert,
	)
}

func (h *AlertHandler) DeleteAlert(c *gin.Context) {
	alertID := c.Param("id")

	err := h.alertService.DeleteAlert(alertID)
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
		"Alert deleted successfully",
		nil,
	)
}

func (h *AlertHandler) GetNearbyAlerts(c *gin.Context) {
	latQuery := c.Query("lat")
	lngQuery := c.Query("lng")
	radiusQuery := c.DefaultQuery("radiusKm", "100")
	limitQuery := c.DefaultQuery("limit", "50")

	lat, err := strconv.ParseFloat(latQuery, 64)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid latitude value",
			err,
		)
		return
	}

	lng, err := strconv.ParseFloat(lngQuery, 64)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid longitude value",
			err,
		)
		return
	}

	radiusKm, err := strconv.ParseFloat(radiusQuery, 64)
	if err != nil || radiusKm <= 0 {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid radiusKm value",
			err,
		)
		return
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit <= 0 {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid limit value",
			err,
		)
		return
	}

	if limit > 200 {
		limit = 200
	}

	filter := repositories.AlertFilter{
		Category:   c.Query("category"),
		Severity:   c.Query("severity"),
		SourceName: c.Query("sourceName"),
		SourceType: c.Query("sourceType"),
		Country:    c.Query("country"),
		Limit:      limit,
	}

	alerts, err := h.alertService.GetNearbyAlertsWithFilters(
		lat,
		lng,
		radiusKm,
		filter,
	)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch nearby alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Nearby alerts fetched successfully",
		alerts,
	)
}
func (h *AlertHandler) GetCriticalGlobalAlerts(c *gin.Context) {
	limit := 20

	limitQuery := c.Query("limit")
	if limitQuery != "" {
		parsedLimit, err := strconv.Atoi(limitQuery)
		if err != nil || parsedLimit <= 0 {
			utils.ErrorResponse(
				c,
				http.StatusBadRequest,
				"Invalid limit value",
				err,
			)
			return
		}

		if parsedLimit > 100 {
			parsedLimit = 100
		}

		limit = parsedLimit
	}

	alerts, err := h.alertService.GetCriticalGlobalAlerts(limit)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch critical global alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Critical global alerts fetched successfully",
		alerts,
	)
}
