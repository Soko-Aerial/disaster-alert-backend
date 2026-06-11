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

	country := ""
	if user.Location != nil {
		country = user.Location.Country
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