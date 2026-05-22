package middleware

import (
	"net/http"
	"strings"

	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"

)
func AuthMiddleware(jwtService *services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Authorization header is required",
				nil,
			)
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Invalid authorization header format",
				nil,
			)
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtService.ValidateToken(tokenString)

		if err != nil {
			utils.ErrorResponse(
				c,
				http.StatusUnauthorized,
				"Invalid or expired token",
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