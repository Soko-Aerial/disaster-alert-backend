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
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when you want to give a department, organisation, responder group, or admin user controlled access to admin endpoints.
// @Description
// @Description HOW PRIVILEGE CODES WORK:
// @Description 1. This endpoint creates a full UUID privilege code.
// @Description 2. The full UUID is returned only once.
// @Description 3. The backend stores a secure hash, not the raw UUID.
// @Description 4. The admin copies the UUID and sends it as X-Privilege-Code when calling protected admin endpoints.
// @Description
// @Description REQUIRED HEADER:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description
// @Description REQUIRED BODY FIELD:
// @Description - permissions: At least one permission is required.
// @Description
// @Description OPTIONAL BODY FIELDS:
// @Description - label
// @Description - purpose
// @Description - organisationId
// @Description - organisationName
// @Description - levelId
// @Description - levelName
// @Description - expiresAt
// @Description
// @Description IMPORTANT:
// @Description The full UUID is returned only once during creation.
// @Description codePrefix is only for display and audit logs.
// @Description Do not use codePrefix as X-Privilege-Code.
// @Description If the full UUID is lost, revoke the old code and generate a new one.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "permissions": [
// @Description     "dashboard:read",
// @Description     "reports:read",
// @Description     "sos:read",
// @Description     "sos:update_status",
// @Description     "chats:send"
// @Description   ],
// @Description   "expiresAt": "2026-09-30T23:59:00Z"
// @Description }
// @Description
// @Description FULL OPTIONAL REQUEST BODY EXAMPLE:
// @Description {
// @Description   "label": "Police Traffic Unit Access",
// @Description   "purpose": "Allow Police Traffic Unit to view reports, respond to SOS, and send chat replies",
// @Description   "organisationId": "firebase_police_org_id",
// @Description   "organisationName": "Ghana Police Service",
// @Description   "levelId": "firebase_traffic_unit_id",
// @Description   "levelName": "Traffic Unit",
// @Description   "permissions": [
// @Description     "reports:read",
// @Description     "sos:read",
// @Description     "sos:update_status",
// @Description     "chats:send"
// @Description   ],
// @Description   "expiresAt": "2026-09-30T23:59:00Z"
// @Description }
// @Description
// @Description COMMON PERMISSIONS CURRENTLY SUPPORTED IN CODE:
// @Description - alerts:read
// @Description - alerts:create
// @Description - alerts:update
// @Description - alerts:delete
// @Description - reports:read
// @Description - reports:approve
// @Description - sos:read
// @Description - sos:respond
// @Description - sos:update_status
// @Description - assistance:read
// @Description - assistance:update_status
// @Description - chats:read
// @Description - chats:send
// @Description - notifications:read
// @Description - notifications:send
// @Description - privilege_codes:read
// @Description - privilege_codes:create
// @Description - privilege_codes:revoke
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAdminPrivilegeCodeRequest true "Privilege code creation payload. permissions is required. label, purpose, organisationId, organisationName, levelId, levelName, and expiresAt are optional."
// @Success 201 {object} map[string]interface{} "Privilege code created successfully. Copy the returned full UUID and use it as X-Privilege-Code."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation error, invalid permission, or invalid expiry date."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 500 {object} map[string]interface{} "Server error while creating privilege code."
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
// @Summary List admin privilege codes
// @Description Returns privilege-code records created for organisations, departments, units, or admin levels.
// @Description
// @Description IMPORTANT:
// @Description This endpoint does not return the full UUID.
// @Description It returns codePrefix only so admins can identify records without exposing the secret privilege code.
// @Description
// @Description REQUIRED HEADER:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Produce json
// @Param limit query int false "Maximum number of privilege-code records to return. Default is 50. Maximum is 200." example(50)
// @Success 200 {object} map[string]interface{} "Privilege codes fetched successfully."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege codes."
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
// @Description IMPORTANT:
// @Description This endpoint does not return the full UUID.
// @Description If the full UUID is lost, revoke the old privilege code and generate a new one.
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Produce json
// @Param id path string true "Privilege-code database ID. This is the MongoDB ObjectID of the privilege-code record." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Privilege code fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid privilege-code ID."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 404 {object} map[string]interface{} "Privilege code not found."
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege code."
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
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this after generating a privilege code to confirm that the full UUID works before using protected admin endpoints.
// @Description
// @Description IMPORTANT:
// @Description Send the full UUID in the request body.
// @Description Do not send codePrefix.
// @Description codePrefix is only for display and audit logs.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "uuid": "4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"
// @Description }
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param request body dto.ValidateAdminPrivilegeCodeRequest true "Full privilege UUID validation payload. uuid is required."
// @Success 200 {object} map[string]interface{} "Privilege code is valid."
// @Failure 400 {object} map[string]interface{} "Invalid request body, invalid UUID, expired code, revoked code, or inactive code."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 500 {object} map[string]interface{} "Server error while validating privilege code."
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
// @Description WHAT REVOKE MEANS:
// @Description Revoke does not delete the database record.
// @Description It changes the privilege-code status from active to revoked.
// @Description The old UUID stops working, but the record remains for audit history.
// @Description
// @Description WHEN TO REVOKE:
// @Description Revoke a code when it is leaked, no longer trusted, assigned to the wrong level, or needs to be replaced by a new UUID.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "reason": "Access no longer needed"
// @Description }
// @Tags Admin Privilege Codes
// @Security AdminApiKeyAuth
// @Accept json
// @Produce json
// @Param id path string true "Privilege-code database ID. This is the MongoDB ObjectID of the privilege-code record." example(66e19b71c8f2a2b4d1234567)
// @Param request body dto.RevokeAdminPrivilegeCodeRequest false "Optional revocation reason."
// @Success 200 {object} map[string]interface{} "Privilege code revoked successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request or privilege code already revoked."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 404 {object} map[string]interface{} "Privilege code not found."
// @Failure 500 {object} map[string]interface{} "Server error while revoking privilege code."
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
// @Description WHAT THIS LOG SHOWS:
// @Description - When a privilege code was created.
// @Description - When a privilege code was validated.
// @Description - When a privilege code was revoked.
// @Description - Which admin endpoint was accessed.
// @Description - Which permission was required.
// @Description - Whether access was allowed or denied.
// @Description - The codePrefix, organisation, level, IP address, and user agent involved.
// @Description
// @Description REQUIRED HEADER:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Tags Admin Privilege Logs
// @Security AdminApiKeyAuth
// @Produce json
// @Param limit query int false "Maximum number of logs to return. Default is 100. Maximum is 500." example(100)
// @Param privilegeCodeId query string false "Filter logs by privilege-code database ID." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Privilege logs fetched successfully."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 500 {object} map[string]interface{} "Failed to fetch privilege logs."
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
