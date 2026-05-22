package services

import (
	"errors"
	"strings"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserLocationService struct {
	userRepo *repositories.UserRepository
}

func NewUserLocationService(
	userRepo *repositories.UserRepository,
) *UserLocationService {
	return &UserLocationService{
		userRepo: userRepo,
	}
}

func (s *UserLocationService) GetUserLocation(
	userID primitive.ObjectID,
) (*models.UserLocation, error) {
	return s.userRepo.FindLocationByUserID(userID)
}

func (s *UserLocationService) UpdateUserLocation(
	userID primitive.ObjectID,
	req dto.UpdateUserLocationRequest,
) (*models.UserLocation, error) {
	if req.Latitude < -90 || req.Latitude > 90 {
		return nil, errors.New("latitude must be between -90 and 90")
	}

	if req.Longitude < -180 || req.Longitude > 180 {
		return nil, errors.New("longitude must be between -180 and 180")
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		source = "manual"
	}

	if source != "gps" && source != "manual" {
		return nil, errors.New("source must be gps or manual")
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "Selected Location"
	}

	location := models.UserLocation{
		Name:      name,
		Country:   strings.TrimSpace(req.Country),
		Region:    strings.TrimSpace(req.Region),
		Address:   strings.TrimSpace(req.Address),
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Source:    source,
		IsDefault: req.IsDefault,
	}

	return s.userRepo.UpdateLocationByUserID(userID, location)
}