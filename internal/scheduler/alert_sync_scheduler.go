package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"disaster_alert_backend/internal/aggregator"
)

type AlertSyncScheduler struct {
	alertAggregator *aggregator.AlertAggregator
	interval        time.Duration
	initialDelay    time.Duration

	mu        sync.Mutex
	isRunning bool
}

func NewAlertSyncScheduler(
	alertAggregator *aggregator.AlertAggregator,
	interval time.Duration,
) *AlertSyncScheduler {
	return &AlertSyncScheduler{
		alertAggregator: alertAggregator,
		interval:        interval,

		initialDelay: 2 * time.Minute,
	}
}

func NewAlertSyncSchedulerWithDelay(
	alertAggregator *aggregator.AlertAggregator,
	interval time.Duration,
	initialDelay time.Duration,
) *AlertSyncScheduler {
	return &AlertSyncScheduler{
		alertAggregator: alertAggregator,
		interval:        interval,
		initialDelay:    initialDelay,
	}
}

func (s *AlertSyncScheduler) Start(ctx context.Context) {
	log.Printf(
		"Alert sync scheduler started. Interval: %s InitialDelay: %s\n",
		s.interval,
		s.initialDelay,
	)

	if s.initialDelay > 0 {
		select {
		case <-ctx.Done():
			log.Println("Alert sync scheduler stopped before initial sync")
			return

		case <-time.After(s.initialDelay):
			s.syncOnce()
		}
	} else {
		s.syncOnce()
	}

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
	if !s.tryStartSync() {
		log.Println("External alert sync skipped: previous sync still running")
		return
	}

	defer s.finishSync()

	startedAt := time.Now()

	log.Println("Starting external alert sync...")

	err := s.alertAggregator.SyncExternalAlerts()
	if err != nil {
		log.Println("External alert sync failed:", err)
		return
	}

	log.Printf(
		"External alert sync completed. Duration: %s\n",
		time.Since(startedAt).Round(time.Millisecond),
	)
}

func (s *AlertSyncScheduler) tryStartSync() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return false
	}

	s.isRunning = true
	return true
}

func (s *AlertSyncScheduler) finishSync() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.isRunning = false
}