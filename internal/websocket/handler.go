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


// Connect Admin WebSocket
//
// @Summary Connect Admin WebSocket
// @Description Establishes a real-time WebSocket connection for the admin dashboard.
// @Description
// @Description Admin WebSocket URL:
// @Description wss://disaster-alert-backend-tiql.onrender.com/api/v1/admin/ws
// @Description
// @Description Required header:
// @Description Sigtrack-Admin-API-Key: YOUR_ADMIN_API_KEY
// @Description
// @Description Event format:
// @Description {"type":"EVENT_NAME","data":{}}
// @Description
// @Description Admin dashboard events:
// @Description SOS_CREATED - sent when a user creates an SOS request.
// @Description ASSISTANCE_CREATED - sent when a user submits an assistance request.
// @Description REPORT_CREATED - sent when a user submits a report.
// @Description ALERT_CREATED - sent when an active alert is created.
// @Description ALERT_APPROVED - sent when an alert/report is approved.
// @Description CHAT_MESSAGE_CREATED - sent when a user sends a chat message.
// @Description
// @Description SOS_CREATED example:
// @Description {"type":"SOS_CREATED","data":{"id":"sos_id","user":{"id":"user_id","name":"User","email":"user@email.com","phone":"0240000000","location":{"country":"Ghana","region":"Greater Accra","address":"Accra"}},"emergencyType":"medical","message":"Need help","latitude":5.6037,"longitude":-0.1870,"address":"Accra","status":"active","createdAt":"2026-06-11T10:00:00Z"}}
// @Description
// @Description ASSISTANCE_CREATED example:
// @Description {"type":"ASSISTANCE_CREATED","data":{"id":"assistance_id","user":{"id":"user_id","name":"User","email":"user@email.com"},"assistanceType":"medical","urgencyLevel":"high","affectedIndividuals":3,"address":"Accra","status":"pending","createdAt":"2026-06-11T10:00:00Z"}}
// @Description
// @Tags Admin WebSocket
// @Security AdminApiKeyAuth
// @Produce json
// @Success 101 {string} string "Switching Protocols"
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