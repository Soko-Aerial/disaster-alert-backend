package scheduler

import (
	"context"
	"log"
	"time"

	"disaster_alert_backend/internal/aggregator"
)

type AlertSyncScheduler struct {
	alertAggregator *aggregator.AlertAggregator
	interval        time.Duration
}

func NewAlertSyncScheduler(
	alertAggregator *aggregator.AlertAggregator,
	interval time.Duration,
) *AlertSyncScheduler {
	return &AlertSyncScheduler{
		alertAggregator: alertAggregator,
		interval:        interval,
	}
}

func (s *AlertSyncScheduler) Start(ctx context.Context) {
	log.Printf("Alert sync scheduler started. Interval: %s\n", s.interval)

	// Run once immediately when backend starts
	s.syncOnce()

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Alert sync scheduler stopped")
			return

		case <-ticker.C:
			s.syncOnce()
		}
	}
}

func (s *AlertSyncScheduler) syncOnce() {
	log.Println("Starting external alert sync...")

	err := s.alertAggregator.SyncExternalAlerts()
	if err != nil {
		log.Println("External alert sync failed:", err)
		return
	}

	log.Println("External alert sync completed")
}