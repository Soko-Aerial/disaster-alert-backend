package services

import (
	"errors"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AlertPreferenceService struct {
	preferenceRepo *repositories.AlertPreferenceRepository
}

func NewAlertPreferenceService(
	preferenceRepo *repositories.AlertPreferenceRepository,
) *AlertPreferenceService {
	return &AlertPreferenceService{
		preferenceRepo: preferenceRepo,
	}
}

func (s *AlertPreferenceService) GetUserPreferences(
	userID primitive.ObjectID,
) (*models.AlertPreference, error) {
	preference, err := s.preferenceRepo.FindByUserID(userID)
	if err == nil {
		return preference, nil
	}

	if errors.Is(err, mongo.ErrNoDocuments) {
		defaultPreference := s.defaultPreferences(userID)
		return &defaultPreference, nil
	}

	return nil, err
}

func (s *AlertPreferenceService) UpdateUserPreferences(
	userID primitive.ObjectID,
	req dto.UpdateAlertPreferenceRequest,
) (*models.AlertPreference, error) {
	currentPreference, err := s.GetUserPreferences(userID)
	if err != nil {
		return nil, err
	}

	if req.Fire != nil {
		currentPreference.Fire = *req.Fire
	}

	if req.Flood != nil {
		currentPreference.Flood = *req.Flood
	}

	if req.Weather != nil {
		currentPreference.Weather = *req.Weather
	}

	if req.Earthquake != nil {
		currentPreference.Earthquake = *req.Earthquake
	}

	if req.Health != nil {
		currentPreference.Health = *req.Health
	}

	if req.Conflict != nil {
		currentPreference.Conflict = *req.Conflict
	}

	if req.Drought != nil {
		currentPreference.Drought = *req.Drought
	}

	if req.Protests != nil {
		currentPreference.Protests = *req.Protests
	}

	if req.Robbery != nil {
		currentPreference.Robbery = *req.Robbery
	}

	if req.Munitions != nil {
		currentPreference.Munitions = *req.Munitions
	}

	if req.Galamsey != nil {
		currentPreference.Galamsey = *req.Galamsey
	}

	if req.UnverifiedActivity != nil {
		currentPreference.UnverifiedActivity = *req.UnverifiedActivity
	}

	if req.CriticalAlerts != nil {
		currentPreference.CriticalAlerts = *req.CriticalAlerts
	}

	return s.preferenceRepo.UpsertByUserID(userID, *currentPreference)
}

func (s *AlertPreferenceService) defaultPreferences(
	userID primitive.ObjectID,
) models.AlertPreference {
	return models.AlertPreference{
		UserID: userID,

		Fire:               true,
		Flood:              true,
		Weather:            true,
		Earthquake:         true,
		Health:             true,
		Conflict:           true,
		Drought:            true,
		Protests:           true,
		Robbery:            true,
		Munitions:          true,
		Galamsey:           true,
		UnverifiedActivity: true,
		CriticalAlerts:     true,
	}
}