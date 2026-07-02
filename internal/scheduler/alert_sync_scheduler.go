package scheduler

import (
	"context"
	"log"
	"sync"
	"time"

	"disaster_alert_backend/internal/aggregator"
	"disaster_alert_backend/internal/observability"
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
		initialDelay:    2 * time.Minute,
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

	observability.Info(ctx, "Alert sync scheduler started", observability.Fields{
		"module":        "alert_sync_scheduler",
		"interval":      s.interval.String(),
		"initial_delay": s.initialDelay.String(),
	})

	if s.initialDelay > 0 {
		select {
		case <-ctx.Done():
			log.Println("Alert sync scheduler stopped before initial sync")

			observability.Warn(ctx, "Alert sync scheduler stopped before initial sync", observability.Fields{
				"module": "alert_sync_scheduler",
				"reason": "context_cancelled",
			})

			return

		case <-time.After(s.initialDelay):
			s.syncOnce(ctx)
		}
	} else {
		s.syncOnce(ctx)
	}

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Alert sync scheduler stopped")

			observability.Info(ctx, "Alert sync scheduler stopped", observability.Fields{
				"module": "alert_sync_scheduler",
				"reason": "context_cancelled",
			})

			return

		case <-ticker.C:
			s.syncOnce(ctx)
		}
	}
}

func (s *AlertSyncScheduler) syncOnce(ctx context.Context) {
	if !s.tryStartSync() {
		log.Println("External alert sync skipped: previous sync still running")

		observability.Warn(ctx, "External alert sync skipped because previous sync is still running", observability.Fields{
			"module": "alert_sync_scheduler",
			"reason": "previous_sync_still_running",
		})

		return
	}

	defer s.finishSync()

	startedAt := time.Now()

	log.Println("Starting external alert sync...")

	observability.Info(ctx, "External alert sync started", observability.Fields{
		"module": "alert_sync_scheduler",
	})

	err := s.alertAggregator.SyncExternalAlerts()
	if err != nil {
		duration := time.Since(startedAt).Round(time.Millisecond)

		log.Println("External alert sync failed:", err)

		observability.Error(ctx, "External alert sync failed", err, observability.Fields{
			"module":   "alert_sync_scheduler",
			"duration": duration.String(),
		})

		return
	}

	duration := time.Since(startedAt).Round(time.Millisecond)

	log.Printf(
		"External alert sync completed. Duration: %s\n",
		duration,
	)

	observability.Info(ctx, "External alert sync completed", observability.Fields{
		"module":   "alert_sync_scheduler",
		"duration": duration.String(),
	})
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
