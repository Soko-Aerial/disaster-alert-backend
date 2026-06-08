package middleware

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func AdminAPIKeyMiddleware(adminAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		expectedKey := strings.TrimSpace(adminAPIKey)

		if expectedKey == "" {
			utils.ErrorResponse(
				c,
				http.StatusInternalServerError,
				"Admin API key is not configured",
				nil,
			)
			c.Abort()
			return
		}

		providedKey := strings.TrimSpace(c.GetHeader("Sigtrack-Admin-API-Key"))

		if providedKey == "" {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Admin API key is required",
				nil,
			)
			c.Abort()
			return
		}

		if providedKey != expectedKey {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Invalid admin API key",
				nil,
			)
			c.Abort()
			return
		}

		c.Set("userId", "000000000000000000000001")
		c.Set("role", "admin")

		c.Next()
	}
}