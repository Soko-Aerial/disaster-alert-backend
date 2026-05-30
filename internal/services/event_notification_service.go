package services

import (
	"fmt"
	"strings"

	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventNotificationService struct {
	userRepo               *repositories.UserRepository
	notificationDispatcher NotificationDispatcher
	appNotificationService *AppNotificationService
}

func NewEventNotificationService(
	userRepo *repositories.UserRepository,
	notificationDispatcher NotificationDispatcher,
	appNotificationService *AppNotificationService,
) *EventNotificationService {
	return &EventNotificationService{
		userRepo:                userRepo,
		notificationDispatcher: notificationDispatcher,
		appNotificationService: appNotificationService,
	}
}

func (s *EventNotificationService) NotifyAdminsForSOS(
	sosID primitive.ObjectID,
	userName string,
	emergencyType string,
	locationText string,
) {
	title := "New SOS Alert"
	body := buildEventBody(userName, emergencyType, locationText)

	data := map[string]string{
		"type":          "sos",
		"referenceId":   sosID.Hex(),
		"sosId":         sosID.Hex(),
		"emergencyType": emergencyType,
	}

	s.notifyAdmins(
		title,
		body,
		"sos",
		sosID.Hex(),
		data,
	)
}

func (s *EventNotificationService) NotifyAdminsForAssistance(
	assistanceID primitive.ObjectID,
	userName string,
	assistanceType string,
	locationText string,
) {
	title := "New Assistance Request"
	body := buildEventBody(userName, assistanceType, locationText)

	data := map[string]string{
		"type":           "assistance",
		"referenceId":    assistanceID.Hex(),
		"assistanceId":   assistanceID.Hex(),
		"assistanceType": assistanceType,
	}

	s.notifyAdmins(
		title,
		body,
		"assistance",
		assistanceID.Hex(),
		data,
	)
}

func (s *EventNotificationService) NotifyAdminsForReport(
	reportID primitive.ObjectID,
	userName string,
	reportType string,
	locationText string,
) {
	title := "New Incident Report"
	body := buildEventBody(userName, reportType, locationText)

	data := map[string]string{
		"type":        "report",
		"referenceId": reportID.Hex(),
		"reportId":    reportID.Hex(),
		"reportType":  reportType,
	}

	s.notifyAdmins(
		title,
		body,
		"report",
		reportID.Hex(),
		data,
	)
}

func (s *EventNotificationService) NotifyUserForChatResponse(
	userID primitive.ObjectID,
	conversationID primitive.ObjectID,
	senderName string,
	messagePreview string,
) {
	if userID.IsZero() || conversationID.IsZero() {
		return
	}

	title := "New Chat Response"

	body := strings.TrimSpace(messagePreview)
	if body == "" {
		body = "You have a new message from emergency response."
	}

	if strings.TrimSpace(senderName) != "" {
		body = senderName + ": " + body
	}

	data := map[string]string{
		"type":           "chat",
		"referenceId":    conversationID.Hex(),
		"conversationId": conversationID.Hex(),
	}

	s.notifyUser(
		userID,
		title,
		body,
		"chat",
		conversationID.Hex(),
		data,
	)
}

func (s *EventNotificationService) NotifyUserForApprovedAlert(
	userID primitive.ObjectID,
	alertID primitive.ObjectID,
	title string,
	body string,
	category string,
	severity string,
) {
	if userID.IsZero() || alertID.IsZero() {
		return
	}

	if strings.TrimSpace(title) == "" {
		title = "New Disaster Alert"
	}

	if strings.TrimSpace(body) == "" {
		body = "A verified alert has been issued for your area."
	}

	data := map[string]string{
		"type":        "alert",
		"referenceId": alertID.Hex(),
		"alertId":     alertID.Hex(),
		"category":    category,
		"severity":    severity,
	}

	s.notifyUser(
		userID,
		title,
		body,
		"alert",
		alertID.Hex(),
		data,
	)
}

func (s *EventNotificationService) notifyAdmins(
	title string,
	body string,
	notificationType string,
	referenceID string,
	data map[string]string,
) {
	if s.userRepo == nil {
		return
	}

	admins, err := s.userRepo.FindAdmins()
	if err != nil {
		return
	}

	for _, admin := range admins {
		if admin.ID.IsZero() {
			continue
		}

		s.notifyUser(
			admin.ID,
			title,
			body,
			notificationType,
			referenceID,
			data,
		)
	}
}

func (s *EventNotificationService) notifyUser(
	userID primitive.ObjectID,
	title string,
	body string,
	notificationType string,
	referenceID string,
	data map[string]string,
) {
	if userID.IsZero() {
		return
	}

	if s.appNotificationService != nil {
		_, _ = s.appNotificationService.CreateForUser(
			userID,
			title,
			body,
			notificationType,
			referenceID,
			data,
		)
	}

	if s.notificationDispatcher != nil {
		s.notificationDispatcher.Dispatch(jobs.NotificationJob{
			TargetType: jobs.TargetUser,
			UserID:     userID.Hex(),
			Title:      title,
			Body:       body,
			Data:       data,
		})
	}
}

func buildEventBody(
	userName string,
	eventType string,
	locationText string,
) string {
	name := strings.TrimSpace(userName)
	if name == "" {
		name = "A user"
	}

	event := strings.TrimSpace(eventType)
	if event == "" {
		event = "emergency"
	}

	location := strings.TrimSpace(locationText)
	if location == "" {
		return fmt.Sprintf("%s submitted a %s request.", name, event)
	}

	return fmt.Sprintf("%s submitted a %s request at %s.", name, event, location)
}

func (s *EventNotificationService) NotifyUsersForApprovedAlert(
	alert *models.Alert,
) {
	if alert == nil || alert.ID.IsZero() {
		return
	}

	country := strings.TrimSpace(alert.Location.Country)
	if country == "" {
		return
	}

	users, err := s.userRepo.FindUsersByCountry(country)
	if err != nil {
		return
	}

	title := strings.TrimSpace(alert.Title)
	if title == "" {
		title = "New Disaster Alert"
	}

	body := strings.TrimSpace(alert.Description)
	if body == "" {
		body = "A verified alert has been issued in your country."
	}

	data := map[string]string{
		"type":        "alert",
		"referenceId": alert.ID.Hex(),
		"alertId":     alert.ID.Hex(),
		"category":    alert.Category,
		"severity":    alert.Severity,
		"country":     country,
	}

	for _, user := range users {
		if user.ID.IsZero() {
			continue
		}

		s.notifyUser(
			user.ID,
			title,
			body,
			"alert",
			alert.ID.Hex(),
			data,
		)
	}
}