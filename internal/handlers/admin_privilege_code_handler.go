package handlers

import (
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AdminPrivilegeCodeHandler struct {
	service   *services.AdminPrivilegeCodeService
	validator *validator.Validate
}

func NewAdminPrivilegeCodeHandler(
	service *services.AdminPrivilegeCodeService,
) *AdminPrivilegeCodeHandler {
	return &AdminPrivilegeCodeHandler{
		service:   service,
		validator: validator.New(),
	}
}

// CreatePrivilegeCode godoc
// @Summary Generate a new admin privilege UUID
// @Description Creates a new privilege code for an organisation, unit, department, or admin level.
// @Description
// @Description This endpoint is used to generate the UUID that will later be sent in the X-Privilege-Code header.
// @Description It requires only the Admin API Key because this endpoint is used to create privilege codes.
// @Description
// @Description HOW TO USE IN SWAGGER:
// @Description 1. Click Authorize.
// @Description 2. Enter Sigtrack-Admin-API-Key under AdminApiKeyAuth.
// @Description 3. Execute this endpoint.
// @Description 4. Copy data.code from the response.
// @Description 5. Click Authorize again.
// @Description 6. Paste data.code under PrivilegeCodeAuth.
// @Description
// @Description IMPORTANT:
// @Description The full UUID is returned only once during creation.
// @Description The codePrefix is only for display and audit logs. Do not use codePrefix as X-Privilege-Code.
// @Description
// @Description EXAMPLE PERMISSIONS:
// @Description assistance:read, assistance:update_status, sos:read, sos:update_status, reports:read, reports:approve, alerts:read, alerts:create, alerts:update, alerts:delete
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAdminPrivilegeCodeRequest true "Privilege code creation payload"
// @Success 201 {object} map[string]interface{} "Privilege code created successfully. Copy data.code and use it as X-Privilege-Code."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, or invalid permissions"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Failure 500 {object} map[string]interface{} "Failed to create privilege code"
// @Router /admin/privilege-codes [post]
func (h *AdminPrivilegeCodeHandler) CreatePrivilegeCode(c *gin.Context) {
	var req dto.CreateAdminPrivilegeCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	createdBy := getAdminActor(c)

	code, err := h.service.CreatePrivilegeCode(
		c.Request.Context(),
		req,
		createdBy,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Privilege code created successfully",
		code,
	)
}

// GetPrivilegeCodes godoc
// @Summary List generated admin privilege codes
// @Description Returns privilege-code records created for organisations, departments, units, or admin levels.
// @Description
// @Description SECURITY:
// @Description This endpoint requires only AdminApiKeyAuth.
// @Description
// @Description IMPORTANT:
// @Description The full UUID is not returned here for security reasons.
// @Description Only codePrefix is returned so admins can identify the record without exposing the full secret UUID.
// @Description
// @Description Use this endpoint to review active, revoked, expired, and previously created privilege codes.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Produce json
// @Param limit query int false "Maximum number of privilege-code records to return. Default is 50. Maximum is 200."
// @Success 200 {object} map[string]interface{} "Privilege codes fetched successfully"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege codes"
// @Router /admin/privilege-codes [get]
func (h *AdminPrivilegeCodeHandler) GetPrivilegeCodes(c *gin.Context) {
	limit := parseInt64Query(c, "limit", 50, 200)

	codes, err := h.service.GetPrivilegeCodes(c.Request.Context(), limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch privilege codes", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privilege codes fetched successfully", codes)
}

// GetPrivilegeCodeByID godoc
// @Summary Get one privilege-code record
// @Description Fetches one privilege-code record by database ID.
// @Description
// @Description SECURITY:
// @Description This endpoint requires only AdminApiKeyAuth.
// @Description
// @Description IMPORTANT:
// @Description This endpoint does not return the full UUID.
// @Description If the full UUID is lost, revoke the old privilege code and generate a new one.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Privilege-code database ID"
// @Success 200 {object} map[string]interface{} "Privilege code fetched successfully"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Failure 404 {object} map[string]interface{} "Privilege code not found"
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege code"
// @Router /admin/privilege-codes/{id} [get]
func (h *AdminPrivilegeCodeHandler) GetPrivilegeCodeByID(c *gin.Context) {
	id := c.Param("id")

	code, err := h.service.GetPrivilegeCodeByID(c.Request.Context(), id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privilege code fetched successfully", code)
}

// ValidatePrivilegeCode godoc
// @Summary Validate a privilege UUID
// @Description Checks whether a full UUID privilege code is valid, active, not expired, and usable.
// @Description
// @Description SECURITY:
// @Description This endpoint requires only AdminApiKeyAuth.
// @Description
// @Description WHEN TO USE:
// @Description Use this endpoint when testing a generated UUID before applying it to protected admin endpoints.
// @Description
// @Description IMPORTANT:
// @Description Send the full UUID in the request body.
// @Description Do not send codePrefix. codePrefix is only for display and logs.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.ValidateAdminPrivilegeCodeRequest true "Full privilege UUID validation payload"
// @Success 200 {object} map[string]interface{} "Privilege code is valid"
// @Failure 400 {object} map[string]interface{} "Invalid request body, invalid UUID, expired code, or revoked code"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Router /admin/privilege-codes/validate [post]
func (h *AdminPrivilegeCodeHandler) ValidatePrivilegeCode(c *gin.Context) {
	var req dto.ValidateAdminPrivilegeCodeRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	code, err := h.service.ValidatePrivilegeCode(
		c.Request.Context(),
		req.UUID,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privilege code is valid", code)
}

// RevokePrivilegeCode godoc
// @Summary Revoke a privilege code
// @Description Disables a privilege UUID so it can no longer be used on protected admin endpoints.
// @Description
// @Description SECURITY:
// @Description This endpoint requires only AdminApiKeyAuth.
// @Description
// @Description WHAT REVOKE MEANS:
// @Description Revoke does not delete the database record.
// @Description It changes the privilege-code status from active to revoked.
// @Description The old UUID stops working, but the record remains for audit history.
// @Description
// @Description WHEN TO REVOKE:
// @Description Revoke a code when it is leaked, no longer trusted, assigned to the wrong level, or needs to be replaced by a new UUID.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Privilege-code database ID"
// @Param request body dto.RevokeAdminPrivilegeCodeRequest true "Revocation reason"
// @Success 200 {object} map[string]interface{} "Privilege code revoked successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request or privilege code already revoked"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Failure 404 {object} map[string]interface{} "Privilege code not found"
// @Router /admin/privilege-codes/{id}/revoke [put]
func (h *AdminPrivilegeCodeHandler) RevokePrivilegeCode(c *gin.Context) {
	id := c.Param("id")

	var req dto.RevokeAdminPrivilegeCodeRequest
	_ = c.ShouldBindJSON(&req)

	revokedBy := getAdminActor(c)

	code, err := h.service.RevokePrivilegeCode(
		c.Request.Context(),
		id,
		req.Reason,
		revokedBy,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privilege code revoked successfully", code)
}

// GetPrivilegeLogs godoc
// @Summary List privilege audit logs
// @Description Fetches audit logs for privilege-code creation, validation, revocation, and permission checks.
// @Description
// @Description SECURITY:
// @Description This endpoint requires only AdminApiKeyAuth.
// @Description
// @Description WHAT THIS LOG SHOWS:
// @Description - When a privilege code was created
// @Description - When a privilege code was validated
// @Description - When a privilege code was revoked
// @Description - Which admin endpoint was accessed
// @Description - Which permission was required
// @Description - Whether access was allowed or denied
// @Description - The codePrefix, organisation, level, IP address, and user agent involved
// @Tags Admin Privilege Logs
// @Security AdminApiKeyAuth
// @Produce json
// @Param limit query int false "Maximum number of logs to return. Default is 100. Maximum is 500."
// @Param privilegeCodeId query string false "Filter logs by privilege-code database ID"
// @Success 200 {object} map[string]interface{} "Privilege logs fetched successfully"
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid"
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege logs"
// @Router /admin/privilege-logs [get]
func (h *AdminPrivilegeCodeHandler) GetPrivilegeLogs(c *gin.Context) {
	limit := parseInt64Query(c, "limit", 100, 500)
	privilegeCodeID := c.Query("privilegeCodeId")

	logs, err := h.service.GetPrivilegeLogs(
		c.Request.Context(),
		limit,
		privilegeCodeID,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch privilege logs", err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Privilege logs fetched successfully", logs)
}

func parseInt64Query(c *gin.Context, key string, defaultValue int64, maxValue int64) int64 {
	value := defaultValue

	query := c.Query(key)
	if query != "" {
		parsed, err := strconv.ParseInt(query, 10, 64)
		if err == nil && parsed > 0 {
			value = parsed
		}
	}

	if value > maxValue {
		value = maxValue
	}

	return value
}

func getAdminActor(c *gin.Context) string {
	if userID, exists := c.Get("userId"); exists {
		if str, ok := userID.(string); ok && str != "" {
			return str
		}
	}

	if actor := c.GetHeader("X-Admin-Actor"); actor != "" {
		return actor
	}

	return "admin_api_key"
}
