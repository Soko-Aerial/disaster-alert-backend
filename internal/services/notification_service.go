package services

import (
	"errors"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type NotificationService struct {
	fcmTokenRepo *repositories.FCMTokenRepository
}

func NewNotificationService(
	fcmTokenRepo *repositories.FCMTokenRepository,
) *NotificationService {
	return &NotificationService{
		fcmTokenRepo: fcmTokenRepo,
	}
}

func (s *NotificationService) SaveFCMToken(
	userID string,
	req dto.SaveFCMTokenRequest,
) error {
	objectID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return errors.New("invalid user id")
	}

	deviceType := req.DeviceType

	if deviceType == "" {
		deviceType = "android"
	}

	now := time.Now()

	fcmToken := models.FCMToken{
		UserID:     objectID,
		Token:      req.Token,
		DeviceType: deviceType,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	return s.fcmTokenRepo.SaveOrUpdateToken(fcmToken)
}