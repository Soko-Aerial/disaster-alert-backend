package handlers

import (
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

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

// CreateAlert godoc
// @Summary Create alert
// @Description Admin dashboard creates a manual alert.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Create alert request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/alerts [post]
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

// GetAlerts godoc
// @Summary Get all alerts
// @Description Admin dashboard fetches all alerts.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /admin/alerts [get]
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
	limit := parseAlertLimit(c, 50, 200)

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

func (h *AlertHandler) GetLocalAlerts(c *gin.Context) {
	country := c.Query("country")
	limit := parseAlertLimit(c, 50, 200)

	if country == "" {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Country is required for local alerts",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetLocalAlerts(country, limit)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch local alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Local alerts fetched successfully",
		alerts,
	)
}

func (h *AlertHandler) GetGlobalAlerts(c *gin.Context) {
	country := c.Query("country")
	limit := parseAlertLimit(c, 50, 200)

	alerts, err := h.alertService.GetGlobalAlerts(country, limit)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch global alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Global alerts fetched successfully",
		alerts,
	)
}

func (h *AlertHandler) GetWeatherAlerts(c *gin.Context) {
	country := c.Query("country")
	limit := parseAlertLimit(c, 50, 200)

	alerts, err := h.alertService.GetWeatherAlerts(country, limit)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch weather alerts",
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

func (h *AlertHandler) GetHealthAlerts(c *gin.Context) {
	country := c.Query("country")
	limit := parseAlertLimit(c, 50, 200)

	alerts, err := h.alertService.GetHealthAlerts(country, limit)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch health alerts",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Health alerts fetched successfully",
		alerts,
	)
}

// GetAlertByID godoc
// @Summary Get alert by ID
// @Description Admin dashboard fetches one alert.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/alerts/{id} [get]
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

// UpdateAlertStatus godoc
// @Summary Update alert status
// @Description Admin dashboard updates an alert status.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body map[string]string true "Status update body"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/alerts/{id}/status [put]
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

// DeleteAlert godoc
// @Summary Delete alert
// @Description Admin dashboard deletes an alert.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /admin/alerts/{id} [delete]
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

func parseAlertLimit(c *gin.Context, defaultLimit int, maxLimit int) int {
	limit := defaultLimit

	limitQuery := c.Query("limit")
	if limitQuery != "" {
		parsedLimit, err := strconv.Atoi(limitQuery)
		if err == nil && parsedLimit > 0 {
			if parsedLimit > maxLimit {
				parsedLimit = maxLimit
			}

			limit = parsedLimit
		}
	}

	return limit
}