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

type SOSService struct {
	sosRepo                  *repositories.SOSRepository
	userRepo                 *repositories.UserRepository
	eventNotificationService *EventNotificationService
	broadcaster              *websocket.Broadcaster
}

func NewSOSService(
	sosRepo *repositories.SOSRepository,
	userRepo *repositories.UserRepository,
	eventNotificationService *EventNotificationService,
	broadcaster *websocket.Broadcaster,
) *SOSService {
	return &SOSService{
		sosRepo:                  sosRepo,
		userRepo:                 userRepo,
		eventNotificationService: eventNotificationService,
		broadcaster:              broadcaster,
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

	locationText := strings.TrimSpace(createdSOS.Location.Address)
	if locationText == "" {
		locationText = "Lat: " +
			floatToString(createdSOS.Location.Latitude) +
			", Lng: " +
			floatToString(createdSOS.Location.Longitude)
	}

	userData := s.buildUserSummary(createdSOS.UserID)

	if s.eventNotificationService != nil {
		go s.eventNotificationService.NotifyAdminsForSOS(
			createdSOS.ID,
			getUserDisplayName(userData),
			createdSOS.EmergencyType,
			locationText,
		)
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastSOSCreated(map[string]interface{}{
			"id":             createdSOS.ID.Hex(),
			"user":           userData,
			"emergencyType":  createdSOS.EmergencyType,
			"message":        createdSOS.Message,
			"latitude":       createdSOS.Location.Latitude,
			"longitude":      createdSOS.Location.Longitude,
			"address":        createdSOS.Location.Address,
			"accuracy":       createdSOS.Location.Accuracy,
			"status":         createdSOS.Status,
			"isLiveTracking": createdSOS.IsLiveTracking,
			"createdAt":      createdSOS.CreatedAt,
		})
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

	updatedSOS, err := s.sosRepo.UpdateStatus(objectID, status)
	if err != nil {
		return nil, err
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastSOSStatusUpdated(
			updatedSOS.UserID.Hex(),
			map[string]interface{}{
				"id":        updatedSOS.ID.Hex(),
				"userId":    updatedSOS.UserID.Hex(),
				"status":    updatedSOS.Status,
				"updatedAt": updatedSOS.UpdatedAt,
			},
		)
	}

	return updatedSOS, nil
}

func (s *SOSService) buildUserSummary(
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

func getUserDisplayName(userData map[string]interface{}) string {
	name, ok := userData["name"].(string)
	if ok && strings.TrimSpace(name) != "" {
		return name
	}

	email, ok := userData["email"].(string)
	if ok && strings.TrimSpace(email) != "" {
		return email
	}

	return "User"
}