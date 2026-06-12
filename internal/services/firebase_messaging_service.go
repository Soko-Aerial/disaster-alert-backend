package services

import (
	"context"
	"errors"
	"time"

	"disaster_alert_backend/internal/repositories"

	"firebase.google.com/go/v4/messaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type FirebaseMessagingService struct {
	messagingClient *messaging.Client
	fcmTokenRepo    *repositories.FCMTokenRepository
}

func NewFirebaseMessagingService(
	messagingClient *messaging.Client,
	fcmTokenRepo *repositories.FCMTokenRepository,
) *FirebaseMessagingService {
	return &FirebaseMessagingService{
		messagingClient: messagingClient,
		fcmTokenRepo:    fcmTokenRepo,
	}
}

func (s *FirebaseMessagingService) SendToToken(
	token string,
	title string,
	body string,
	data map[string]string,
) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
	}

	response, err := s.messagingClient.Send(ctx, message)

	if err != nil {
		return "", err
	}

	return response, nil
}

func (s *FirebaseMessagingService) SendToUser(
	userID string,
	title string,
	body string,
	data map[string]string,
) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return errors.New("invalid user id")
	}

	tokens, err := s.fcmTokenRepo.FindActiveTokensByUserID(objectID)
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return errors.New("no active FCM tokens found for user")
	}

	successCount := 0
	var lastErr error

	for _, token := range tokens {
		_, err := s.SendToToken(token, title, body, data)
		if err != nil {
			lastErr = err
			continue
		}

		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	return nil
}

func (s *FirebaseMessagingService) SendToAll(
	title string,
	body string,
	data map[string]string,
) error {
	tokens, err := s.fcmTokenRepo.FindAllActiveTokens()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return errors.New("no active FCM tokens found")
	}

	successCount := 0
	var lastErr error

	for _, token := range tokens {
		_, err := s.SendToToken(token, title, body, data)
		if err != nil {
			lastErr = err
			continue
		}

		successCount++
	}

	if successCount == 0 && lastErr != nil {
		return lastErr
	}

	return nil
}