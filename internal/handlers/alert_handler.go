package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"disaster_alert_backend/internal/authz"
	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
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
// @Description HOW ADMIN ALERT ROUTING WORKS:
// @Description Admin-created alerts support auto-routing.
// @Description The backend checks the alert category and automatically fills access-control routing fields unless the admin provides routing overrides.
// @Description
// @Description CATEGORY ROUTING EXAMPLES:
// @Description - fire -> Fire Service
// @Description - flood/weather/drought/earthquake -> Disaster management agency
// @Description - health/medical -> Health or ambulance service
// @Description - robbery/security/protests -> Police/security agency
// @Description - accident -> Police and ambulance
// @Description - conflict/munitions -> National security, police, and armed forces
// @Description - galamsey -> Police, minerals/mining authority, and national security
// @Description - other -> system/manual review
// @Description
// @Description ADMIN OVERRIDE:
// @Description For manual alerts, an admin can optionally provide ownerOrganisationId, leadOrganisationId, assignedOrgIds, and visibleToOrgIds.
// @Description If those fields are omitted, the backend uses the default auto-routing rules.
// @Description
// @Description TECHNICAL ACCESS RULE:
// @Description Read, update, and delete actions are protected by both permission checks and record-level scope checks.
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

// PreviewAlertTargeting godoc
// @Summary Preview alert recipients before sending
// @Description Calculates how many users would receive a push notification before the admin sends the alert.
// @Description
// @Description WHY THIS ENDPOINT EXISTS:
// @Description It prevents accidental mass panic.
// @Description The admin can draw/select an affected area on the map and preview the number of users who will receive the alert.
// @Description
// @Description TARGETING MODES:
// @Description radius - Notify users within a radius around latitude/longitude.
// @Description polygon - Notify users inside a drawn map area.
// @Description region - Notify users in a selected region.
// @Description country - Notify users in a selected country.
// @Description national - Restricted national alert preview.
// @Description
// @Description ZONES:
// @Description Danger zone users receive the urgent alert.
// @Description Awareness zone users receive softer nearby-warning messaging.
// @Description
// @Description IMPORTANT:
// @Description This endpoint does not send push notifications.
// @Description It only previews recipient counts.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param request body dto.AlertTargetingPreviewRequest true "Alert targeting preview payload"
// @Success 200 {object} dto.AlertTargetingPreviewResponse "Alert targeting preview calculated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or targeting rules"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 500 {object} map[string]interface{} "Failed to preview alert targeting"
// @Router /admin/alerts/preview-targeting [post]
func (h *AlertHandler) PreviewAlertTargeting(c *gin.Context) {
	var req dto.AlertTargetingPreviewRequest

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

	preview, err := h.alertService.PreviewAlertTargeting(req)
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
		"Alert targeting preview calculated successfully",
		preview,
	)
}

// EscalateAlert godoc
// @Summary Escalate an alert
// @Description Escalates an existing alert when its affected area expands or severity becomes critical.
// @Description
// @Description ESCALATION RULES:
// @Description If the alert area expands, only newly affected users receive a push notification.
// @Description If severity is upgraded to critical, previous affected users and newly affected users receive an update.
// @Description Duplicate delivery is blocked using alert delivery history.
// @Description
// @Description NATIONAL ALERT SAFETY:
// @Description If targeting.mode is national, the privilege code must include alerts:send_national.
// @Description confirmNationalAlert must be true and nationalAlertReason must be provided.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "Alert ID"
// @Param request body dto.AlertEscalationRequest true "Alert escalation payload"
// @Success 200 {object} map[string]interface{} "Alert escalated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or escalation rules"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "Alert not found"
// @Router /admin/alerts/{id}/escalate [put]
func (h *AlertHandler) EscalateAlert(c *gin.Context) {
	alertID := c.Param("id")

	var req dto.AlertEscalationRequest

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

	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alert, err := h.alertService.EscalateAlertForPrivilege(
		alertID,
		req,
		privilegeCtx,
	)
	if err != nil {
		statusCode := http.StatusBadRequest

		switch err.Error() {
		case "alert not found":
			statusCode = http.StatusNotFound
		case "you do not have access to escalate this alert",
			"national alerts require alerts:send_national permission":
			statusCode = http.StatusForbidden
		}

		utils.ErrorResponse(
			c,
			statusCode,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alert escalated successfully",
		alert,
	)
}

// GetAlerts godoc
// @Summary List alerts
// @Description Returns alerts from the system for authenticated mobile users.
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

// AdminGetAlerts godoc
// @Summary List admin alerts
// @Description Returns alerts for the admin dashboard.
// @Description
// @Description ACCESS CONTROL:
// @Description Super admin/global privilege codes can see all alerts.
// @Description Organisation privilege codes only see alerts allowed by their category grants and record scope.
// @Description Optional filters: category, severity, status, sourceName, sourceType, country, and limit.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param category query string false "Filter by category." example(fire)
// @Param severity query string false "Filter by severity." Enums(low, medium, high, critical) example(high)
// @Param status query string false "Filter by status." example(active)
// @Param sourceName query string false "Filter by source name." example(manual)
// @Param sourceType query string false "Filter by source type." example(internal)
// @Param country query string false "Filter by country name or country code." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 200, maximum is 500." example(200)
// @Success 200 {object} map[string]interface{} "Alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch alerts."
// @Router /admin/alerts [get]
func (h *AlertHandler) AdminGetAlerts(c *gin.Context) {
	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 200, 500)
	alerts = filterAdminAlertsFromQuery(c, alerts, "", false, false, false, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Alerts fetched successfully",
		alerts,
	)
}

// AdminGetActiveAlerts godoc
// @Summary List active admin alerts
// @Description Returns currently active alerts for the admin dashboard.
// @Description Super admin can see all active alerts. Organisation privilege codes only see alerts inside their permitted scope.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param category query string false "Filter by category." example(flood)
// @Param severity query string false "Filter by severity." Enums(low, medium, high, critical) example(high)
// @Param sourceName query string false "Filter by source name." example(manual)
// @Param sourceType query string false "Filter by source type." example(internal)
// @Param country query string false "Filter by country name or country code." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Active alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch active alerts."
// @Router /admin/alerts/active [get]
func (h *AlertHandler) AdminGetActiveAlerts(c *gin.Context) {
	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch active alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 50, 200)
	alerts = filterAdminAlertsFromQuery(c, alerts, "", true, false, false, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Active alerts fetched successfully",
		alerts,
	)
}

// AdminGetLocalAlerts godoc
// @Summary List local admin alerts
// @Description Fetches active local alerts for the admin dashboard using a country filter.
// @Description The country can be a country name or Alpha-2 country code.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param country query string true "Country name or country code used to filter local alerts." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Local alerts fetched successfully."
// @Failure 400 {object} map[string]interface{} "Country query parameter is required."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch local alerts."
// @Router /admin/alerts/local [get]
func (h *AlertHandler) AdminGetLocalAlerts(c *gin.Context) {
	country := strings.TrimSpace(c.Query("country"))
	if country == "" {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Country is required for local alerts",
			nil,
		)
		return
	}

	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch local alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 50, 200)
	alerts = filterAdminAlertsFromQuery(c, alerts, "", true, true, false, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Local alerts fetched successfully",
		alerts,
	)
}

// AdminGetGlobalAlerts godoc
// @Summary List global admin alerts
// @Description Fetches active global alerts for the admin dashboard.
// @Description If country is supplied, alerts from that country are excluded. If country is omitted, all active scoped alerts are returned.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param country query string false "Optional country name or country code to exclude from global view." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Global alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch global alerts."
// @Router /admin/alerts/global [get]
func (h *AlertHandler) AdminGetGlobalAlerts(c *gin.Context) {
	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch global alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 50, 200)
	alerts = filterAdminAlertsFromQuery(c, alerts, "", true, false, true, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Global alerts fetched successfully",
		alerts,
	)
}

// AdminGetWeatherAlerts godoc
// @Summary List weather admin alerts
// @Description Fetches active weather-related alerts for the admin dashboard.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param country query string false "Optional country filter." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Weather alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch weather alerts."
// @Router /admin/alerts/weather [get]
func (h *AlertHandler) AdminGetWeatherAlerts(c *gin.Context) {
	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch weather alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 50, 200)
	alerts = filterAdminAlertsFromQuery(c, alerts, "weather", true, false, false, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Weather alerts fetched successfully",
		alerts,
	)
}

// AdminGetHealthAlerts godoc
// @Summary List health admin alerts
// @Description Fetches active health-related alerts for the admin dashboard.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param country query string false "Optional country filter." example(Ghana)
// @Param limit query int false "Maximum number of alerts to return. Default is 50, maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Health alerts fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch health alerts."
// @Router /admin/alerts/health [get]
func (h *AlertHandler) AdminGetHealthAlerts(c *gin.Context) {
	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alerts, err := h.alertService.GetAlertsForPrivilege(privilegeCtx)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch health alerts",
			err.Error(),
		)
		return
	}

	limit := parseAlertLimit(c, 50, 200)
	alerts = filterAdminAlertsFromQuery(c, alerts, "health", true, false, false, limit)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Health alerts fetched successfully",
		alerts,
	)
}

// GetActiveAlerts godoc
// @Summary List active alerts
// @Description Returns currently active alerts for authenticated mobile users.
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Param category query string false "Filter by category." example(flood)
// @Param severity query string false "Filter by severity." Enums(low, medium, high, critical) example(high)
// @Param sourceName query string false "Filter by source name." example(manual)
// @Param sourceType query string false "Filter by source type." example(internal)
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

// GetAlertDeliveryHistory godoc
// @Summary Get alert delivery history
// @Description Returns delivery batch history for an alert.
// @Description Shows how many users were targeted, sent, failed, or skipped.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Success 200 {object} map[string]interface{} "Alert delivery history fetched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid alert ID"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code"
// @Router /admin/alerts/{id}/delivery-history [get]
func (h *AlertHandler) GetAlertDeliveryHistory(c *gin.Context) {
	alertID := c.Param("id")

	if privilegeCtx, ok := authz.FromGin(c); ok {
		if _, err := h.alertService.GetAlertByIDForPrivilege(alertID, privilegeCtx); err != nil {
			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				"Alert not found or access denied",
				err.Error(),
			)
			return
		}
	}

	history, err := h.alertService.GetAlertDeliveryHistory(alertID)
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
		"Alert delivery history fetched successfully",
		history,
	)
}

// GetAlertRecipientDeliveries godoc
// @Summary Get alert recipient delivery records
// @Description Returns recipient-level delivery records for an alert.
// @Description This should be restricted because it exposes user-level notification delivery data.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Alert ID"
// @Param limit query int false "Maximum number of recipient records to return"
// @Success 200 {object} map[string]interface{} "Alert recipient deliveries fetched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid alert ID"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Unauthorized privilege code"
// @Router /admin/alerts/{id}/recipient-deliveries [get]
func (h *AlertHandler) GetAlertRecipientDeliveries(c *gin.Context) {
	alertID := c.Param("id")

	if privilegeCtx, ok := authz.FromGin(c); ok {
		if _, err := h.alertService.GetAlertByIDForPrivilege(alertID, privilegeCtx); err != nil {
			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				"Alert not found or access denied",
				err.Error(),
			)
			return
		}
	}

	limit := int64(500)

	if value := c.Query("limit"); value != "" {
		parsedLimit, err := strconv.ParseInt(value, 10, 64)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	deliveries, err := h.alertService.GetAlertRecipientDeliveries(alertID, limit)
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
		"Alert recipient deliveries fetched successfully",
		deliveries,
	)
}

// GetLocalAlerts godoc
// @Summary List local alerts
// @Description Fetches alerts relevant to a specific country for authenticated mobile users.
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

	if strings.TrimSpace(country) == "" {
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
// @Description Fetches global alerts for authenticated mobile users.
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
// @Description Fetches weather-related alerts for authenticated mobile users.
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
// @Description Fetches health-related alerts for authenticated mobile users.
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

// AdminGetAlertByID godoc
// @Summary Get admin alert by ID
// @Description Fetches one alert for the admin dashboard.
// @Description
// @Description ACCESS CONTROL:
// @Description The privilege code must have alerts:read permission.
// @Description The alert must also be inside the organisation/category scope unless the privilege code has global access.
// @Tags Admin Alerts
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Alert ID. This is the MongoDB ObjectID of the alert." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Alert fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid alert ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, unauthorized privilege code, or alert outside organisation scope."
// @Failure 404 {object} map[string]interface{} "Alert not found."
// @Router /admin/alerts/{id} [get]
func (h *AlertHandler) AdminGetAlertByID(c *gin.Context) {
	alertID := c.Param("id")

	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alert, err := h.alertService.GetAlertByIDForPrivilege(alertID, privilegeCtx)
	if err != nil {
		statusCode := http.StatusNotFound

		if strings.Contains(strings.ToLower(err.Error()), "access") ||
			strings.Contains(strings.ToLower(err.Error()), "permission") {
			statusCode = http.StatusForbidden
		}

		utils.ErrorResponse(
			c,
			statusCode,
			"Alert not found or access denied",
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

// GetAlertByID godoc
// @Summary Get one alert
// @Description Fetches details of a single alert by ID for authenticated mobile users.
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

	var alert *models.Alert
	var err error

	if privilegeCtx, ok := authz.FromGin(c); ok {
		alert, err = h.alertService.GetAlertByIDForPrivilege(alertID, privilegeCtx)
	} else {
		alert, err = h.alertService.GetAlertByID(alertID)
	}

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

// AdminUpdateAlertStatus godoc
// @Summary Update admin alert status
// @Description Updates only the status of an alert from the admin dashboard.
// @Description This does not fully edit the alert content. It only changes status such as draft, active, resolved, expired, or cancelled.
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
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, unauthorized privilege code, or alert outside organisation scope."
// @Failure 404 {object} map[string]interface{} "Alert not found."
// @Failure 500 {object} map[string]interface{} "Server error while updating alert status."
// @Router /admin/alerts/{id}/status [put]
func (h *AlertHandler) AdminUpdateAlertStatus(c *gin.Context) {
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

	privilegeCtx, ok := authz.FromGin(c)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Privilege context not found",
			nil,
		)
		return
	}

	alert, err := h.alertService.UpdateAlertStatusForPrivilege(
		alertID,
		req.Status,
		privilegeCtx,
	)
	if err != nil {
		statusCode := http.StatusBadRequest

		if strings.Contains(strings.ToLower(err.Error()), "access") ||
			strings.Contains(strings.ToLower(err.Error()), "permission") {
			statusCode = http.StatusForbidden
		}

		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			statusCode = http.StatusNotFound
		}

		utils.ErrorResponse(
			c,
			statusCode,
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
// @Description IMPORTANT:
// @Description For normal emergency operations, resolving or expiring an alert is usually better than deleting it.
// @Description Delete should be used carefully because it may remove the alert from history depending on repository logic.
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

	var err error

	if privilegeCtx, ok := authz.FromGin(c); ok {
		err = h.alertService.DeleteAlertForPrivilege(alertID, privilegeCtx)
	} else {
		err = h.alertService.DeleteAlert(alertID)
	}

	if err != nil {
		statusCode := http.StatusBadRequest

		if strings.Contains(strings.ToLower(err.Error()), "do not have access") ||
			strings.Contains(strings.ToLower(err.Error()), "permission") {
			statusCode = http.StatusForbidden
		}

		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			statusCode = http.StatusNotFound
		}

		utils.ErrorResponse(
			c,
			statusCode,
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

func filterAdminAlertsFromQuery(
	c *gin.Context,
	alerts []models.Alert,
	forcedCategory string,
	activeOnly bool,
	requireCountryMatch bool,
	excludeCountry bool,
	limit int,
) []models.Alert {
	category := strings.TrimSpace(c.Query("category"))
	if forcedCategory != "" {
		category = forcedCategory
	}

	severity := strings.TrimSpace(c.Query("severity"))
	status := strings.TrimSpace(c.Query("status"))
	sourceName := strings.TrimSpace(c.Query("sourceName"))
	sourceType := strings.TrimSpace(c.Query("sourceType"))
	country := strings.TrimSpace(c.Query("country"))

	filtered := make([]models.Alert, 0, len(alerts))

	for _, alert := range alerts {
		if activeOnly && !strings.EqualFold(strings.TrimSpace(alert.Status), "active") {
			continue
		}

		if category != "" && !strings.EqualFold(strings.TrimSpace(alert.Category), category) {
			continue
		}

		if severity != "" && !strings.EqualFold(strings.TrimSpace(alert.Severity), severity) {
			continue
		}

		if status != "" && !strings.EqualFold(strings.TrimSpace(alert.Status), status) {
			continue
		}

		if sourceName != "" && !strings.EqualFold(strings.TrimSpace(alert.SourceName), sourceName) {
			continue
		}

		if sourceType != "" && !strings.EqualFold(strings.TrimSpace(alert.SourceType), sourceType) {
			continue
		}

		if requireCountryMatch && country != "" && !alertCountryMatches(alert, country) {
			continue
		}

		if !requireCountryMatch && country != "" && !excludeCountry && !alertCountryMatches(alert, country) {
			continue
		}

		if excludeCountry && country != "" && alertCountryMatches(alert, country) {
			continue
		}

		filtered = append(filtered, alert)

		if limit > 0 && len(filtered) >= limit {
			break
		}
	}

	return filtered
}

func alertCountryMatches(alert models.Alert, country string) bool {
	queryCountry := utils.NormalizeCountryCode(country)
	alertCountry := utils.NormalizeCountryCode(alert.Location.Country)

	if alertCountry == "" {
		alertCountry = utils.NormalizeCountryCode(alert.Targeting.Country)
	}

	if queryCountry != "" && alertCountry != "" {
		return queryCountry == alertCountry
	}

	return strings.EqualFold(
		strings.TrimSpace(country),
		strings.TrimSpace(alert.Location.Country),
	) || strings.EqualFold(
		strings.TrimSpace(country),
		strings.TrimSpace(alert.Targeting.Country),
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
