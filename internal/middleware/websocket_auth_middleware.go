package middleware

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func WebSocketAuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := strings.TrimSpace(c.Query("token"))

		if tokenString == "" {
			authHeader := c.GetHeader("Authorization")

			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"WebSocket token is required",
				nil,
			)
			c.Abort()
			return
		}

		claims, err := jwtService.ValidateToken(tokenString)
		if err != nil {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Invalid or expired WebSocket token",
				err.Error(),
			)
			c.Abort()
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("role", claims.Role)

		c.Next()
	}
}