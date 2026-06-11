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

type AssistanceService struct {
	assistanceRepo           *repositories.AssistanceRepository
	userRepo                 *repositories.UserRepository
	eventNotificationService *EventNotificationService
	broadcaster              *websocket.Broadcaster
}

func NewAssistanceService(
	assistanceRepo *repositories.AssistanceRepository,
	userRepo *repositories.UserRepository,
	eventNotificationService *EventNotificationService,
	broadcaster *websocket.Broadcaster,
) *AssistanceService {
	return &AssistanceService{
		assistanceRepo:           assistanceRepo,
		userRepo:                 userRepo,
		eventNotificationService: eventNotificationService,
		broadcaster:              broadcaster,
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
		UserID:              objectID,
		AssistanceType:      strings.TrimSpace(req.AssistanceType),
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

	userData := s.buildUserSummary(createdRequest.UserID)

	if s.eventNotificationService != nil {
		go s.eventNotificationService.NotifyAdminsForAssistance(
			createdRequest.ID,
			getUserDisplayName(userData),
			assistanceType,
			locationText,
		)
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastAssistanceCreated(map[string]interface{}{
			"id":                  createdRequest.ID.Hex(),
			"user":                userData,
			"assistanceType":      createdRequest.AssistanceType,
			"urgencyLevel":        createdRequest.UrgencyLevel,
			"affectedIndividuals": createdRequest.AffectedIndividuals,
			"otherInformation":    createdRequest.OtherInformation,
			"latitude":            createdRequest.Location.Latitude,
			"longitude":           createdRequest.Location.Longitude,
			"address":             createdRequest.Location.Address,
			"status":              createdRequest.Status,
			"createdAt":           createdRequest.CreatedAt,
		})
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

func (s *AssistanceService) buildUserSummary(
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