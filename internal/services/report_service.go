package services

import (
	"errors"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReportService struct {
	reportRepo *repositories.ReportRepository
	alertRepo  *repositories.AlertRepository
}

func NewReportService(reportRepo *repositories.ReportRepository, alertRepo *repositories.AlertRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		alertRepo:  alertRepo,
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

	now := time.Now()

	report := models.Report{
		UserID:           objectID,
		Category:         req.Category,
		Description:      req.Description,
		TimeOfOccurrence: req.TimeOfOccurrence,
		Location: models.ReportLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   req.Address,
			Country:   req.Country,
			Region:    req.Region,
		},
		MediaURLs: req.MediaURLs,
		Media:     req.Media,
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.reportRepo.Create(report)
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

	return createdAlert, nil
}