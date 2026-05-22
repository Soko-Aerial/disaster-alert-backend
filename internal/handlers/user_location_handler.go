package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserLocationHandler struct {
	locationService *services.UserLocationService
}

func NewUserLocationHandler(
	locationService *services.UserLocationService,
) *UserLocationHandler {
	return &UserLocationHandler{
		locationService: locationService,
	}
}

func (h *UserLocationHandler) GetLocation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	location, err := h.locationService.GetUserLocation(userID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch user location",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"User location fetched successfully",
		location,
	)
}

func (h *UserLocationHandler) UpdateLocation(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.UpdateUserLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	location, err := h.locationService.UpdateUserLocation(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"User location updated successfully",
		location,
	)
}