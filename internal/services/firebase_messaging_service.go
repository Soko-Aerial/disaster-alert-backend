package services

import (
	"context"
	"errors"
	"strconv"
	"time"

	"disaster_alert_backend/internal/observability"
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
		observability.Error(ctx, "FCM send to token failed", err, observability.Fields{
			"module":       "firebase_messaging",
			"target_type":  "token",
			"title":        title,
			"token_prefix": safeTokenPrefix(token),
		})

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
	ctx := context.Background()

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		observability.Warn(ctx, "FCM send to user failed because user ID is invalid", observability.Fields{
			"module":      "firebase_messaging",
			"target_type": "user",
			"title":       title,
		})

		return errors.New("invalid user id")
	}

	tokens, err := s.fcmTokenRepo.FindActiveTokensByUserID(objectID)
	if err != nil {
		observability.Error(ctx, "Failed to fetch active FCM tokens for user", err, observability.Fields{
			"module":      "firebase_messaging",
			"target_type": "user",
			"user_id":     userID,
			"title":       title,
		})

		return err
	}

	if len(tokens) == 0 {
		observability.Warn(ctx, "No active FCM tokens found for user", observability.Fields{
			"module":      "firebase_messaging",
			"target_type": "user",
			"user_id":     userID,
			"title":       title,
		})

		return errors.New("no active FCM tokens found for user")
	}

	successCount := 0
	failureCount := 0
	var lastErr error

	for _, token := range tokens {
		_, err := s.SendToToken(token, title, body, data)
		if err != nil {
			lastErr = err
			failureCount++
			continue
		}

		successCount++
	}

	if failureCount > 0 {
		observability.Warn(ctx, "Some FCM user tokens failed", observability.Fields{
			"module":        "firebase_messaging",
			"target_type":   "user",
			"user_id":       userID,
			"title":         title,
			"success_count": strconv.Itoa(successCount),
			"failure_count": strconv.Itoa(failureCount),
			"token_count":   strconv.Itoa(len(tokens)),
		})
	}

	if successCount == 0 && lastErr != nil {
		observability.Error(ctx, "FCM send to user failed for all tokens", lastErr, observability.Fields{
			"module":        "firebase_messaging",
			"target_type":   "user",
			"user_id":       userID,
			"title":         title,
			"failure_count": strconv.Itoa(failureCount),
			"token_count":   strconv.Itoa(len(tokens)),
		})

		return lastErr
	}

	return nil
}

func (s *FirebaseMessagingService) SendToAll(
	title string,
	body string,
	data map[string]string,
) error {
	ctx := context.Background()

	tokens, err := s.fcmTokenRepo.FindAllActiveTokens()
	if err != nil {
		observability.Error(ctx, "Failed to fetch all active FCM tokens", err, observability.Fields{
			"module":      "firebase_messaging",
			"target_type": "all",
			"title":       title,
		})

		return err
	}

	if len(tokens) == 0 {
		observability.Warn(ctx, "No active FCM tokens found", observability.Fields{
			"module":      "firebase_messaging",
			"target_type": "all",
			"title":       title,
		})

		return errors.New("no active FCM tokens found")
	}

	successCount := 0
	failureCount := 0
	var lastErr error

	for _, token := range tokens {
		_, err := s.SendToToken(token, title, body, data)
		if err != nil {
			lastErr = err
			failureCount++
			continue
		}

		successCount++
	}

	if failureCount > 0 {
		observability.Warn(ctx, "Some broadcast FCM tokens failed", observability.Fields{
			"module":        "firebase_messaging",
			"target_type":   "all",
			"title":         title,
			"success_count": strconv.Itoa(successCount),
			"failure_count": strconv.Itoa(failureCount),
			"token_count":   strconv.Itoa(len(tokens)),
		})
	}

	if successCount == 0 && lastErr != nil {
		observability.Error(ctx, "Broadcast FCM send failed for all tokens", lastErr, observability.Fields{
			"module":        "firebase_messaging",
			"target_type":   "all",
			"title":         title,
			"failure_count": strconv.Itoa(failureCount),
			"token_count":   strconv.Itoa(len(tokens)),
		})

		return lastErr
	}

	return nil
}

func safeTokenPrefix(token string) string {
	if len(token) <= 10 {
		return "short-token"
	}

	return token[:6] + "..." + token[len(token)-4:]
}
