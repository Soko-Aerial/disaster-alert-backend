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
		s.broadcaster.BroadcastSOSCreated(
			s.buildSOSResponse(createdSOS),
		)
	}

	return createdSOS, nil
}

func (s *SOSService) GetSOSRequests() ([]map[string]interface{}, error) {
	requests, err := s.sosRepo.FindAll()
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(requests))

	for i := range requests {
		response = append(
			response,
			s.buildSOSResponse(&requests[i]),
		)
	}

	return response, nil
}

func (s *SOSService) GetSOSByID(
	sosID string,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(sosID)
	if err != nil {
		return nil, errors.New("invalid sos request id")
	}

	sos, err := s.sosRepo.FindByID(objectID)
	if err != nil {
		return nil, err
	}

	return s.buildSOSResponse(sos), nil
}

func (s *SOSService) UpdateSOSStatus(
	sosID string,
	status string,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(sosID)
	if err != nil {
		return nil, errors.New("invalid sos request id")
	}

	updatedSOS, err := s.sosRepo.UpdateStatus(objectID, status)
	if err != nil {
		return nil, err
	}

	response := s.buildSOSResponse(updatedSOS)

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

	return response, nil
}

func (s *SOSService) buildSOSResponse(
	sos *models.SOSRequest,
) map[string]interface{} {
	if sos == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":             sos.ID.Hex(),
		"userId":         sos.UserID.Hex(),
		"user":           s.buildUserSummary(sos.UserID),
		"emergencyType":  sos.EmergencyType,
		"message":        sos.Message,
		"location":       sos.Location,
		"status":         sos.Status,
		"isLiveTracking": sos.IsLiveTracking,
		"createdAt":      sos.CreatedAt,
		"updatedAt":      sos.UpdatedAt,
	}
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