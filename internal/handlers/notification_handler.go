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

// SaveFCMToken godoc
// @Summary Save user FCM token
// @Description Saves the Firebase Cloud Messaging token for the authenticated user's device.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description The mobile app should call this after login, app start, or whenever Firebase refreshes the push token.
// @Description
// @Description WHY THIS IS IMPORTANT:
// @Description Without saving this token, the backend may not be able to send push notifications to the user's phone.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - token: Firebase Cloud Messaging token from the mobile app.
// @Description
// @Description OPTIONAL FIELDS:
// @Description - deviceType: android, ios, or web.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "token": "fcm_token_example_abc123",
// @Description   "deviceType": "android"
// @Description }
// @Tags User Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SaveFCMTokenRequest true "FCM token payload. token is required."
// @Success 200 {object} map[string]interface{} "FCM token saved successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, or token could not be saved."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while saving FCM token."
// @Router /notifications/token [post]
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

// SendTestToMe godoc
// @Summary Send test notification to current user
// @Description Queues a test push notification for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this to confirm that the user's saved FCM token, Firebase setup, and notification queue are working.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED FIELDS:
// @Description - title: Notification title.
// @Description - body: Notification body.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "title": "Disaster Alert Test",
// @Description   "body": "This is a test notification from Disaster Alert."
// @Description }
// @Tags User Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.SendTestNotificationRequest true "Test notification payload. title and body are required."
// @Success 202 {object} map[string]interface{} "Test notification queued successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation failed."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 503 {object} map[string]interface{} "Notification queue is full."
// @Router /notifications/test/me [post]
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
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this from the admin dashboard to verify Firebase Cloud Messaging, queue processing, and app push delivery.
// @Description
// @Description IMPORTANT:
// @Description This endpoint should be admin-only because it sends a notification to every registered user.
// @Description Do not expose this to normal mobile users in production.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `notifications:send`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description notifications:send
// @Description
// @Description REQUIRED FIELDS:
// @Description - title: Notification title.
// @Description - body: Notification message.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "title": "Disaster Alert Test",
// @Description   "body": "This is a test broadcast notification."
// @Description }
// @Tags Admin Notifications
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param request body dto.SendTestNotificationRequest true "Broadcast test notification payload. title and body are required."
// @Success 202 {object} map[string]interface{} "Broadcast notification queued successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request body or validation error."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 503 {object} map[string]interface{} "Notification queue is full."
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
