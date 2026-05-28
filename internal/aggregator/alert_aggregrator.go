package aggregator

import (
	"log"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/sources"
)

type AlertAggregator struct {
	alertRepo    *repositories.AlertRepository
	sources      []sources.AlertSource
	maxPerSource int
	sourceDelay  time.Duration
	writeDelay   time.Duration
	maxAlertAge  time.Duration
}

func NewAlertAggregator(
	alertRepo *repositories.AlertRepository,
	sources []sources.AlertSource,
) *AlertAggregator {
	return &AlertAggregator{
		alertRepo:    alertRepo,
		sources:      sources,
		maxPerSource: 50,
		sourceDelay:  6 * time.Second,
		writeDelay:   100 * time.Millisecond,
		maxAlertAge:  7 * 24 * time.Hour,
	}
}

func (a *AlertAggregator) SyncExternalAlerts() error {
	startedAt := time.Now()

	deactivatedCount, err := a.alertRepo.DeactivateExpiredExternalAlerts()
	if err != nil {
		log.Println("Failed to cleanup expired external alerts:", err)
	} else if deactivatedCount > 0 {
		log.Printf("Expired external alerts cleaned up: %d\n", deactivatedCount)
	}

	totalFetched := 0
	totalProcessed := 0
	totalSynced := 0
	totalSkippedOld := 0
	totalFailed := 0

	for index, source := range a.sources {
		if index > 0 && a.sourceDelay > 0 {
			log.Printf("Waiting %s before fetching next external source...\n", a.sourceDelay)
			time.Sleep(a.sourceDelay)
		}

		sourceStartedAt := time.Now()

		log.Println("Fetching alerts from source:", source.Name())

		alerts, err := source.FetchAlerts()
		if err != nil {
			totalFailed++

			if isRateLimitError(err) {
				log.Printf(
					"Rate limited by %s. Skipping this source until next scheduled sync: %v\n",
					source.Name(),
					err,
				)
				continue
			}

			log.Printf("Failed to fetch alerts from %s: %v\n", source.Name(), err)
			continue
		}

		fetchedCount := len(alerts)
		totalFetched += fetchedCount

		alerts = filterRecentExternalAlerts(alerts, a.maxAlertAge)

		skippedOldCount := fetchedCount - len(alerts)
		totalSkippedOld += skippedOldCount

		if a.maxPerSource > 0 && len(alerts) > a.maxPerSource {
			alerts = alerts[:a.maxPerSource]
		}

		processedCount := len(alerts)
		totalProcessed += processedCount

		sourceSynced := 0
		sourceFailed := 0

		for _, alert := range alerts {
			_, err := a.alertRepo.UpsertExternalAlert(alert)
			if err != nil {
				sourceFailed++
				totalFailed++
				log.Printf("Failed to sync external alert from %s: %v\n", source.Name(), err)
				continue
			}

			sourceSynced++
			totalSynced++

			if a.writeDelay > 0 {
				time.Sleep(a.writeDelay)
			}
		}

		log.Printf(
			"External source sync completed: source=%s fetched=%d skippedOld=%d processed=%d synced=%d failed=%d duration=%s\n",
			source.Name(),
			fetchedCount,
			skippedOldCount,
			processedCount,
			sourceSynced,
			sourceFailed,
			time.Since(sourceStartedAt).Round(time.Millisecond),
		)
	}

	log.Printf(
		"External alert sync completed: fetched=%d skippedOld=%d processed=%d synced=%d failed=%d duration=%s\n",
		totalFetched,
		totalSkippedOld,
		totalProcessed,
		totalSynced,
		totalFailed,
		time.Since(startedAt).Round(time.Millisecond),
	)

	return nil
}

func filterRecentExternalAlerts(
	alerts []models.Alert,
	maxAge time.Duration,
) []models.Alert {
	if maxAge <= 0 {
		return alerts
	}

	now := time.Now().UTC()
	cutoff := now.Add(-maxAge)

	filteredAlerts := make([]models.Alert, 0, len(alerts))

	for _, alert := range alerts {
		if alert.EventTime == nil {
			filteredAlerts = append(filteredAlerts, alert)
			continue
		}

		if alert.EventTime.Before(cutoff) {
			continue
		}

		filteredAlerts = append(filteredAlerts, alert)
	}

	return filteredAlerts
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	lowerMessage := strings.ToLower(message)

	return strings.Contains(message, "429") ||
		strings.Contains(lowerMessage, "rate limit") ||
		strings.Contains(lowerMessage, "limit requests")
}