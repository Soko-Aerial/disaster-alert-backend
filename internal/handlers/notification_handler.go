package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"
	"disaster_alert_backend/internal/queue"
	"disaster_alert_backend/internal/jobs"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type NotificationHandler struct {
	notificationService        *services.NotificationService
	firebaseMessagingService  *services.FirebaseMessagingService
	notificationQueue        *queue.NotificationQueue
	validator                  *validator.Validate
}

func NewNotificationHandler(
	notificationService *services.NotificationService,
	firebaseMessagingService *services.FirebaseMessagingService,
	notificationQueue *queue.NotificationQueue,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService:       notificationService,
		firebaseMessagingService: firebaseMessagingService,
		notificationQueue:        notificationQueue,
		validator:                 validator.New(),
	}
}

func (h *NotificationHandler) SaveFCMToken(c *gin.Context) {
	userIDValue, exists := c.Get("userId")

	if !exists {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"User not found in request context",
			nil,
		)
		return
	}

	userID, ok := userIDValue.(string)

	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"Invalid user id in request context",
			nil,
		)
		return
	}

	var req dto.SaveFCMTokenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	err := h.notificationService.SaveFCMToken(userID, req)

	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"FCM token saved successfully",
		nil,
	)
}

func (h *NotificationHandler) SendTestToMe(c *gin.Context) {
	userIDValue, exists := c.Get("userId")

	if !exists {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"User not found in request context",
			nil,
		)
		return
	}

	userID, ok := userIDValue.(string)

	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"Invalid user id in request context",
			nil,
		)
		return
	}

	var req dto.SendTestNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	queued := h.notificationQueue.Dispatch(jobs.NotificationJob{
		TargetType: jobs.TargetUser,
		UserID:     userID,
		Title:      req.Title,
		Body:       req.Body,
		Data: map[string]string{
			"type": "test",
		},
	})

	if !queued {
		utils.ErrorResponse(
			c,
			http.StatusServiceUnavailable,
			"Notification queue is full. Try again later.",
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusAccepted,
		"Test notification queued successfully",
		nil,
	)
}

func (h *NotificationHandler) SendTestToAll(c *gin.Context) {
	var req dto.SendTestNotificationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid request body",
			err.Error(),
		)
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Validation failed",
			err.Error(),
		)
		return
	}

	queued := h.notificationQueue.Dispatch(jobs.NotificationJob{
		TargetType: jobs.TargetAll,
		Title:      req.Title,
		Body:       req.Body,
		Data: map[string]string{
			"type": "broadcast_test",
		},
	})

	if !queued {
		utils.ErrorResponse(
			c,
			http.StatusServiceUnavailable,
			"Notification queue is full. Try again later.",
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusAccepted,
		"Broadcast notification queued successfully",
		nil,
	)
}