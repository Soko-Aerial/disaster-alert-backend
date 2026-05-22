package sources

import (
	"time"

	"disaster_alert_backend/internal/models"
)

type MockAlertSource struct{}

func NewMockAlertSource() *MockAlertSource {
	return &MockAlertSource{}
}

func (s *MockAlertSource) Name() string {
	return "mock_external"
}

func (s *MockAlertSource) FetchAlerts() ([]models.Alert, error) {
	now := time.Now()
	expiresAt := now.Add(6 * time.Hour)
	eventTime := now

	alerts := []models.Alert{
		{
			Title:       "External Flood Watch",
			Description: "Mock external source reports possible flooding in low-lying areas.",
			Category:    "flood",
			Severity:    "high",
			Status:      "active",
			Location: models.AlertLocation{
				Latitude:  5.6037,
				Longitude: -0.1870,
				Address:   "Accra, Ghana",
				Country:   "Ghana",
				Region:    "Greater Accra",
			},
			RadiusKm: 20,
			SafetyInstructions: []string{
				"Move to higher ground",
				"Avoid flood waters",
				"Follow local emergency instructions",
			},
			SourceType: "external",
			SourceName: "mock_external",
			ExternalID: "mock-flood-accra-001",
			SourceURL:  "https://example.com/mock-flood-accra-001",
			EventTime:  &eventTime,
			ExpiresAt:  &expiresAt,
			Confidence: 0.85,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			Title:       "External Fire Hotspot",
			Description: "Mock external satellite feed reports a possible fire hotspot.",
			Category:    "fire",
			Severity:    "medium",
			Status:      "active",
			Location: models.AlertLocation{
				Latitude:  5.6200,
				Longitude: -0.1700,
				Address:   "East Legon, Accra",
				Country:   "Ghana",
				Region:    "Greater Accra",
			},
			RadiusKm: 10,
			SafetyInstructions: []string{
				"Avoid affected area",
				"Stay away from smoke",
				"Report visible fire to emergency services",
			},
			SourceType: "external",
			SourceName: "mock_external",
			ExternalID: "mock-fire-east-legon-001",
			SourceURL:  "https://example.com/mock-fire-east-legon-001",
			EventTime:  &eventTime,
			ExpiresAt:  &expiresAt,
			Confidence: 0.70,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
	}

	return alerts, nil
}