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
) (map[string]interface{}, error) {
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

	response := s.buildAssistanceResponse(createdRequest)

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
		s.broadcaster.BroadcastAssistanceCreated(response)
	}

	return response, nil
}

func (s *AssistanceService) GetAssistanceRequests() ([]map[string]interface{}, error) {
	requests, err := s.assistanceRepo.FindAll()
	if err != nil {
		return nil, err
	}

	response := make([]map[string]interface{}, 0, len(requests))

	for i := range requests {
		response = append(response, s.buildAssistanceResponse(&requests[i]))
	}

	return response, nil
}

func (s *AssistanceService) GetAssistanceRequestByID(
	requestID string,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, errors.New("invalid assistance request id")
	}

	request, err := s.assistanceRepo.FindByID(objectID)
	if err != nil {
		return nil, err
	}

	return s.buildAssistanceResponse(request), nil
}

func (s *AssistanceService) UpdateAssistanceStatus(
	requestID string,
	status string,
) (map[string]interface{}, error) {
	objectID, err := primitive.ObjectIDFromHex(requestID)
	if err != nil {
		return nil, errors.New("invalid assistance request id")
	}

	cleanStatus := strings.ToLower(strings.TrimSpace(status))
	if !isValidAssistanceStatus(cleanStatus) {
		return nil, errors.New("invalid assistance status")
	}

	updatedRequest, err := s.assistanceRepo.UpdateStatus(objectID, cleanStatus)
	if err != nil {
		return nil, err
	}

	response := s.buildAssistanceResponse(updatedRequest)

	statusMessage := getAssistanceStatusMessage(cleanStatus)
	response["message"] = statusMessage

	if s.eventNotificationService != nil {
		go s.eventNotificationService.NotifyUserForAssistanceStatus(
			updatedRequest.UserID,
			updatedRequest.ID,
			cleanStatus,
			statusMessage,
		)
	}

	if s.broadcaster != nil {
		s.broadcaster.BroadcastAssistanceStatusUpdated(
			updatedRequest.UserID.Hex(),
			response,
		)
	}

	return response, nil
}

func (s *AssistanceService) buildAssistanceResponse(
	request *models.AssistanceRequest,
) map[string]interface{} {
	if request == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":                  request.ID.Hex(),
		"userId":              request.UserID.Hex(),
		"user":                s.buildUserSummary(request.UserID),
		"assistanceType":      request.AssistanceType,
		"urgencyLevel":        request.UrgencyLevel,
		"affectedIndividuals": request.AffectedIndividuals,
		"otherInformation":    request.OtherInformation,
		"location":            request.Location,
		"status":              request.Status,
		"createdAt":           request.CreatedAt,
		"updatedAt":           request.UpdatedAt,
	}
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

func isValidAssistanceStatus(status string) bool {
	switch status {
	case "pending", "accepted", "en_route", "arrived", "completed", "cancelled", "rejected":
		return true
	default:
		return false
	}
}

func getAssistanceStatusMessage(status string) string {
	switch status {
	case "pending":
		return "Your assistance request is pending. Please stay calm while we review it."
	case "accepted":
		return "Your assistance request has been accepted. Help is being arranged."
	case "en_route":
		return "Assistance is on the way. Keep your phone available and stay in a safe location."
	case "arrived":
		return "Help has arrived at or near your location."
	case "completed":
		return "Your assistance request has been completed."
	case "cancelled":
		return "Your assistance request has been cancelled."
	case "rejected":
		return "Your assistance request could not be accepted at this time."
	default:
		return "Your assistance request status has been updated."
	}
}