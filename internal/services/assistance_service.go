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

type AssistanceService struct {
	assistanceRepo          *repositories.AssistanceRepository
	eventNotificationService *EventNotificationService
}

func NewAssistanceService(
	assistanceRepo *repositories.AssistanceRepository,
	eventNotificationService *EventNotificationService,
) *AssistanceService {
	return &AssistanceService{
		assistanceRepo:          assistanceRepo,
		eventNotificationService: eventNotificationService,
	}
}

func (s *AssistanceService) CreateAssistanceRequest(
	userID string,
	req dto.CreateAssistanceRequest,
) (*models.AssistanceRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now().UTC()

	assistanceRequest := models.AssistanceRequest{
		UserID:             objectID,
		AssistanceType:     strings.TrimSpace(req.AssistanceType),
		UrgencyLevel:        strings.TrimSpace(req.UrgencyLevel),
		AffectedIndividuals: req.AffectedIndividuals,
		OtherInformation:    strings.TrimSpace(req.OtherInformation),
		Location: models.AssistanceLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   strings.TrimSpace(req.Address),
		},
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdRequest, err := s.assistanceRepo.Create(assistanceRequest)
	if err != nil {
		return nil, err
	}

	if s.eventNotificationService != nil {
		locationText := strings.TrimSpace(createdRequest.Location.Address)
		if locationText == "" {
			locationText = "Lat: " +
				floatToString(createdRequest.Location.Latitude) +
				", Lng: " +
				floatToString(createdRequest.Location.Longitude)
		}

		assistanceType := createdRequest.AssistanceType
		if assistanceType == "" {
			assistanceType = "assistance"
		}

		go s.eventNotificationService.NotifyAdminsForAssistance(
			createdRequest.ID,
			"User",
			assistanceType,
			locationText,
		)
	}

	return createdRequest, nil
}

func (s *AssistanceService) GetAssistanceRequests() ([]models.AssistanceRequest, error) {
	return s.assistanceRepo.FindAll()
}

func (s *AssistanceService) GetAssistanceRequestByID(
	requestID string,
) (*models.AssistanceRequest, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, errors.New("invalid assistance request id")
	}

	return s.assistanceRepo.FindByID(objectID)
}