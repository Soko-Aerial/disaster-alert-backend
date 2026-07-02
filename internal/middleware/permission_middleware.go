package middleware

import (
	"net/http"

	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

const PrivilegeCodeHeader = "X-Privilege-Code"

func RequirePrivilegePermission(
	privilegeService *services.AdminPrivilegeCodeService,
	requiredPermission string,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawCode := c.GetHeader(PrivilegeCodeHeader)

		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		if rawCode == "" {
			capturePrivilegeMiddlewareEvent(
				c,
				sentry.LevelWarning,
				"Admin privilege code missing",
				requiredPermission,
				endpoint,
				nil,
				nil,
			)

			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				"Privilege code is required for this admin action",
				nil,
			)
			c.Abort()
			return
		}

		result, err := privilegeService.CheckPermission(
			c.Request.Context(),
			rawCode,
			requiredPermission,
			endpoint,
			c.Request.Method,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
		if err != nil {
			capturePrivilegeMiddlewareEvent(
				c,
				sentry.LevelError,
				"Failed to verify admin privilege permission",
				requiredPermission,
				endpoint,
				nil,
				err,
			)

			utils.ErrorResponse(
				c,
				http.StatusInternalServerError,
				"Failed to verify privilege permission",
				err.Error(),
			)
			c.Abort()
			return
		}

		if result == nil || !result.Allowed {
			message := "You do not have permission to perform this action"
			if result != nil && result.Message != "" {
				message = result.Message
			}

			capturePrivilegeMiddlewareEvent(
				c,
				sentry.LevelWarning,
				"Admin privilege permission denied",
				requiredPermission,
				endpoint,
				result,
				nil,
			)

			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				message,
				nil,
			)
			c.Abort()
			return
		}

		if result.Code != nil {
			c.Set("privilegeCodeId", result.Code.ID.Hex())
			c.Set("organisationId", result.Code.OrganisationID)
			c.Set("organisationName", result.Code.OrganisationName)
			c.Set("levelId", result.Code.LevelID)
			c.Set("levelName", result.Code.LevelName)
			c.Set("privilegePermissions", result.Code.Permissions)
		}

		c.Next()
	}
}

func capturePrivilegeMiddlewareEvent(
	c *gin.Context,
	level sentry.Level,
	message string,
	requiredPermission string,
	endpoint string,
	result *services.PermissionCheckResult,
	err error,
) {
	hub := sentrygin.GetHubFromContext(c)
	if hub == nil {
		return
	}

	hub.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)

		scope.SetTag("security_area", "admin_privilege")
		scope.SetTag("required_permission", requiredPermission)
		scope.SetTag("endpoint", endpoint)
		scope.SetTag("method", c.Request.Method)

		contextData := map[string]interface{}{
			"client_ip":           c.ClientIP(),
			"user_agent":          c.GetHeader("User-Agent"),
			"requiredPermission":  requiredPermission,
			"endpoint":            endpoint,
			"method":              c.Request.Method,
			"hasPrivilegeHeader":  c.GetHeader(PrivilegeCodeHeader) != "",
		}

		if result != nil {
			contextData["allowed"] = result.Allowed
			contextData["message"] = result.Message

			if result.Code != nil {
				scope.SetTag("code_prefix", result.Code.CodePrefix)
				scope.SetTag("organisation_id", result.Code.OrganisationID)
				scope.SetTag("level_id", result.Code.LevelID)

				contextData["codePrefix"] = result.Code.CodePrefix
				contextData["organisationId"] = result.Code.OrganisationID
				contextData["organisationName"] = result.Code.OrganisationName
				contextData["levelId"] = result.Code.LevelID
				contextData["levelName"] = result.Code.LevelName
				contextData["permissions"] = result.Code.Permissions
			}
		}

		scope.SetContext("admin_privilege_check", contextData)

		if err != nil {
			hub.CaptureException(err)
			return
		}

		hub.CaptureMessage(message)
	})
}