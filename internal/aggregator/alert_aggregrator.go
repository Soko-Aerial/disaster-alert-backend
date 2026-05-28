package aggregator

import (
	"log"
	"strings"
	"time"

	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/sources"
)

type AlertAggregator struct {
	alertRepo *repositories.AlertRepository
	sources   []sources.AlertSource
	maxPerSource int
	sourceDelay   time.Duration
}

func NewAlertAggregator(
	alertRepo *repositories.AlertRepository,
	sources []sources.AlertSource,
) *AlertAggregator {
	return &AlertAggregator{
		alertRepo:     alertRepo,
		sources:       sources,
		maxPerSource:  50,
		sourceDelay:  6 * time.Second,
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

			time.Sleep(100 * time.Millisecond)
		}

		log.Printf(
			"External source sync completed: source=%s fetched=%d processed=%d synced=%d failed=%d duration=%s\n",
			source.Name(),
			fetchedCount,
			processedCount,
			sourceSynced,
			sourceFailed,
			time.Since(sourceStartedAt).Round(time.Millisecond),
		)
	}

	log.Printf(
		"External alert sync completed: fetched=%d processed=%d synced=%d failed=%d duration=%s\n",
		totalFetched,
		totalProcessed,
		totalSynced,
		totalFailed,
		time.Since(startedAt).Round(time.Millisecond),
	)

	return nil
}

func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()

	return strings.Contains(message, "429") ||
		strings.Contains(strings.ToLower(message), "rate limit") ||
		strings.Contains(strings.ToLower(message), "limit requests")
}