package services

import (
	"errors"
	"strconv"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertService struct {
	alertRepo              *repositories.AlertRepository
	notificationDispatcher NotificationDispatcher
}

func NewAlertService(
	alertRepo *repositories.AlertRepository,
	notificationDispatcher NotificationDispatcher,
) *AlertService {
	return &AlertService{
		alertRepo:              alertRepo,
		notificationDispatcher: notificationDispatcher,
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

	status := req.Status
	if status == "" {
		status = "active"
	}

	sourceType := req.SourceType
	if sourceType == "" {
		sourceType = "internal"
	}

	sourceName := req.SourceName
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

	alert := models.Alert{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Status:      status,
		Location: models.AlertLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   req.Address,
			Country:   req.Country,
			Region:    req.Region,
		},
		RadiusKm:           radiusKm,
		SafetyInstructions: req.SafetyInstructions,
		SourceType:         sourceType,
		SourceName:         sourceName,
		ExternalID:         req.ExternalID,
		SourceURL:          req.SourceURL,
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
		s.queueAlertNotification(createdAlert)
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
	if country == "" {
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

	if country != "" {
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

	if country != "" {
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
		s.queueAlertNotification(updatedAlert)
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

func (s *AlertService) queueAlertNotification(alert *models.Alert) {
	if s.notificationDispatcher == nil {
		return
	}

	title := alert.Title
	body := alert.Description

	if body == "" {
		body = "New disaster alert near your area"
	}

	queued := s.notificationDispatcher.Dispatch(jobs.NotificationJob{
		TargetType: jobs.TargetAll,
		Title:      title,
		Body:       body,
		Data: map[string]string{
			"type":      "alert",
			"alertId":   alert.ID.Hex(),
			"category":  alert.Category,
			"severity":  alert.Severity,
			"source":    alert.SourceName,
			"latitude":  floatToString(alert.Location.Latitude),
			"longitude": floatToString(alert.Location.Longitude),
		},
	})

	if !queued {
		return
	}
}

func parseOptionalTime(value string) *time.Time {
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