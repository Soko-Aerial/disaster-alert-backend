package handlers

import (
	"net/http"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserProfileDetailsHandler struct {
	profileDetailsService *services.UserProfileDetailsService
}

func NewUserProfileDetailsHandler(
	profileDetailsService *services.UserProfileDetailsService,
) *UserProfileDetailsHandler {
	return &UserProfileDetailsHandler{
		profileDetailsService: profileDetailsService,
	}
}

func (h *UserProfileDetailsHandler) GetProfileDetails(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	user, err := h.profileDetailsService.GetProfileDetails(userID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch profile details",
			err,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Profile details fetched successfully",
		gin.H{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"phone":       user.Phone,
			"location":    user.Location,
			"medicalInfo": user.MedicalInfo,
			"insurance":   user.Insurance,
			"workplace":   user.Workplace,
			"role":        user.Role,
			"createdAt":   user.CreatedAt,
			"updatedAt":   user.UpdatedAt,
			"gender":      user.Gender,
			"dateOfBirth": user.DateOfBirth,
		},
	)
}

func (h *UserProfileDetailsHandler) UpdateProfileDetails(c *gin.Context) {
	userID, ok := getUserIDFromContext(c)
	if !ok {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized user", nil)
		return
	}

	var req dto.UpdateUserProfileDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	user, err := h.profileDetailsService.UpdateProfileDetails(userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), err)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Profile details updated successfully",
		gin.H{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"phone":       user.Phone,
			"location":    user.Location,
			"medicalInfo": user.MedicalInfo,
			"insurance":   user.Insurance,
			"workplace":   user.Workplace,
			"role":        user.Role,
			"updatedAt":   user.UpdatedAt,
			"gender":      user.Gender,
			"dateOfBirth": user.DateOfBirth,
		},
	)
}