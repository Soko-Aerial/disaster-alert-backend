package services

import (
	"errors"
	"strings"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmergencyContactService struct {
	contactRepo *repositories.EmergencyContactRepository
}

func NewEmergencyContactService(
	contactRepo *repositories.EmergencyContactRepository,
) *EmergencyContactService {
	return &EmergencyContactService{
		contactRepo: contactRepo,
	}
}

func (s *EmergencyContactService) CreateContact(
	userID primitive.ObjectID,
	req dto.CreateEmergencyContactRequest,
) (*models.EmergencyContact, error) {
	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)

	if name == "" {
		return nil, errors.New("contact name is required")
	}

	if phone == "" {
		return nil, errors.New("contact phone is required")
	}

	contactType := strings.TrimSpace(req.Type)
	if contactType == "" {
		contactType = "other"
	}

	contact := models.EmergencyContact{
		UserID:       userID,
		Name:         name,
		Phone:        phone,
		Email:        strings.TrimSpace(req.Email),
		Relationship: strings.TrimSpace(req.Relationship),
		Type:         contactType,
		Organization: strings.TrimSpace(req.Organization),
		Address:      strings.TrimSpace(req.Address),
		IsPrimary:    req.IsPrimary,
		IsGovernment: req.IsGovernment,
		IsActive:     true,
	}

	return s.contactRepo.Create(contact)
}

func (s *EmergencyContactService) GetUserContacts(
	userID primitive.ObjectID,
) ([]models.EmergencyContact, error) {
	return s.contactRepo.FindByUserID(userID)
}

func (s *EmergencyContactService) GetContactByID(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) (*models.EmergencyContact, error) {
	return s.contactRepo.FindByID(id, userID)
}

func (s *EmergencyContactService) UpdateContact(
	id primitive.ObjectID,
	userID primitive.ObjectID,
	req dto.UpdateEmergencyContactRequest,
) (*models.EmergencyContact, error) {
	update := bson.M{}

	if strings.TrimSpace(req.Name) != "" {
		update["name"] = strings.TrimSpace(req.Name)
	}

	if strings.TrimSpace(req.Phone) != "" {
		update["phone"] = strings.TrimSpace(req.Phone)
	}

	if strings.TrimSpace(req.Email) != "" {
		update["email"] = strings.TrimSpace(req.Email)
	}

	if strings.TrimSpace(req.Relationship) != "" {
		update["relationship"] = strings.TrimSpace(req.Relationship)
	}

	if strings.TrimSpace(req.Type) != "" {
		update["type"] = strings.TrimSpace(req.Type)
	}

	if strings.TrimSpace(req.Organization) != "" {
		update["organization"] = strings.TrimSpace(req.Organization)
	}

	if strings.TrimSpace(req.Address) != "" {
		update["address"] = strings.TrimSpace(req.Address)
	}

	if req.IsPrimary != nil {
		update["isPrimary"] = *req.IsPrimary
	}

	if req.IsGovernment != nil {
		update["isGovernment"] = *req.IsGovernment
	}

	if req.IsActive != nil {
		update["isActive"] = *req.IsActive
	}

	if len(update) == 0 {
		return nil, errors.New("no valid fields provided for update")
	}

	return s.contactRepo.Update(id, userID, update)
}

func (s *EmergencyContactService) DeleteContact(
	id primitive.ObjectID,
	userID primitive.ObjectID,
) error {
	return s.contactRepo.SoftDelete(id, userID)
}