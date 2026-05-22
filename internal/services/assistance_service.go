package services

import (
	"errors"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AssistanceService struct {
	assistanceRepo *repositories.AssistanceRepository
}

func NewAssistanceService(
	assistanceRepo *repositories.AssistanceRepository,
) *AssistanceService {
	return &AssistanceService{
		assistanceRepo: assistanceRepo,
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

	now := time.Now()

	assistanceRequest := models.AssistanceRequest{
		UserID:             objectID,
		AssistanceType:     req.AssistanceType,
		UrgencyLevel:        req.UrgencyLevel,
		AffectedIndividuals: req.AffectedIndividuals,
		OtherInformation:    req.OtherInformation,
		Location: models.AssistanceLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   req.Address,
		},
		Status:    "pending",
		CreatedAt: now,
		UpdatedAt: now,
	}

	return s.assistanceRepo.Create(assistanceRequest)
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