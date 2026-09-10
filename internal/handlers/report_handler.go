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

// CreateReport godoc
// @Summary Submit a new incident report
// @Description Allows an authenticated mobile user to report an emergency, disaster, hazard, or incident.
// @Description
// @Description The report is saved as `pending` first.
// @Description It does not become a public alert until an admin reviews and approves it.
// @Description
// @Description WHEN TO USE THIS ENDPOINT:
// @Description Use this when a mobile user wants to report flooding, fire, road accident, health outbreak, weather danger, security incident, or another emergency.
// @Description
// @Description AUTH REQUIRED:
// @Description This endpoint requires a valid user JWT token.
// @Description Click Authorize and paste: Bearer YOUR_JWT_TOKEN
// @Description
// @Description REQUEST OPTIONS:
// @Description This same endpoint supports two request types:
// @Description
// @Description 1. JSON request for reports without file upload.
// @Description Content-Type: application/json
// @Description
// @Description Example JSON body:
// @Description {
// @Description   "category": "flood",
// @Description   "description": "Flooding has started around the roadside and vehicles cannot pass.",
// @Description   "timeOfOccurrence": "2026-09-10T08:30:00Z",
// @Description   "latitude": 5.6037,
// @Description   "longitude": -0.1870,
// @Description   "address": "Circle, Accra",
// @Description   "country": "Ghana",
// @Description   "region": "Greater Accra"
// @Description }
// @Description
// @Description 2. Multipart request for reports with image or video upload.
// @Description Content-Type: multipart/form-data
// @Description Use the form field named `media` for the uploaded file.
// @Description
// @Description REQUIRED FIELDS:
// @Description - category: Type of incident. Example: flood
// @Description - description: What happened. Example: Flooding has blocked the road.
// @Description - timeOfOccurrence: When it happened. Example: 2026-09-10T08:30:00Z
// @Description - latitude: Incident latitude. Example: 5.6037
// @Description - longitude: Incident longitude. Example: -0.1870
// @Description
// @Description OPTIONAL FIELDS:
// @Description - address: Human-readable location or landmark.
// @Description - country: Country name.
// @Description - region: Region/state name.
// @Description - media: Optional image or video file.
// @Tags User Reports
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param category formData string true "Incident category. Allowed examples: flood, fire, security, health, weather, accident, other" Enums(flood, fire, security, health, weather, accident, other) example(flood)
// @Param description formData string true "Describe what happened clearly." example(Flooding has started around the roadside and vehicles cannot pass.)
// @Param timeOfOccurrence formData string true "When the incident happened. Recommended format: ISO date/time." example(2026-09-10T08:30:00Z)
// @Param latitude formData number true "Incident latitude." example(5.6037)
// @Param longitude formData number true "Incident longitude." example(-0.1870)
// @Param address formData string false "Readable location, area, landmark, or street." example(Circle, Accra)
// @Param country formData string false "Country where the incident happened." example(Ghana)
// @Param region formData string false "Region or state where the incident happened." example(Greater Accra)
// @Param media formData file false "Optional image or video evidence."
// @Success 201 {object} map[string]interface{} "Report submitted successfully. The report status will be pending."
// @Failure 400 {object} map[string]interface{} "Invalid request body, validation failed, invalid location, or media upload failed."
// @Failure 401 {object} map[string]interface{} "Missing or invalid user token."
// @Failure 500 {object} map[string]interface{} "Server error while submitting report."
// @Router /reports [post]
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
		Country:          c.PostForm("country"),
		Region:           c.PostForm("region"),
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
// @Summary List incident reports
// @Description Returns incident reports.
// @Description
// @Description ADMIN USE:
// @Description When called from /admin/reports, this returns user-submitted reports for admin review.
// @Description Admin must provide both AdminApiKeyAuth and PrivilegeCodeAuth with `reports:read` permission.
// @Description
// @Description MOBILE USER NOTE:
// @Description If this handler is exposed on /reports, make sure normal users only see reports they are allowed to see.
// @Description Do not expose all users' reports to normal mobile users.
// @Tags Admin Reports
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "Reports fetched successfully."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 500 {object} map[string]interface{} "Failed to fetch reports."
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
// @Summary Get one incident report
// @Description Returns the details of one user-submitted incident report.
// @Description
// @Description ADMIN USE:
// @Description Use this when an admin wants to inspect a report before approving it.
// @Description The admin should review the description, location, media evidence, reporter information, and status.
// @Description
// @Description REQUIRED PERMISSION:
// @Description reports:read
// @Tags Admin Reports
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Report ID. This is the MongoDB ObjectID of the report." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Report fetched successfully."
// @Failure 400 {object} map[string]interface{} "Invalid report ID."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Report not found."
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
// @Summary Approve a report and publish it as a public alert
// @Description Approves a pending user-submitted report and converts it into a public alert.
// @Description
// @Description WHAT HAPPENS WHEN THIS ENDPOINT IS CALLED:
// @Description 1. Backend finds the report by ID.
// @Description 2. Backend checks that the report is not already approved or rejected.
// @Description 3. Backend creates a new public alert from the report.
// @Description 4. Backend copies report media into alert imageUrls or videoUrls.
// @Description 5. Backend updates the report status to approved.
// @Description 6. Backend sends notifications and WebSocket updates where configured.
// @Description
// @Description IMPORTANT:
// @Description Use this endpoint only after an admin has verified that the report is credible.
// @Description A normal mobile user should not be allowed to call this endpoint.
// @Description
// @Description REQUIRED HEADERS:
// @Description - Sigtrack-Admin-API-Key: Your admin API key.
// @Description - X-Privilege-Code: A valid privilege code with `reports:approve`.
// @Description
// @Description REQUIRED PERMISSION:
// @Description reports:approve
// @Tags Admin Reports
// @Security AdminApiKeyAuth
// @Security PrivilegeCodeAuth
// @Produce json
// @Param id path string true "Report ID to approve. This must be a pending report ID." example(66e19b71c8f2a2b4d1234567)
// @Success 200 {object} map[string]interface{} "Report approved and public alert created successfully. Response includes alertId and alert."
// @Failure 400 {object} map[string]interface{} "Report cannot be approved. It may already be approved, rejected, or invalid."
// @Failure 401 {object} map[string]interface{} "Missing or invalid Admin API Key."
// @Failure 403 {object} map[string]interface{} "Missing, revoked, expired, or unauthorized privilege code."
// @Failure 404 {object} map[string]interface{} "Report not found."
// @Failure 500 {object} map[string]interface{} "Failed to approve report because of server, repository, or database error."
// @Router /admin/reports/{id}/approve [put]
func (h *ReportHandler) ApproveReport(c *gin.Context) {
	reportID := c.Param("id")

	alert, err := h.reportService.ApproveReport(reportID)
	if err != nil {
		statusCode := http.StatusBadRequest

		errorText := strings.ToLower(err.Error())

		if strings.Contains(errorText, "not found") {
			statusCode = http.StatusNotFound
		}

		if strings.Contains(errorText, "repository") ||
			strings.Contains(errorText, "database") {
			statusCode = http.StatusInternalServerError
		}

		utils.ErrorResponse(
			c,
			statusCode,
			err.Error(),
			nil,
		)
		return
	}

	utils.SuccessResponse(
		c,
		http.StatusOK,
		"Report approved and public alert created successfully",
		map[string]interface{}{
			"alertId": alert.ID.Hex(),
			"alert":   alert,
		},
	)
}
