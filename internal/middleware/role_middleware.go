package middleware

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				"Access denied. User role not found",
				nil,
			)
			c.Abort()
			return
		}

		role, ok := roleValue.(string)
		if !ok || strings.TrimSpace(role) == "" {
			utils.ErrorResponse(
				c,
				http.StatusForbidden,
				"Access denied. Invalid user role",
				nil,
			)
			c.Abort()
			return
		}

		role = strings.ToLower(strings.TrimSpace(role))

		for _, allowedRole := range allowedRoles {
			if role == strings.ToLower(strings.TrimSpace(allowedRole)) {
				c.Next()
				return
			}
		}

		utils.ErrorResponse(
			c,
			http.StatusForbidden,
			"Access denied. Admin permission required",
			nil,
		)
		c.Abort()
	}
}