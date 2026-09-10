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
// @Summary Create a manual public alert
// @Description Creates a verified public alert from the admin dashboard.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin wants to manually publish a public safety alert such as flood warning, fire outbreak, weather danger, health risk, security threat, conflict warning, or other emergency.
// @Description
// @Description IMPORTANT:
// @Description This is for admin-created alerts.
// @Description User-submitted reports should normally be approved through /admin/reports/{id}/approve so that the report is converted into an alert.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `alerts:create`.
// @Description
// @Description REQUIRED FIELDS:
// @Description - title: Short headline users will see.
// @Description - description: Full alert details.
// @Description - category: Type of emergency.
// @Description - severity: low, medium, high, or critical.
// @Description - latitude: Alert latitude.
// @Description - longitude: Alert longitude.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - summary: Short preview text.
// @Description - status: draft, active, resolved, expired, or cancelled.
// @Description - address, country, region: Human-readable location details.
// @Description - radiusKm: Approximate affected radius in kilometers.
// @Description - safetyInstructions: List of instructions for users.
// @Description - imageUrls/videoUrls: Media links.
// @Description - eventTime/expiresAt: ISO date/time values.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "title": "Heavy rainfall warning",
// @Description   "description": "Heavy rainfall is expected in Accra with possible flooding in low-lying areas.",
// @Description   "summary": "Heavy rainfall expected in Accra with possible flooding.",
// @Description   "category": "weather",
// @Description   "severity": "high",
// @Description   "status": "active",
// @Description   "latitude": 5.6037,
// @Description   "longitude": -0.1870,
// @Description   "address": "Circle, Accra",
// @Description   "country": "Ghana",
// @Description   "region": "Greater Accra",
// @Description   "radiusKm": 10,
// @Description   "safetyInstructions": ["Move to higher ground", "Avoid flooded roads", "Follow official instructions"],
// @Description   "sourceType": "admin",
// @Description   "sourceName": "Admin Dashboard",
// @Description   "priorityScore": 85,
// @Description   "priorityLabel": "serious",
// @Description   "eventTime": "2026-09-10T08:30:00Z",
// @Description   "expiresAt": "2026-09-11T18:00:00Z",
// @Description   "confidence": 0.95
// @Description }
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAlertRequest true "Alert creation payload. title, description, category, severity, latitude, and longitude are required."
// @Success 201 {object} map[string]interface{} "Alert created successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, or alert could not be created."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Server error while creating alert."
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
// @Summary List alerts
// @Description Returns alerts from the system.
// @Description
// @Description ADMIN USE:
// @Description When called from /admin/alerts, this is used by the admin dashboard to list all alerts.
// @Description Requires AdminApiKeyAuth and PrivilegeCodeAuth with `alerts:read` permission.
// @Description
// @Description MOBILE USER USE:
// @Description When called from /alerts, this is used by the mobile app to show alerts to authenticated users.
// @Description Mobile users must provide BearerAuth.
// @Description
// @Description NOTE:
// @Description This handler is currently shared by both mobile and admin routes.
// @Description Make sure your service response hides admin-only/internal fields from normal mobile users if required.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch alerts."
// @Router /alerts [get]
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

// GetActiveAlerts godoc
// @Summary List active alerts
// @Description Returns currently active alerts.
// @Description
// @Description QUERY FILTERS:
// @Description - category: Optional category filter such as flood, weather, fire, health, security.
// @Description - severity: Optional severity filter such as low, medium, high, critical.
// @Description - sourceName: Optional source name filter.
// @Description - sourceType: Optional source type filter such as admin, weather, health, external.
// @Description - country: Optional country filter such as Ghana.
// @Description - limit: Optional result limit. Default is 50. Maximum is 200.
// @Description
// @Description MOBILE USE:
// @Description The mobile app can use this endpoint to show active alerts on the home screen, alert feed, and map.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param category query string false "Filter by category." example(flood)
// @Param severity query string false "Filter by severity." Enums(low, medium, high, critical) example(high)
// @Param sourceName query string false "Filter by source name." example(Admin Dashboard)
// @Param sourceType query string false "Filter by source type." Enums(internal, external, system, user_report, weather, health, news, admin) example(admin)
// @Param country query string false "Filter by country." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Active alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch active alerts."
// @Router /alerts/active [get]
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

// GetLocalAlerts godoc
// @Summary List local alerts
// @Description Fetches alerts relevant to a specific country.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the app or admin dashboard wants alerts affecting a specific country.
// @Description
// @Description EXAMPLE:
// @Description country=Ghana returns alerts affecting Ghana.
// @Description
// @Description REQUIRED QUERY PARAMETER:
// @Description - country: Country name used to filter local alerts.
// @Description
// @Description OPTIONAL QUERY PARAMETER:
// @Description - limit: Maximum number of alerts to return. Default is 50. Maximum is 200.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param country query string true "Country name used to filter local alerts." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Local alerts fetched successfully."
// @Failure 400 {object} map[string]interface{} "Country query parameter is required."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch local alerts."
// @Router /alerts/local [get]
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

// GetGlobalAlerts godoc
// @Summary List global alerts
// @Description Fetches global alerts.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the app or dashboard wants alerts that are not limited to one local country view.
// @Description
// @Description OPTIONAL QUERY PARAMETERS:
// @Description - country: Optional country context depending on service logic.
// @Description - limit: Maximum number of alerts to return. Default is 50. Maximum is 200.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param country query string false "Optional country context." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Global alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch global alerts."
// @Router /alerts/global [get]
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

// GetWeatherAlerts godoc
// @Summary List weather alerts
// @Description Fetches weather-related alerts.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this for weather warnings such as storms, heavy rainfall, flooding risk, heat, wind, or severe weather.
// @Description
// @Description OPTIONAL QUERY PARAMETERS:
// @Description - country: Optional country filter. Example: Ghana.
// @Description - limit: Maximum number of alerts to return. Default is 50. Maximum is 200.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param country query string false "Optional country filter." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Weather alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch weather alerts."
// @Router /alerts/weather [get]
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

// GetHealthAlerts godoc
// @Summary List health alerts
// @Description Fetches health-related alerts.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this for disease outbreaks, public health risks, contamination warnings, epidemic updates, or medical emergency alerts.
// @Description
// @Description OPTIONAL QUERY PARAMETERS:
// @Description - country: Optional country filter. Example: Ghana.
// @Description - limit: Maximum number of alerts to return. Default is 50. Maximum is 200.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param country query string false "Optional country filter." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Health alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch health alerts."
// @Router /alerts/health [get]
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
// @Summary Get one alert
// @Description Fetches details of a single alert by ID.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the mobile app or admin dashboard opens an alert detail screen.
// @Description
// @Description PATH PARAMETER:
// @Description - id: MongoDB ObjectID of the alert.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param id path string true "Alert ID. This is the MongoDB ObjectID of the alert." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Alert fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid alert ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid token."
// @Failure 404 {object} map[string]interface{} "Alert not found."
// @Router /alerts/{id} [get]
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
// @Description Updates the status of an alert from the admin dashboard.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin wants to mark an alert as draft, active, resolved, expired, or cancelled.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `alerts:update`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description alerts:update
// @Description
// @Description ALLOWED STATUS VALUES:
// @Description - draft: Alert is saved but not active.
// @Description - active: Alert is currently active.
// @Description - resolved: The emergency has been resolved.
// @Description - expired: The alert is no longer valid.
// @Description - cancelled: The alert was cancelled.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "status": "resolved"
// @Description }
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert ID. This is the MongoDB ObjectID of the alert." example(66e19b71c8f2a2b4d1234567)
// @Param request body dto.UpdateAlertStatusRequest true "Alert status update payload."
// @Success 200 {object} map[string]interface{} "Alert status updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid alert status."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Alert not found."
// @Failure 500 {object} map[string]interface{} "Server error while updating alert status."
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
// @Description Deletes an alert from the admin dashboard.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when an admin needs to permanently remove an alert.
// @Description
// @Description IMPORTANT:
// @Description For normal emergency operations, resolving or expiring an alert is usually better than deleting it.
// @Description Delete should be used carefully because it may remove the alert from history depending on repository logic.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `alerts:delete`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description alerts:delete
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Alert ID. This is the MongoDB ObjectID of the alert." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Alert deleted successfully."
// @Failure 400 {object} map[string]interface{} "Invalid alert ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Alert not found."
// @Failure 500 {object} map[string]interface{} "Server error while deleting alert."
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
