package aggregator

import (
	"log"

	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/sources"
)

type AlertAggregator struct {
	alertRepo *repositories.AlertRepository
	sources   []sources.AlertSource
}

func NewAlertAggregator(
	alertRepo *repositories.AlertRepository,
	sources []sources.AlertSource,
) *AlertAggregator {
	return &AlertAggregator{
		alertRepo: alertRepo,
		sources:   sources,
	}
}

func (a *AlertAggregator) SyncExternalAlerts() error {
	deactivatedCount, err := a.alertRepo.DeactivateExpiredExternalAlerts()
	if err != nil {
		log.Println("Failed to cleanup expired external alerts:", err)
	} else {
		log.Printf("Expired external alerts cleaned up: %d\n", deactivatedCount)
	}

	for _, source := range a.sources {
		log.Println("Fetching alerts from source:", source.Name())

		alerts, err := source.FetchAlerts()
		if err != nil {
			log.Printf("Failed to fetch alerts from %s: %v\n", source.Name(), err)
			continue
		}

		for _, alert := range alerts {
			savedAlert, err := a.alertRepo.UpsertExternalAlert(alert)
			if err != nil {
				log.Printf("Failed to sync external alert from %s: %v\n", source.Name(), err)
				continue
			}

			log.Printf("External alert synced: %s from %s\n", savedAlert.ExternalID, source.Name())
		}
	}

	return nil
}