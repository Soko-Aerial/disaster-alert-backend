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
) (*models.Report, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now().UTC()

	report := models.Report{
		UserID:           objectID,
		Category:         strings.TrimSpace(req.Category),
		Description:      strings.TrimSpace(req.Description),
		TimeOfOccurrence: req.TimeOfOccurrence,
		Location: models.ReportLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   strings.TrimSpace(req.Address),
			Country:   strings.TrimSpace(req.Country),
			Region:    strings.TrimSpace(req.Region),
		},
		MediaURLs: req.MediaURLs,
		Media:     req.Media,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
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
	if reportType == "" {
		reportType = "incident"
	}

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
		s.broadcaster.BroadcastReportCreated(map[string]interface{}{
			"id":               createdReport.ID.Hex(),
			"user":             userData,
			"category":         createdReport.Category,
			"description":      createdReport.Description,
			"timeOfOccurrence": createdReport.TimeOfOccurrence,
			"latitude":         createdReport.Location.Latitude,
			"longitude":        createdReport.Location.Longitude,
			"address":          createdReport.Location.Address,
			"country":          createdReport.Location.Country,
			"region":           createdReport.Location.Region,
			"mediaUrls":        createdReport.MediaURLs,
			"media":            createdReport.Media,
			"status":           createdReport.Status,
			"createdAt":        createdReport.CreatedAt,
		})
	}

	return createdReport, nil
}

func (s *ReportService) GetReports() ([]models.Report, error) {
	return s.reportRepo.FindAll()
}

func (s *ReportService) GetReportByID(reportID string) (*models.Report, error) {
	objectID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report id")
	}

	return s.reportRepo.FindByID(objectID)
}

func (s *ReportService) ApproveReport(
	reportID string,
) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, errors.New("invalid report id")
	}

	report, err := s.reportRepo.FindByID(objectID)
	if err != nil {
		return nil, errors.New("report not found")
	}

	if report.Status == "approved" {
		return nil, errors.New("report already approved")
	}

	approvedReport, err := s.reportRepo.UpdateStatus(objectID, "approved")
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)

	alert := models.Alert{
		Title:       approvedReport.Category,
		Description: approvedReport.Description,
		Category:    approvedReport.Category,
		Severity:    "medium",
		Status:      "active",
		Location: models.AlertLocation{
			Latitude:  approvedReport.Location.Latitude,
			Longitude: approvedReport.Location.Longitude,
			Address:   approvedReport.Location.Address,
			Country:   approvedReport.Location.Country,
			Region:    approvedReport.Location.Region,
		},
		RadiusKm:       10,
		SourceType:     "user_report",
		SourceName:     "User Report",
		LinkedReportID: &approvedReport.ID,
		EventTime:      &now,
		ExpiresAt:      &expiresAt,
		Confidence:     0.75,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	createdAlert, err := s.alertRepo.Create(alert)
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
				"id":          createdAlert.ID.Hex(),
				"title":       createdAlert.Title,
				"description": createdAlert.Description,
				"category":    createdAlert.Category,
				"severity":    createdAlert.Severity,
				"status":      createdAlert.Status,
				"latitude":    createdAlert.Location.Latitude,
				"longitude":   createdAlert.Location.Longitude,
				"address":     createdAlert.Location.Address,
				"country":     createdAlert.Location.Country,
				"region":      createdAlert.Location.Region,
				"radiusKm":    createdAlert.RadiusKm,
				"sourceType":  createdAlert.SourceType,
				"sourceName":  createdAlert.SourceName,
				"createdAt":   createdAlert.CreatedAt,
			},
		)
	}

	return createdAlert, nil
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