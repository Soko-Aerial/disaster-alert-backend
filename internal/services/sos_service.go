package services

import (
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SOSService struct {
	sosRepo                  *repositories.SOSRepository
	eventNotificationService *EventNotificationService
}

func NewSOSService(
	sosRepo *repositories.SOSRepository,
	eventNotificationService *EventNotificationService,
) *SOSService {
	return &SOSService{
		sosRepo:                  sosRepo,
		eventNotificationService: eventNotificationService,
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

	now := time.Now().UTC()

	sos := models.SOSRequest{
		UserID:        objectID,
		EmergencyType: strings.TrimSpace(req.EmergencyType),
		Message:       strings.TrimSpace(req.Message),
		Location: models.SOSLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   strings.TrimSpace(req.Address),
			Accuracy:  req.Accuracy,
		},
		Status:         "active",
		IsLiveTracking: req.IsLiveTracking,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	createdSOS, err := s.sosRepo.Create(sos)
	if err != nil {
		return nil, err
	}

	if s.eventNotificationService != nil {
		locationText := strings.TrimSpace(createdSOS.Location.Address)
		if locationText == "" {
			locationText = "Lat: " +
				floatToString(createdSOS.Location.Latitude) +
				", Lng: " +
				floatToString(createdSOS.Location.Longitude)
		}

		go s.eventNotificationService.NotifyAdminsForSOS(
			createdSOS.ID,
			"User",
			createdSOS.EmergencyType,
			locationText,
		)
	}

	return createdSOS, nil
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