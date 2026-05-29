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

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Notification marked as read",
		nil,
	)
}

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

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"All notifications marked as read",
		nil,
	)
}