package services

import (
	"errors"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AppNotificationService struct {
	notificationRepo *repositories.AppNotificationRepository
	broadcaster      *websocket.Broadcaster
}

func NewAppNotificationService(
	notificationRepo *repositories.AppNotificationRepository,
	broadcaster *websocket.Broadcaster,
) *AppNotificationService {
	return &AppNotificationService{
		notificationRepo: notificationRepo,
		broadcaster:      broadcaster,
	}
}

func (s *AppNotificationService) GetUserNotifications(
	userID primitive.ObjectID,
	limit int,
) ([]models.AppNotification, error) {
	return s.notificationRepo.FindByRecipientID(userID, limit)
}

func (s *AppNotificationService) GetUnreadCount(
	userID primitive.ObjectID,
) (int64, error) {
	return s.notificationRepo.CountUnreadByRecipientID(userID)
}

func (s *AppNotificationService) MarkRead(
	userID primitive.ObjectID,
	req dto.MarkAppNotificationReadRequest,
) error {
	ids := make([]primitive.ObjectID, 0)

	for _, idString := range req.NotificationIDs {
		id, err := primitive.ObjectIDFromHex(idString)
		if err != nil {
			return errors.New("invalid notification id: " + idString)
		}

		ids = append(ids, id)
	}

	return s.notificationRepo.MarkRead(userID, ids)
}

func (s *AppNotificationService) MarkAllRead(
	userID primitive.ObjectID,
) error {
	return s.notificationRepo.MarkAllRead(userID)
}

func (s *AppNotificationService) CreateForUser(
	userID primitive.ObjectID,
	title string,
	body string,
	notificationType string,
	referenceID string,
	data map[string]string,
) (*models.AppNotification, error) {
	notification := models.AppNotification{
		RecipientID:   userID,
		RecipientRole: "user",
		Title:         title,
		Body:          body,
		Type:          notificationType,
		ReferenceID:   referenceID,
		Data:          data,
		IsRead:        false,
	}

	createdNotification, err := s.notificationRepo.Create(notification)
	if err != nil {
		return nil, err
	}

	s.broadcastNotification(createdNotification)

	return createdNotification, nil
}

func (s *AppNotificationService) CreateForAdmin(
	adminID primitive.ObjectID,
	title string,
	body string,
	notificationType string,
	referenceID string,
	data map[string]string,
) (*models.AppNotification, error) {
	notification := models.AppNotification{
		RecipientID:   adminID,
		RecipientRole: "admin",
		Title:         title,
		Body:          body,
		Type:          notificationType,
		ReferenceID:   referenceID,
		Data:          data,
		IsRead:        false,
	}

	createdNotification, err := s.notificationRepo.Create(notification)
	if err != nil {
		return nil, err
	}

	s.broadcastNotification(createdNotification)

	return createdNotification, nil
}

func (s *AppNotificationService) CreateManyForUsers(
	userIDs []primitive.ObjectID,
	title string,
	body string,
	notificationType string,
	referenceID string,
	data map[string]string,
) error {
	notifications := make([]models.AppNotification, 0, len(userIDs))

	for _, userID := range userIDs {
		notifications = append(notifications, models.AppNotification{
			RecipientID:   userID,
			RecipientRole: "user",
			Title:         title,
			Body:          body,
			Type:          notificationType,
			ReferenceID:   referenceID,
			Data:          data,
			IsRead:        false,
		})
	}

	if err := s.notificationRepo.CreateMany(notifications); err != nil {
		return err
	}

	for _, notification := range notifications {
		s.broadcastNotification(&notification)
	}

	return nil
}

func (s *AppNotificationService) broadcastNotification(
	notification *models.AppNotification,
) {
	if s.broadcaster == nil || notification == nil {
		return
	}

	s.broadcaster.BroadcastNotificationCreated(
		notification.RecipientID.Hex(),
		map[string]interface{}{
			"id":            notification.ID.Hex(),
			"recipientId":   notification.RecipientID.Hex(),
			"recipientRole": notification.RecipientRole,
			"title":         notification.Title,
			"body":          notification.Body,
			"type":          notification.Type,
			"referenceId":   notification.ReferenceID,
			"data":          notification.Data,
			"isRead":        notification.IsRead,
			"createdAt":     notification.CreatedAt,
			"updatedAt":     notification.UpdatedAt,
		},
	)
}