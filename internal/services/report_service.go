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
}

func NewReportService(reportRepo *repositories.ReportRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
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