package websocket

import (
	"context"
	"net/http"

	"disaster_alert_backend/internal/repositories"

	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Handler struct {
	Hub            *Hub
	UserRepository *repositories.UserRepository
}

func NewHandler(
	hub *Hub,
	userRepository *repositories.UserRepository,
) *Handler {
	return &Handler{
		Hub:            hub,
		UserRepository: userRepository,
	}
}

// Connect godoc
// @Summary Connect to real-time WebSocket
// @Description Establishes a real-time WebSocket connection for mobile users or admin dashboards.
// @Description
// @Description USER WEBSOCKET URL:
// @Description ws://localhost:8080/api/v1/ws
// @Description wss://disaster-alert-backend-tiql.onrender.com/api/v1/ws
// @Description
// @Description USER AUTH:
// @Description Mobile/user WebSocket connections require a valid JWT token.
// @Description Recommended connection style:
// @Description Authorization: Bearer YOUR_JWT_TOKEN
// @Description
// @Description ADMIN WEBSOCKET URL:
// @Description ws://localhost:8080/api/v1/admin/ws
// @Description wss://disaster-alert-backend-tiql.onrender.com/api/v1/admin/ws
// @Description
// @Description ADMIN AUTH:
// @Description Admin WebSocket connections require the admin API key.
// @Description Required header:
// @Description Sigtrack-Admin-API-Key: YOUR_ADMIN_API_KEY
// @Description
// @Description EVENT FORMAT:
// @Description All WebSocket events follow this structure:
// @Description {
// @Description   "type": "EVENT_NAME",
// @Description   "data": {}
// @Description }
// @Description
// @Description EVENTS SENT TO ADMINS:
// @Description - SOS_CREATED: A user triggered SOS.
// @Description - ASSISTANCE_CREATED: A user submitted an assistance request.
// @Description - REPORT_CREATED: A user submitted an incident report.
// @Description - CHAT_MESSAGE_CREATED: A user sent a chat message.
// @Description
// @Description EVENTS SENT TO USERS BY COUNTRY:
// @Description - ALERT_CREATED: A new active alert was created for the user's country.
// @Description - ALERT_APPROVED: A report was approved and converted into an alert for the user's country.
// @Description
// @Description EVENTS SENT TO SPECIFIC USERS:
// @Description - NOTIFICATION_CREATED: A new in-app notification was created.
// @Description - CHAT_MESSAGE_CREATED: A new chat message was sent to the user.
// @Description - SOS_STATUS_UPDATED: The user's SOS status changed.
// @Description - ASSISTANCE_STATUS_UPDATED: The user's assistance request status changed.
// @Description
// @Description EXAMPLE EVENT:
// @Description {
// @Description   "type": "ALERT_CREATED",
// @Description   "data": {
// @Description     "id": "66e19b71c8f2a2b4d1234567",
// @Description     "title": "Heavy rainfall warning",
// @Description     "category": "weather",
// @Description     "severity": "high",
// @Description     "country": "Ghana",
// @Description     "createdAt": "2026-09-10T10:00:00Z"
// @Description   }
// @Description }
// @Description
// @Description IMPORTANT:
// @Description Swagger UI does not fully test WebSocket connections like normal REST endpoints.
// @Description Use Postman, Hoppscotch, browser client code, or your Flutter app to test WebSocket connections.
// @Tags WebSocket
// @Security BearerAuth
// @Produce json
// @Success 101 {string} string "Switching Protocols"
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired WebSocket authentication."
// @Router /ws [get]
// @Router /admin/ws [get]
func (h *Handler) Connect(c *gin.Context) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unauthorized WebSocket connection",
		})
		return
	}

	roleValue, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User role missing",
		})
		return
	}

	userID, ok := userIDValue.(string)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid user ID",
		})
		return
	}

	role, ok := roleValue.(string)
	if !ok || role == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid user role",
		})
		return
	}

	country := ""

	if role != "admin" && role != "super_admin" {
		objectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Invalid user ID format",
			})
			return
		}

		user, err := h.UserRepository.FindUserByID(objectID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "User not found",
			})
			return
		}

		if user.Location != nil {
			country = user.Location.Country
		}
	}

	conn, err := coderws.Accept(c.Writer, c.Request, &coderws.AcceptOptions{
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}

	client := &Client{
		UserID:  userID,
		Role:    role,
		Country: country,
		Conn:    conn,
	}

	h.Hub.AddClient(client)
	defer h.Hub.RemoveClient(client)

	for {
		_, _, err := conn.Read(context.Background())
		if err != nil {
			break
		}
	}
}
