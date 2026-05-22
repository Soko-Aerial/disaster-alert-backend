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

type EmergencyMessageService struct {
	messageRepo *repositories.EmergencyMessageRepository
	contactRepo *repositories.EmergencyContactRepository
}

func NewEmergencyMessageService(
	messageRepo *repositories.EmergencyMessageRepository,
	contactRepo *repositories.EmergencyContactRepository,
) *EmergencyMessageService {
	return &EmergencyMessageService{
		messageRepo: messageRepo,
		contactRepo: contactRepo,
	}
}

func (s *EmergencyMessageService) CreateMessage(
	userID primitive.ObjectID,
	req dto.CreateEmergencyMessageRequest,
) (*models.EmergencyMessage, error) {
	messageText := strings.TrimSpace(req.Message)
	if messageText == "" {
		return nil, errors.New("message is required")
	}

	messageType := strings.TrimSpace(req.Type)
	if messageType == "" {
		messageType = "free_text"
	}

	message := models.EmergencyMessage{
		UserID:  userID,
		Title:   strings.TrimSpace(req.Title),
		Message: messageText,
		Type:    messageType,
		Status:  "draft",
	}

	if req.Location != nil {
		message.Location = &models.EmergencyMessageLocation{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Address:   strings.TrimSpace(req.Location.Address),
		}
	}

	return s.messageRepo.Create(message)
}

func (s *EmergencyMessageService) GetUserMessages(
	userID primitive.ObjectID,
) ([]models.EmergencyMessage, error) {
	return s.messageRepo.FindByUserID(userID)
}

func (s *EmergencyMessageService) GetMessageByID(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) (*models.EmergencyMessage, error) {
	return s.messageRepo.FindByID(id, userID)
}

func (s *EmergencyMessageService) DeleteMessage(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) error {
	return s.messageRepo.Delete(id, userID)
}

func (s *EmergencyMessageService) SendFreeMessage(
	userID primitive.ObjectID,
	req dto.SendEmergencyMessageRequest,
) (*models.EmergencyMessage, error) {
	messageText := strings.TrimSpace(req.Message)
	if messageText == "" {
		return nil, errors.New("message is required")
	}

	contacts, err := s.resolveContacts(userID, req)
	if err != nil {
		return nil, err
	}

	if len(contacts) == 0 {
		return nil, errors.New("no emergency contacts selected")
	}

	now := time.Now().UTC()

	message := models.EmergencyMessage{
		UserID:   userID,
		Title:    strings.TrimSpace(req.Title),
		Message:  messageText,
		Type:     "free_text",
		Status:   "pending",
		Contacts: contacts,
		SentAt:   &now,
	}

	if req.Location != nil {
		message.Location = &models.EmergencyMessageLocation{
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Address:   strings.TrimSpace(req.Location.Address),
		}
	}

	// For now, we only store/queue the message.
	// Later, SMS/WhatsApp/Email provider will be called here.
	return s.messageRepo.Create(message)
}

func (s *EmergencyMessageService) resolveContacts(
	userID primitive.ObjectID,
	req dto.SendEmergencyMessageRequest,
) ([]models.EmergencyMessageContactSnapshot, error) {
	userContacts, err := s.contactRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	if len(userContacts) == 0 {
		return []models.EmergencyMessageContactSnapshot{}, nil
	}

	selectedMap := map[primitive.ObjectID]bool{}

	if !req.SendToAllContacts {
		for _, contactIDString := range req.ContactIDs {
			contactID, err := primitive.ObjectIDFromHex(contactIDString)
			if err != nil {
				return nil, errors.New("invalid contact ID: " + contactIDString)
			}

			selectedMap[contactID] = true
		}
	}

	snapshots := make([]models.EmergencyMessageContactSnapshot, 0)

	for _, contact := range userContacts {
		if !req.SendToAllContacts {
			if !selectedMap[contact.ID] {
				continue
			}
		}

		snapshot := models.EmergencyMessageContactSnapshot{
			ContactID:    contact.ID,
			Name:         contact.Name,
			Phone:        contact.Phone,
			Email:        contact.Email,
			Relationship: contact.Relationship,
			Type:         contact.Type,
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}