package handlers

import (
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AppNotificationHandler struct {
	appNotificationService *services.AppNotificationService
}

func NewAppNotificationHandler(
	appNotificationService *services.AppNotificationService,
) *AppNotificationHandler {
	return &AppNotificationHandler{
		appNotificationService: appNotificationService,
	}
}

// GetMyNotifications godoc
// @Summary Get my notifications
// @Description Returns in-app notifications for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this to show the user's notification inbox inside the mobile app.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description QUERY PARAMETERS:
// @Description - limit: Optional number of notifications to return. Default is 50.
// @Tags User Notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Maximum number of notifications to return. Default is 50." example(50)
// @Success 200 {object} map[string]interface{} "Notifications fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch notifications."
// @Router /notifications [get]
func (h *AppNotificationHandler) GetMyNotifications(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	limit := 50

	if limitQuery := c.Query("limit"); limitQuery != "" {
		parsedLimit, err := strconv.Atoi(limitQuery)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	notifications, err := h.appNotificationService.GetUserNotifications(
		userID,
		limit,
	)

	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch notifications",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Notifications fetched successfully",
		notifications,
	)
}

// GetMyNotificationHistory godoc
// @Summary Get my notification history
// @Description Returns longer in-app notification history for the authenticated user.
// @Description
// @Description MAIN DIFFERENCE:
// @Description /notifications returns the active inbox.
// @Description /notifications/history returns older read notifications too.
// @Description
// @Tags User Notifications
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Maximum number of notifications to return. Default is 100." example(100)
// @Success 200 {object} map[string]interface{} "Notification history fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch notification history."
// @Router /notifications/history [get]
func (h *AppNotificationHandler) GetMyNotificationHistory(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	limit := 100

	if limitQuery := c.Query("limit"); limitQuery != "" {
		parsedLimit, err := strconv.Atoi(limitQuery)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	notifications, err := h.appNotificationService.GetUserNotificationHistory(
		userID,
		limit,
	)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch notification history",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Notification history fetched successfully",
		notifications,
	)
}

// GetUnreadCount godoc
// @Summary Get unread notification count
// @Description Returns the number of unread in-app notifications for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this to show a badge count on the notification icon in the mobile app.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Tags User Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Unread notification count fetched successfully. Response data contains count."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to fetch unread notification count."
// @Router /notifications/unread-count [get]
func (h *AppNotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	count, err := h.appNotificationService.GetUnreadCount(userID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch unread notification count",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Unread notification count fetched successfully",
		gin.H{
			"count": count,
		},
	)
}

// MarkRead godoc
// @Summary Mark notifications as read
// @Description Marks selected notifications as read for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the user opens one or more notifications and they should no longer appear as unread.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Description
// @Description REQUIRED BODY:
// @Description Send notificationIds as a list of notification IDs.
// @Description
// @Description EXAMPLE REQUEST BODY:
// @Description {
// @Description   "notificationIds": [
// @Description     "66e19b71c8f2a2b4d1234567"
// @Description   ]
// @Description }
// @Tags User Notifications
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.MarkAppNotificationReadRequest true "Notification read payload. notificationIds is the list of notification IDs to mark as read."
// @Success 200 {object} map[string]interface{} "Notification marked as read."
// @Failure 400 {object} map[string]interface{} "Invalid request body or notification could not be marked as read."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Server error while marking notification as read."
// @Router /notifications/read [put]
func (h *AppNotificationHandler) MarkRead(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.MarkAppNotificationReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.appNotificationService.MarkRead(userID, req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}
	_ = h.appNotificationService.ArchiveOldReadNotificationsForUser(userID, 48)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Notification marked as read",
		nil,
	)
}

// MarkAllRead godoc
// @Summary Mark all notifications as read
// @Description Marks all in-app notifications as read for the authenticated user.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when the user taps a “Mark all as read” button in the mobile app.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN.
// @Tags User Notifications
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "All notifications marked as read."
// @Failure 401 {object} map[string]interface{} "Missing, invalid, or expired user token."
// @Failure 500 {object} map[string]interface{} "Failed to mark notifications as read."
// @Router /notifications/read-all [put]
func (h *AppNotificationHandler) MarkAllRead(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	if err := h.appNotificationService.MarkAllRead(userID); err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to mark notifications as read",
			err,
		)
		return
	}
	_ = h.appNotificationService.ArchiveOldReadNotificationsForUser(userID, 48)

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"All notifications marked as read",
		nil,
	)
}
