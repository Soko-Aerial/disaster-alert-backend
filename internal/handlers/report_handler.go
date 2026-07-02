package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/services"
	"disaster_alert_backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ReportHandler struct {
	reportService     *services.ReportService
	cloudinaryService *services.CloudinaryService
	validator         *validator.Validate
}

func NewReportHandler(
	reportService *services.ReportService,
	cloudinaryService *services.CloudinaryService,
) *ReportHandler {
	return &ReportHandler{
		reportService:     reportService,
		cloudinaryService: cloudinaryService,
		validator:         validator.New(),
	}
}

func (h *ReportHandler) CreateReport(c *gin.Context) {
	contentType := c.GetHeader("Content-Type")

	if strings.Contains(contentType, "multipart/form-data") {
		h.createReportMultipart(c)
		return
	}

	h.createReportJSON(c)
}

func (h *ReportHandler) createReportJSON(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req dto.CreateReportRequest

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

	report, err := h.reportService.CreateReport(userID, req)
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
		http.StatusCreated,
		"Report submitted successfully",
		report,
	)
}

func (h *ReportHandler) createReportMultipart(c *gin.Context) {
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	latitude, err := strconv.ParseFloat(c.PostForm("latitude"), 64)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid latitude",
			err.Error(),
		)
		return
	}

	longitude, err := strconv.ParseFloat(c.PostForm("longitude"), 64)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusBadRequest,
			"Invalid longitude",
			err.Error(),
		)
		return
	}

	req := dto.CreateReportRequest{
		Category:         c.PostForm("category"),
		Description:      c.PostForm("description"),
		TimeOfOccurrence: c.PostForm("timeOfOccurrence"),
		Latitude:         latitude,
		Longitude:        longitude,
		Address:          c.PostForm("address"),
		MediaURLs:        []string{},
		Media:            []models.ReportMedia{},
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

	file, header, err := c.Request.FormFile("media")

	if err == nil {
		defer file.Close()

		if h.cloudinaryService == nil {
			utils.ErrorResponse(
				c,
				http.StatusInternalServerError,
				"Cloudinary service is not configured",
				nil,
			)
			return
		}

		uploadResult, uploadErr := h.cloudinaryService.UploadReportMedia(
			c.Request.Context(),
			file,
			header,
		)
		if uploadErr != nil {
			utils.ErrorResponse(
				c,
				http.StatusBadRequest,
				"Failed to upload report media",
				uploadErr.Error(),
			)
			return
		}

		req.MediaURLs = append(req.MediaURLs, uploadResult.MediaURL)

		req.Media = append(req.Media, models.ReportMedia{
			URL:      uploadResult.MediaURL,
			Type:     uploadResult.MediaType,
			PublicID: uploadResult.MediaPublicID,
		})
	} else {
		if !errors.Is(err, http.ErrMissingFile) {
			utils.ErrorResponse(
				c,
				http.StatusBadRequest,
				"Invalid media file",
				err.Error(),
			)
			return
		}
	}

	report, err := h.reportService.CreateReport(userID, req)
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
		http.StatusCreated,
		"Report submitted successfully",
		report,
	)
}

func (h *ReportHandler) getUserID(c *gin.Context) (string, bool) {
	userIDValue, exists := c.Get("userId")
	if !exists {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"User not found in request context",
			nil,
		)
		return "", false
	}

	userID, ok := userIDValue.(string)
	if !ok {
		utils.ErrorResponse(
			c,
			http.StatusUnauthorized,
			"Invalid user id in request context",
			nil,
		)
		return "", false
	}

	return userID, true
}

// GetReports godoc
// @Summary List incident reports for admin
// @Description Returns all user-submitted incident reports for review in the admin dashboard.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description reports:read
// @Tags Admin Reports
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Reports fetched successfully"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 500 {object} map[string]interface{} "Failed to fetch reports"
// @Router /admin/reports [get]

func (h *ReportHandler) GetReports(c *gin.Context) {
	reports, err := h.reportService.GetReports()
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusInternalServerError,
			"Failed to fetch reports",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Reports fetched successfully",
		reports,
	)
}

// GetReportByID godoc
// @Summary Get one incident report for admin
// @Description Returns details of a single user-submitted incident report.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description reports:read
// @Tags Admin Reports
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{} "Report fetched successfully"
// @Failure 400 {object} map[string]interface{} "Invalid report ID"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Router /admin/reports/{id} [get]

func (h *ReportHandler) GetReportByID(c *gin.Context) {
	reportID := c.Param("id")

	report, err := h.reportService.GetReportByID(reportID)
	if err != nil {
		utils.ErrorResponse(
			c,
			http.StatusNotFound,
			"Report not found",
			err.Error(),
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Report fetched successfully",
		report,
	)
}

// ApproveReport godoc
// @Summary Approve user report and publish alert
// @Description Approves a user-submitted incident report and converts it into an alert that can be shown to users through the alert feed, notifications, and WebSocket updates.
// @Description
// @Description SECURITY:
// @Description This endpoint requires BOTH AdminApiKeyAuth and PrivilegeCodeAuth.
// @Description
// @Description REQUIRED PERMISSION:
// @Description reports:approve
// @Description
// @Description Use this endpoint only after an admin verifies that the submitted report is credible and should become an official alert.
// @Tags Admin Reports
// @Security AdminApiKeyAuth && PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Report ID"
// @Success 200 {object} map[string]interface{} "Report approved and alert published successfully"
// @Failure 400 {object} map[string]interface{} "Report cannot be approved"
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key"
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code"
// @Failure 404 {object} map[string]interface{} "Report not found"
// @Failure 500 {object} map[string]interface{} "Failed to approve report"
// @Router /admin/reports/{id}/approve [put]
func (h *ReportHandler) ApproveReport(c *gin.Context) {
	reportID := c.Param("id")

	alert, err := h.reportService.ApproveReport(reportID)
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
		"Report approved and alert published successfully",
		alert,
	)
}
