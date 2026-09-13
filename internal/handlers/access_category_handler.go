package handlers

import (
	"net/http"
	"strconv"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AccessCategoryHandler struct {
	service   *services.AccessCategoryService
	validator *validator.Validate
}

func NewAccessCategoryHandler(
	service *services.AccessCategoryService,
) *AccessCategoryHandler {
	return &AccessCategoryHandler{
		service:   service,
		validator: validator.New(),
	}
}

// CreateCategory godoc
// @Summary Create an access category
// @Description Creates a public, access, or both-type category for user-facing reporting and/or organisation privilege grants.
// @Description
// @Description Existing system categories are seeded from the current alert preference categories such as fire, flood, weather, health, conflict, protests, robbery, munitions, galamsey, unverified activity, and critical alerts.
// @Description Admin-created categories can be added later without changing code.
// @Tags Admin Access Categories
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param request body dto.CreateAccessCategoryRequest true "Access category creation payload."
// @Success 201 {object} map[string]interface{} "Access category created successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request or invalid allowedActions."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 403 {object} map[string]interface{} "Privilege code missing or permission denied."
// @Router /admin/access-categories [post]
func (h *AccessCategoryHandler) CreateCategory(c *gin.Context) {
	var req dto.CreateAccessCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	category, err := h.service.CreateCategory(c.Request.Context(), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusCreated,
		"Access category created successfully",
		category,
	)
}

// GetCategories godoc
// @Summary List access categories
// @Description Returns system and admin-created access categories.
// @Tags Admin Access Categories
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param kind query string false "Filter by category kind: public, access, or both."
// @Param ownerOrganisationId query string false "Filter by owner organisation ID."
// @Param includeInactive query bool false "Include inactive/deactivated categories."
// @Param limit query int false "Maximum records to return. Default 100. Maximum 500."
// @Success 200 {object} map[string]interface{} "Access categories fetched successfully."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 403 {object} map[string]interface{} "Privilege code missing or permission denied."
// @Failure 500 {object} map[string]interface{} "Failed to fetch access categories."
// @Router /admin/access-categories [get]
func (h *AccessCategoryHandler) GetCategories(c *gin.Context) {
	limit := int64(100)

	if limitQuery := c.Query("limit"); limitQuery != "" {
		parsed, err := strconv.ParseInt(limitQuery, 10, 64)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}

	includeInactive := c.Query("includeInactive") == "true"

	categories, err := h.service.GetCategories(
		c.Request.Context(),
		c.Query("kind"),
		c.Query("ownerOrganisationId"),
		includeInactive,
		limit,
	)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch access categories", err.Error())
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Access categories fetched successfully",
		categories,
	)
}

// GetCategoryByID godoc
// @Summary Get one access category
// @Description Fetches one access category by MongoDB ObjectID.
// @Tags Admin Access Categories
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Access category ID."
// @Success 200 {object} map[string]interface{} "Access category fetched successfully."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 403 {object} map[string]interface{} "Privilege code missing or permission denied."
// @Failure 404 {object} map[string]interface{} "Access category not found."
// @Router /admin/access-categories/{id} [get]
func (h *AccessCategoryHandler) GetCategoryByID(c *gin.Context) {
	category, err := h.service.GetCategoryByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, err.Error(), nil)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Access category fetched successfully",
		category,
	)
}

// UpdateCategory godoc
// @Summary Update an access category
// @Description Updates an existing access category.
// @Tags Admin Access Categories
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Accept json
// @Produce json
// @Param id path string true "Access category ID."
// @Param request body dto.UpdateAccessCategoryRequest true "Access category update payload."
// @Success 200 {object} map[string]interface{} "Access category updated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request or invalid allowedActions."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 403 {object} map[string]interface{} "Privilege code missing or permission denied."
// @Router /admin/access-categories/{id} [put]
func (h *AccessCategoryHandler) UpdateCategory(c *gin.Context) {
	var req dto.UpdateAccessCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	category, err := h.service.UpdateCategory(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Access category updated successfully",
		category,
	)
}

// DeactivateCategory godoc
// @Summary Deactivate an access category
// @Description Soft-deactivates an access category by setting isActive to false.
// @Tags Admin Access Categories
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Access category ID."
// @Success 200 {object} map[string]interface{} "Access category deactivated successfully."
// @Failure 400 {object} map[string]interface{} "Invalid request."
// @Failure 401 {object} map[string]interface{} "Admin API key missing or invalid."
// @Failure 403 {object} map[string]interface{} "Privilege code missing or permission denied."
// @Router /admin/access-categories/{id} [delete]
func (h *AccessCategoryHandler) DeactivateCategory(c *gin.Context) {
	category, err := h.service.DeactivateCategory(c.Request.Context(), c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error(), nil)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Access category deactivated successfully",
		category,
	)
}
