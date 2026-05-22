package services

import (
	"errors"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SOSService struct {
	sosRepo *repositories.SOSRepository
}

func NewSOSService(sosRepo *repositories.SOSRepository) *SOSService {
	return &SOSService{
		sosRepo: sosRepo,
	}
}

func (s *SOSService) CreateSOSRequest(
	userID string,
	req dto.CreateSOSRequest,
) (*models.SOSRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now()

	sos := models.SOSRequest{
		UserID:        objectID,
		EmergencyType: req.EmergencyType,
		Message:       req.Message,
		Location: models.SOSLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   req.Address,
			Accuracy:  req.Accuracy,
		},
		Status:         "active",
		IsLiveTracking: req.IsLiveTracking,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	return s.sosRepo.Create(sos)
}

func (s *SOSService) GetSOSRequests() ([]models.SOSRequest, error) {
	return s.sosRepo.FindAll()
}

func (s *SOSService) GetSOSByID(sosID string) (*models.SOSRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(sosID)
	if err != nil {
		return nil, errors.New("invalid sos request id")
	}

	return s.sosRepo.FindByID(objectID)
}

func (s *SOSService) UpdateSOSStatus(
	sosID string,
	status string,
) (*models.SOSRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(sosID)
	if err != nil {
		return nil, errors.New("invalid sos request id")
	}

	return s.sosRepo.UpdateStatus(objectID, status)
}