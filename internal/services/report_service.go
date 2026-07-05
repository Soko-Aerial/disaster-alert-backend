package services

import (
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportService struct {
	reportRepo               *repositories.ReportRepository
	alertRepo                *repositories.AlertRepository
	userRepo                 *repositories.UserRepository
	eventNotificationService *EventNotificationService
	broadcaster              *websocket.Broadcaster
}

func NewReportService(
	reportRepo *repositories.ReportRepository,
	alertRepo *repositories.AlertRepository,
	userRepo *repositories.UserRepository,
	eventNotificationService *EventNotificationService,
	broadcaster *websocket.Broadcaster,
) *ReportService {
	return &ReportService{
		reportRepo:               reportRepo,
		alertRepo:                alertRepo,
		userRepo:                 userRepo,
		eventNotificationService: eventNotificationService,
		broadcaster:              broadcaster,
	}
}

func (s *ReportService) CreateReport(
	userID string,
	req dto.CreateReportRequest,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now().UTC()

	location := models.ReportLocation{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Address:   strings.TrimSpace(req.Address),
		Country:   strings.TrimSpace(req.Country),
		Region:    strings.TrimSpace(req.Region),
	}

	location = s.enrichReportLocationFromUser(objectID, location)

	report := models.Report{
		UserID:           objectID,
		Category:         strings.TrimSpace(req.Category),
		Description:      strings.TrimSpace(req.Description),
		TimeOfOccurrence: strings.TrimSpace(req.TimeOfOccurrence),
		Location:         location,
		MediaURLs:        cleanStringList(req.MediaURLs),
		Media:            req.Media,
		Status:           "pending",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	createdReport, err := s.reportRepo.Create(report)
	if err != nil {
		return nil, err
	}

	locationText := strings.TrimSpace(createdReport.Location.Address)
	if locationText == "" {
		locationText = "Lat: " +
			floatToString(createdReport.Location.Latitude) +
			", Lng: " +
			floatToString(createdReport.Location.Longitude)
	}

	reportType := createdReport.Category
	if strings.TrimSpace(reportType) == "" {
		reportType = "incident"
	}

	response := s.buildReportResponse(createdReport)

	userData := s.buildUserSummary(createdReport.UserID)

	if s.eventNotificationService != nil {
		go s.eventNotificationService.NotifyAdminsForReport(
			createdReport.ID,
			getUserDisplayName(userData),
			reportType,
			locationText,
		)
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastReportCreated(response)
	}

	return response, nil
}

func (s *ReportService) GetReports() ([]map[string]interface{}, error) {
	reports, err := s.reportRepo.FindAll()
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(reports))

	for i := range reports {
		response = append(response, s.buildReportResponse(&reports[i]))
	}

	return response, nil
}

func (s *ReportService) GetReportByID(
	reportID string,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report id")
	}

	report, err := s.reportRepo.FindByID(objectID)
	if err != nil || report == nil {
		return nil, errors.New("report not found")
	}

	return s.buildReportResponse(report), nil
}

func (s *ReportService) ApproveReport(
	reportID string,
) (*models.Alert, error) {
	if s.alertRepo == nil {
		return nil, errors.New("alert repository is not configured")
	}

	objectID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report id")
	}

	report, err := s.reportRepo.FindByID(objectID)
	if err != nil || report == nil {
		return nil, errors.New("report not found")
	}

	reportStatus := strings.ToLower(strings.TrimSpace(report.Status))

	if reportStatus == "approved" {
		return nil, errors.New("report already approved")
	}

	if reportStatus == "rejected" {
		return nil, errors.New("rejected report cannot be approved")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)

	location := s.enrichAlertLocationFromReport(report)

	imageURLs, videoURLs := splitReportMedia(report)

	category := strings.TrimSpace(report.Category)
	if category == "" {
		category = "incident"
	}

	severity := inferReportSeverity(category, report.Description)
	priorityScore := inferReportPriorityScore(severity, category)
	priorityLabel := inferReportPriorityLabel(priorityScore)

	locationName := bestLocationName(location)

	title := buildApprovedReportAlertTitle(category, locationName)
	summary := buildApprovedReportAlertSummary(category, locationName)

	description := strings.TrimSpace(report.Description)
	if description == "" {
		description = summary
	}

	eventTime := parseReportEventTime(report.TimeOfOccurrence, now)

	alert := models.Alert{
		Title:       title,
		Summary:     summary,
		Description: description,
		Category:    category,
		Severity:    severity,
		Status:      "active",

		Location: location,

		RadiusKm: 10,

		SafetyInstructions: buildReportSafetyInstructions(category),

		SourceType:     "community_report",
		SourceName:     "Verified Community Report",
		LinkedReportID: &report.ID,

		ImageURLs: imageURLs,
		VideoURLs: videoURLs,

		Tags: buildReportAlertTags(
			category,
			location.Country,
			location.Region,
		),

		PriorityScore: priorityScore,
		PriorityLabel: priorityLabel,
		IsBreaking:    priorityScore >= 80,
		IsVerified:    true,

		EventTime:  eventTime,
		ExpiresAt:  &expiresAt,
		Confidence: 0.85,

		CreatedAt: now,
		UpdatedAt: now,
	}

	createdAlert, err := s.alertRepo.Create(alert)
	if err != nil {
		return nil, err
	}

	_, err = s.reportRepo.UpdateStatus(objectID, "approved")
	if err != nil {
		return nil, err
	}

	if s.eventNotificationService != nil {
		go s.eventNotificationService.NotifyUsersForApprovedAlert(createdAlert)
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastAlertApproved(
			createdAlert.Location.Country,
			map[string]interface{}{
				"id":                 createdAlert.ID.Hex(),
				"title":              createdAlert.Title,
				"summary":            createdAlert.Summary,
				"description":        createdAlert.Description,
				"category":           createdAlert.Category,
				"severity":           createdAlert.Severity,
				"status":             createdAlert.Status,
				"location":           createdAlert.Location,
				"radiusKm":           createdAlert.RadiusKm,
				"safetyInstructions": createdAlert.SafetyInstructions,
				"sourceType":         createdAlert.SourceType,
				"sourceName":         createdAlert.SourceName,
				"imageUrls":          createdAlert.ImageURLs,
				"videoUrls":          createdAlert.VideoURLs,
				"priorityScore":      createdAlert.PriorityScore,
				"priorityLabel":      createdAlert.PriorityLabel,
				"isBreaking":         createdAlert.IsBreaking,
				"isVerified":         createdAlert.IsVerified,
				"eventTime":          createdAlert.EventTime,
				"expiresAt":          createdAlert.ExpiresAt,
				"createdAt":          createdAlert.CreatedAt,
				"updatedAt":          createdAlert.UpdatedAt,
			},
		)
	}

	return createdAlert, nil
}

func (s *ReportService) buildReportResponse(
	report *models.Report,
) map[string]interface{} {
	if report == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":               report.ID.Hex(),
		"userId":           report.UserID.Hex(),
		"user":             s.buildUserSummary(report.UserID),
		"category":         report.Category,
		"description":      report.Description,
		"timeOfOccurrence": report.TimeOfOccurrence,
		"location":         report.Location,
		"mediaUrls":        report.MediaURLs,
		"media":            report.Media,
		"status":           report.Status,
		"createdAt":        report.CreatedAt,
		"updatedAt":        report.UpdatedAt,
	}
}

func (s *ReportService) buildUserSummary(
	userID primitive.ObjectID,
) map[string]interface{} {
	userData := map[string]interface{}{
		"id":       userID.Hex(),
		"name":     "User",
		"email":    "",
		"phone":    "",
		"gender":   "",
		"role":     "",
		"location": nil,
	}

	if s.userRepo == nil {
		return userData
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil || user == nil {
		return userData
	}

	userData["name"] = user.Name
	userData["email"] = user.Email
	userData["phone"] = user.Phone
	userData["gender"] = user.Gender
	userData["role"] = user.Role

	if user.Location != nil {
		userData["location"] = map[string]interface{}{
			"name":      user.Location.Name,
			"country":   user.Location.Country,
			"region":    user.Location.Region,
			"address":   user.Location.Address,
			"latitude":  user.Location.Latitude,
			"longitude": user.Location.Longitude,
			"source":    user.Location.Source,
			"isDefault": user.Location.IsDefault,
		}
	}

	return userData
}

func (s *ReportService) enrichReportLocationFromUser(
	userID primitive.ObjectID,
	location models.ReportLocation,
) models.ReportLocation {
	if s.userRepo == nil {
		return location
	}

	user, err := s.userRepo.FindUserByID(userID)
	if err != nil || user == nil || user.Location == nil {
		return location
	}

	if strings.TrimSpace(location.Address) == "" {
		location.Address = strings.TrimSpace(user.Location.Address)
	}

	if strings.TrimSpace(location.Country) == "" {
		location.Country = strings.TrimSpace(user.Location.Country)
	}

	if strings.TrimSpace(location.Region) == "" {
		location.Region = strings.TrimSpace(user.Location.Region)
	}

	if location.Latitude == 0 {
		location.Latitude = user.Location.Latitude
	}

	if location.Longitude == 0 {
		location.Longitude = user.Location.Longitude
	}

	return location
}

func (s *ReportService) enrichAlertLocationFromReport(
	report *models.Report,
) models.AlertLocation {
	location := models.AlertLocation{
		Latitude:  report.Location.Latitude,
		Longitude: report.Location.Longitude,
		Address:   strings.TrimSpace(report.Location.Address),
		Country:   strings.TrimSpace(report.Location.Country),
		Region:    strings.TrimSpace(report.Location.Region),
	}

	if strings.TrimSpace(location.Country) != "" &&
		strings.TrimSpace(location.Region) != "" {
		return location
	}

	if s.userRepo == nil {
		return location
	}

	user, err := s.userRepo.FindUserByID(report.UserID)
	if err != nil || user == nil || user.Location == nil {
		return location
	}

	if strings.TrimSpace(location.Address) == "" {
		location.Address = strings.TrimSpace(user.Location.Address)
	}

	if strings.TrimSpace(location.Country) == "" {
		location.Country = strings.TrimSpace(user.Location.Country)
	}

	if strings.TrimSpace(location.Region) == "" {
		location.Region = strings.TrimSpace(user.Location.Region)
	}

	if location.Latitude == 0 {
		location.Latitude = user.Location.Latitude
	}

	if location.Longitude == 0 {
		location.Longitude = user.Location.Longitude
	}

	return location
}

func splitReportMedia(report *models.Report) ([]string, []string) {
	imageURLs := []string{}
	videoURLs := []string{}

	if report == nil {
		return imageURLs, videoURLs
	}

	for _, media := range report.Media {
		mediaURL := strings.TrimSpace(media.URL)
		mediaType := strings.ToLower(strings.TrimSpace(media.Type))

		if mediaURL == "" {
			continue
		}

		if mediaType == "video" {
			videoURLs = append(videoURLs, mediaURL)
			continue
		}

		if mediaType == "image" {
			imageURLs = append(imageURLs, mediaURL)
			continue
		}

		if isVideoURL(mediaURL) {
			videoURLs = append(videoURLs, mediaURL)
			continue
		}

		if isImageURL(mediaURL) {
			imageURLs = append(imageURLs, mediaURL)
		}
	}

	for _, mediaURL := range report.MediaURLs {
		cleanURL := strings.TrimSpace(mediaURL)

		if cleanURL == "" {
			continue
		}

		if isVideoURL(cleanURL) {
			videoURLs = append(videoURLs, cleanURL)
			continue
		}

		if isImageURL(cleanURL) {
			imageURLs = append(imageURLs, cleanURL)
			continue
		}

		videoURLs = append(videoURLs, cleanURL)
	}

	return uniqueStrings(imageURLs), uniqueStrings(videoURLs)
}

func isVideoURL(url string) bool {
	value := strings.ToLower(strings.TrimSpace(url))

	return strings.Contains(value, "/video/upload/") ||
		strings.HasSuffix(value, ".mp4") ||
		strings.HasSuffix(value, ".mov") ||
		strings.HasSuffix(value, ".webm") ||
		strings.HasSuffix(value, ".m3u8")
}

func isImageURL(url string) bool {
	value := strings.ToLower(strings.TrimSpace(url))

	return strings.Contains(value, "/image/upload/") ||
		strings.HasSuffix(value, ".jpg") ||
		strings.HasSuffix(value, ".jpeg") ||
		strings.HasSuffix(value, ".png") ||
		strings.HasSuffix(value, ".webp")
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}

	for _, value := range values {
		cleanValue := strings.TrimSpace(value)

		if cleanValue == "" {
			continue
		}

		if seen[cleanValue] {
			continue
		}

		seen[cleanValue] = true
		unique = append(unique, cleanValue)
	}

	return unique
}

func cleanStringList(values []string) []string {
	cleaned := []string{}

	for _, value := range values {
		cleanValue := strings.TrimSpace(value)

		if cleanValue == "" {
			continue
		}

		cleaned = append(cleaned, cleanValue)
	}

	return cleaned
}

func bestLocationName(location models.AlertLocation) string {
	if strings.TrimSpace(location.Address) != "" {
		return strings.TrimSpace(location.Address)
	}

	if strings.TrimSpace(location.Region) != "" {
		return strings.TrimSpace(location.Region)
	}

	if strings.TrimSpace(location.Country) != "" {
		return strings.TrimSpace(location.Country)
	}

	return ""
}

func buildApprovedReportAlertTitle(category string, location string) string {
	cleanCategory := titleCase(strings.TrimSpace(category))
	cleanLocation := strings.TrimSpace(location)

	if cleanCategory == "" {
		cleanCategory = "Incident"
	}

	if cleanLocation == "" {
		return cleanCategory + " Reported"
	}

	return cleanCategory + " Reported in " + cleanLocation
}

func buildApprovedReportAlertSummary(category string, location string) string {
	cleanCategory := strings.ToLower(strings.TrimSpace(category))
	cleanLocation := strings.TrimSpace(location)

	if cleanCategory == "" {
		cleanCategory = "incident"
	}

	if cleanLocation == "" {
		return "A verified " + cleanCategory + " report has been approved by administrators."
	}

	return "A verified " + cleanCategory + " report has been approved around " + cleanLocation + "."
}

func inferReportSeverity(category string, description string) string {
	value := strings.ToLower(category + " " + description)

	if strings.Contains(value, "death") ||
		strings.Contains(value, "dead") ||
		strings.Contains(value, "trapped") ||
		strings.Contains(value, "collapsed") ||
		strings.Contains(value, "severe") ||
		strings.Contains(value, "critical") {
		return "high"
	}

	if strings.Contains(value, "flood") ||
		strings.Contains(value, "fire") ||
		strings.Contains(value, "explosion") ||
		strings.Contains(value, "accident") ||
		strings.Contains(value, "violent") {
		return "high"
	}

	return "medium"
}

func inferReportPriorityScore(severity string, category string) int {
	cleanSeverity := strings.ToLower(strings.TrimSpace(severity))
	cleanCategory := strings.ToLower(strings.TrimSpace(category))

	if cleanSeverity == "critical" {
		return 95
	}

	if cleanSeverity == "high" {
		if strings.Contains(cleanCategory, "flood") ||
			strings.Contains(cleanCategory, "fire") ||
			strings.Contains(cleanCategory, "explosion") {
			return 85
		}

		return 78
	}

	if cleanSeverity == "low" {
		return 35
	}

	return 60
}

func inferReportPriorityLabel(priorityScore int) string {
	if priorityScore >= 85 {
		return "breaking"
	}

	if priorityScore >= 65 {
		return "serious"
	}

	if priorityScore >= 45 {
		return "watch"
	}

	return "notice"
}

func buildReportSafetyInstructions(category string) []string {
	value := strings.ToLower(strings.TrimSpace(category))

	if strings.Contains(value, "flood") {
		return []string{
			"Move to higher ground if your area is flooding.",
			"Do not walk or drive through floodwater.",
			"Avoid open drains, rivers, and fast-moving water.",
			"Follow updates from emergency authorities.",
		}
	}

	if strings.Contains(value, "fire") {
		return []string{
			"Move away from the fire area immediately.",
			"Avoid smoke and closed spaces.",
			"Call emergency services if people are trapped.",
			"Follow instructions from fire officers and local authorities.",
		}
	}

	if strings.Contains(value, "accident") {
		return []string{
			"Avoid the affected road or area.",
			"Do not crowd the accident scene.",
			"Allow emergency responders to access the area.",
			"Follow traffic and police directions.",
		}
	}

	return []string{
		"Avoid the affected area if possible.",
		"Stay alert and follow official safety updates.",
		"Share your location with someone you trust if you may be affected.",
	}
}

func buildReportAlertTags(category string, country string, region string) []string {
	tags := []string{
		strings.ToLower(strings.TrimSpace(category)),
		"community-report",
		"verified",
		strings.ToLower(strings.TrimSpace(country)),
		strings.ToLower(strings.TrimSpace(region)),
	}

	return uniqueStrings(tags)
}

func parseReportEventTime(value string, fallback time.Time) *time.Time {
	cleanValue := strings.TrimSpace(value)

	if cleanValue == "" {
		return &fallback
	}

	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.000000Z",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, layout := range layouts {
		parsedTime, err := time.Parse(layout, cleanValue)
		if err == nil {
			utcTime := parsedTime.UTC()
			return &utcTime
		}
	}

	return &fallback
}

func titleCase(value string) string {
	parts := strings.Fields(strings.TrimSpace(value))

	for index, part := range parts {
		if part == "" {
			continue
		}

		lowerPart := strings.ToLower(part)
		parts[index] = strings.ToUpper(lowerPart[:1]) + lowerPart[1:]
	}

	return strings.Join(parts, " ")
}
