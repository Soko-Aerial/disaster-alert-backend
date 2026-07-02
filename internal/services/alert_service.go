package services

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertService struct {
	alertRepo              *repositories.AlertRepository
	userRepo               *repositories.UserRepository
	notificationDispatcher NotificationDispatcher
	appNotificationService *AppNotificationService
	broadcaster            *websocket.Broadcaster
}

func NewAlertService(
	alertRepo *repositories.AlertRepository,
	userRepo *repositories.UserRepository,
	notificationDispatcher NotificationDispatcher,
	appNotificationService *AppNotificationService,
	broadcaster *websocket.Broadcaster,
) *AlertService {
	return &AlertService{
		alertRepo:              alertRepo,
		userRepo:               userRepo,
		notificationDispatcher: notificationDispatcher,
		appNotificationService: appNotificationService,
		broadcaster:            broadcaster,
	}
}

func (s *AlertService) CreateAlert(
	userID string,
	req dto.CreateAlertRequest,
) (*models.Alert, error) {
	creatorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now().UTC()

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	sourceType := strings.TrimSpace(req.SourceType)
	if sourceType == "" {
		sourceType = "internal"
	}

	sourceName := strings.TrimSpace(req.SourceName)
	if sourceName == "" {
		sourceName = "manual"
	}

	eventTime := parseOptionalTime(req.EventTime)

	expiresAt := parseOptionalTime(req.ExpiresAt)
	if expiresAt == nil {
		defaultExpiry := now.Add(7 * 24 * time.Hour)
		expiresAt = &defaultExpiry
	}

	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 20
	}

	confidence := req.Confidence
	if confidence <= 0 {
		confidence = 1.0
	}

	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = strings.TrimSpace(req.Description)
	}

	priorityScore := req.PriorityScore
	if priorityScore <= 0 {
		priorityScore = manualAlertPriorityScore(req.Severity, req.Category)
	}

	priorityLabel := strings.TrimSpace(req.PriorityLabel)
	if priorityLabel == "" {
		priorityLabel = manualAlertPriorityLabel(priorityScore)
	}

	alert := models.Alert{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Summary:     summary,
		Category:    strings.TrimSpace(req.Category),
		Severity:    strings.TrimSpace(req.Severity),
		Status:      status,
		Location: models.AlertLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   strings.TrimSpace(req.Address),
			Country:   strings.TrimSpace(req.Country),
			Region:    strings.TrimSpace(req.Region),
		},
		RadiusKm:           radiusKm,
		SafetyInstructions: req.SafetyInstructions,
		SourceType:         sourceType,
		SourceName:         sourceName,
		ExternalID:         strings.TrimSpace(req.ExternalID),
		SourceURL:          strings.TrimSpace(req.SourceURL),
		ImageURLs:          req.ImageURLs,
		VideoURLs:          req.VideoURLs,
		Tags:               normalizeAlertTags(req.Tags, req.Category, req.Severity, sourceName),
		PriorityScore:      priorityScore,
		PriorityLabel:      priorityLabel,
		IsBreaking:         priorityScore >= 85,
		IsVerified:         sourceType == "internal",
		CreatedBy:          &creatorID,
		EventTime:          eventTime,
		ExpiresAt:          expiresAt,
		Confidence:         confidence,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	createdAlert, err := s.alertRepo.Create(alert)
	if err != nil {
		return nil, err
	}

	if createdAlert.Status == "active" {
		go s.notifyUsersInAlertCountry(createdAlert)
	}

	if createdAlert.Status == "active" && s.broadcaster != nil {
		s.broadcaster.BroadcastAlertCreated(
			createdAlert.Location.Country,
			buildAlertPayload(createdAlert),
		)
	}

	return createdAlert, nil
}

func (s *AlertService) GetAlerts() ([]models.Alert, error) {
	return s.alertRepo.FindAll()
}

func (s *AlertService) GetActiveAlerts() ([]models.Alert, error) {
	return s.alertRepo.FindActive()
}

func (s *AlertService) GetActiveAlertsWithFilters(
	filter repositories.AlertFilter,
) ([]models.Alert, error) {
	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetLocalAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if strings.TrimSpace(country) == "" {
		return []models.Alert{}, nil
	}

	if limit <= 0 {
		limit = 50
	}

	return s.alertRepo.FindActiveWithFilters(repositories.AlertFilter{
		Country: country,
		Limit:   limit,
	})
}

func (s *AlertService) GetGlobalAlerts(
	userCountry string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	return s.alertRepo.FindActiveWithFilters(repositories.AlertFilter{
		ExcludeCountry: userCountry,
		Limit:          limit,
	})
}

func (s *AlertService) GetWeatherAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := repositories.AlertFilter{
		Category: "weather",
		Limit:    limit,
	}

	if strings.TrimSpace(country) != "" {
		filter.Country = country
	}

	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetHealthAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := repositories.AlertFilter{
		Category: "health",
		Limit:    limit,
	}

	if strings.TrimSpace(country) != "" {
		filter.Country = country
	}

	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetAlertByID(alertID string) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	return s.alertRepo.FindByID(objectID)
}

func (s *AlertService) UpdateAlertStatus(
	alertID string,
	status string,
) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	updatedAlert, err := s.alertRepo.UpdateStatus(objectID, status)
	if err != nil {
		return nil, err
	}

	if updatedAlert.Status == "active" {
		go s.notifyUsersInAlertCountry(updatedAlert)
	}

	if updatedAlert.Status == "active" && s.broadcaster != nil {
		s.broadcaster.BroadcastAlertApproved(
			updatedAlert.Location.Country,
			buildAlertPayload(updatedAlert),
		)
	}

	return updatedAlert, nil
}

func (s *AlertService) DeleteAlert(alertID string) error {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return errors.New("invalid alert id")
	}

	return s.alertRepo.Delete(objectID)
}

func (s *AlertService) GetNearbyAlerts(
	latitude float64,
	longitude float64,
	radiusKm float64,
) ([]models.Alert, error) {
	if radiusKm <= 0 {
		radiusKm = 100
	}

	return s.alertRepo.FindActiveNearby(latitude, longitude, radiusKm)
}

func (s *AlertService) GetNearbyAlertsWithFilters(
	latitude float64,
	longitude float64,
	radiusKm float64,
	filter repositories.AlertFilter,
) ([]models.Alert, error) {
	if radiusKm <= 0 {
		radiusKm = 100
	}

	return s.alertRepo.FindActiveNearbyWithFilters(
		latitude,
		longitude,
		radiusKm,
		filter,
	)
}

func (s *AlertService) GetCriticalGlobalAlerts(
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.alertRepo.FindCriticalGlobal(limit)
}

func (s *AlertService) notifyUsersInAlertCountry(alert *models.Alert) {
	if alert == nil {
		return
	}

	if s.userRepo == nil {
		return
	}

	country := strings.TrimSpace(alert.Location.Country)
	if country == "" {
		return
	}

	users, err := s.userRepo.FindUsersByCountry(country)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	title := strings.TrimSpace(alert.Title)
	if title == "" {
		title = "New Disaster Alert"
	}

	body := strings.TrimSpace(alert.Description)
	if body == "" {
		body = "A new emergency alert has been issued in your country."
	}

	data := map[string]string{
		"type":        "alert",
		"referenceId": alert.ID.Hex(),
		"alertId":     alert.ID.Hex(),
		"category":    alert.Category,
		"severity":    alert.Severity,
		"source":      alert.SourceName,
		"country":     alert.Location.Country,
		"latitude":    floatToString(alert.Location.Latitude),
		"longitude":   floatToString(alert.Location.Longitude),
	}

	userIDs := make([]primitive.ObjectID, 0, len(users))

	for _, user := range users {
		if user.ID.IsZero() {
			continue
		}

		userIDs = append(userIDs, user.ID)

		if s.notificationDispatcher != nil {
			s.notificationDispatcher.Dispatch(jobs.NotificationJob{
				TargetType: jobs.TargetUser,
				UserID:     user.ID.Hex(),
				Title:      title,
				Body:       body,
				Data:       data,
			})
		}
	}

	if len(userIDs) == 0 {
		return
	}

	if s.appNotificationService != nil {
		_ = s.appNotificationService.CreateManyForUsers(
			userIDs,
			title,
			body,
			"alert",
			alert.ID.Hex(),
			data,
		)
	}
}

func parseOptionalTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}

	return &parsed
}

func floatToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func buildAlertPayload(alert *models.Alert) map[string]interface{} {
	if alert == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":                 alert.ID.Hex(),
		"title":              alert.Title,
		"description":        alert.Description,
		"summary":            alert.Summary,
		"category":           alert.Category,
		"severity":           alert.Severity,
		"status":             alert.Status,
		"latitude":           alert.Location.Latitude,
		"longitude":          alert.Location.Longitude,
		"address":            alert.Location.Address,
		"country":            alert.Location.Country,
		"region":             alert.Location.Region,
		"radiusKm":           alert.RadiusKm,
		"safetyInstructions": alert.SafetyInstructions,
		"sourceType":         alert.SourceType,
		"sourceName":         alert.SourceName,
		"externalId":         alert.ExternalID,
		"sourceUrl":          alert.SourceURL,
		"imageUrls":          alert.ImageURLs,
		"videoUrls":          alert.VideoURLs,
		"tags":               alert.Tags,
		"priorityScore":      alert.PriorityScore,
		"priorityLabel":      alert.PriorityLabel,
		"isBreaking":         alert.IsBreaking,
		"isVerified":         alert.IsVerified,
		"eventTime":          alert.EventTime,
		"expiresAt":          alert.ExpiresAt,
		"confidence":         alert.Confidence,
		"lastSyncedAt":       alert.LastSyncedAt,
		"createdAt":          alert.CreatedAt,
		"updatedAt":          alert.UpdatedAt,
	}
}


func manualAlertPriorityScore(severity string, category string) int {
	score := 0

	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		score += 80
	case "high":
		score += 65
	case "medium":
		score += 45
	case "low":
		score += 25
	default:
		score += 35
	}

	switch strings.ToLower(strings.TrimSpace(category)) {
	case "earthquake", "flood", "fire", "weather", "health", "conflict":
		score += 10
	case "volcano", "drought":
		score += 8
	default:
		score += 5
	}

	if score > 100 {
		score = 100
	}

	return score
}

func manualAlertPriorityLabel(score int) string {
	switch {
	case score >= 85:
		return "breaking"
	case score >= 65:
		return "serious"
	case score >= 45:
		return "watch"
	default:
		return "low"
	}
}

func normalizeAlertTags(inputTags []string, category string, severity string, sourceName string) []string {
	seen := map[string]bool{}
	tags := []string{}

	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			return
		}

		seen[value] = true
		tags = append(tags, value)
	}

	for _, tag := range inputTags {
		add(tag)
	}

	add(category)
	add(severity)
	add(sourceName)

	return tags
}