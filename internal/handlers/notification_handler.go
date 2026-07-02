package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/queue"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type NotificationHandler struct {
	notificationService      *services.NotificationService
	firebaseMessagingService *services.FirebaseMessagingService
	notificationQueue        *queue.NotificationQueue
	validator                *validator.Validate
}

func NewNotificationHandler(
	notificationService *services.NotificationService,
	firebaseMessagingService *services.FirebaseMessagingService,
	notificationQueue *queue.NotificationQueue,
) *NotificationHandler {
	return &NotificationHandler{
		notificationService:      notificationService,
		firebaseMessagingService: firebaseMessagingService,
		notificationQueue:        notificationQueue,
		validator:                validator.New(),
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

// SendTestToAll godoc
// @Summary Send notification to all users
// @Description Queues a broadcast test notification to all registered users.
// @Description
// @Description Use this endpoint to verify Firebase Cloud Messaging, notification queue processing, and mobile push delivery.
// @Description
// @Description SECURITY:
// @Description This admin endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description notifications:send
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {"title":"Disaster Alert Test","body":"This is a test broadcast notification."}
// @Tags Admin Notifications
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param request body dto.SendTestNotificationRequest true "Broadcast test notification payload"
// @Success 202 {object} map[string]interface{} "Broadcast notification queued successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation error"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 503 {object} map[string]interface{} "Notification queue is full"
// @Router /admin/notifications/test/all [post]
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
