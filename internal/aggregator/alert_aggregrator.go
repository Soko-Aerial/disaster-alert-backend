package aggregator

import (
	"context"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/observability"
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
		maxPerSource: 80,
		sourceDelay:  6 * time.Second,
		writeDelay:   100 * time.Millisecond,
		maxAlertAge:  7 * 24 * time.Hour,
	}
}

func (a *AlertAggregator) SyncExternalAlerts() error {
	ctx := context.Background()
	startedAt := time.Now()

	deactivatedCount, err := a.alertRepo.DeactivateExpiredExternalAlerts()
	if err != nil {
		log.Println("Failed to cleanup expired external alerts:", err)

		observability.Error(ctx, "Failed to cleanup expired external alerts", err, observability.Fields{
			"module": "external_alert_sync",
			"stage":  "cleanup_expired_alerts",
		})
	} else if deactivatedCount > 0 {
		log.Printf("Expired external alerts cleaned up: %d\n", deactivatedCount)

		observability.Info(ctx, "Expired external alerts cleaned up", observability.Fields{
			"module":            "external_alert_sync",
			"stage":             "cleanup_expired_alerts",
			"deactivated_count": strconv.FormatInt(deactivatedCount, 10),
		})
	}

	totalFetched := 0
	totalProcessed := 0
	totalSynced := 0
	totalSkippedOld := 0
	totalFailed := 0

	for index, source := range a.sources {
		sourceName := source.Name()

		if index > 0 && a.sourceDelay > 0 {
			log.Printf("Waiting %s before fetching next external source...\n", a.sourceDelay)
			time.Sleep(a.sourceDelay)
		}

		sourceStartedAt := time.Now()

		log.Println("Fetching alerts from source:", sourceName)

		observability.Info(ctx, "Fetching alerts from external source", observability.Fields{
			"module": "external_alert_sync",
			"stage":  "source_fetch_started",
			"source": sourceName,
		})

		alerts, err := source.FetchAlerts()
		if err != nil {
			totalFailed++

			if isRateLimitError(err) {
				log.Printf(
					"Rate limited by %s. Skipping this source until next scheduled sync: %v\n",
					sourceName,
					err,
				)

				observability.Warn(ctx, "External alert source rate limited", observability.Fields{
					"module": "external_alert_sync",
					"stage":  "source_fetch_rate_limited",
					"source": sourceName,
					"error":  err.Error(),
				})

				continue
			}

			log.Printf("Failed to fetch alerts from %s: %v\n", sourceName, err)

			observability.Error(ctx, "Failed to fetch alerts from external source", err, observability.Fields{
				"module": "external_alert_sync",
				"stage":  "source_fetch_failed",
				"source": sourceName,
			})

			continue
		}

		fetchedCount := len(alerts)
		totalFetched += fetchedCount

		alerts = filterRecentExternalAlerts(alerts, a.maxAlertAge)

		skippedOldCount := fetchedCount - len(alerts)
		totalSkippedOld += skippedOldCount

		alerts = prepareExternalAlertsForSync(alerts)

		if a.maxPerSource > 0 && len(alerts) > a.maxPerSource {
			alerts = alerts[:a.maxPerSource]
		}

		processedCount := len(alerts)
		totalProcessed += processedCount

		sourceSynced := 0
		sourceFailed := 0
		var lastSourceSyncErr error

		for _, alert := range alerts {
			_, err := a.alertRepo.UpsertExternalAlert(alert)
			if err != nil {
				sourceFailed++
				totalFailed++
				lastSourceSyncErr = err

				log.Printf("Failed to sync external alert from %s: %v\n", sourceName, err)
				continue
			}

			sourceSynced++
			totalSynced++

			if a.writeDelay > 0 {
				time.Sleep(a.writeDelay)
			}
		}

		sourceDuration := time.Since(sourceStartedAt).Round(time.Millisecond)

		log.Printf(
			"External source sync completed: source=%s fetched=%d skippedOld=%d processed=%d synced=%d failed=%d duration=%s\n",
			sourceName,
			fetchedCount,
			skippedOldCount,
			processedCount,
			sourceSynced,
			sourceFailed,
			sourceDuration,
		)

		fields := observability.Fields{
			"module":      "external_alert_sync",
			"stage":       "source_sync_completed",
			"source":      sourceName,
			"fetched":     strconv.Itoa(fetchedCount),
			"skipped_old": strconv.Itoa(skippedOldCount),
			"processed":   strconv.Itoa(processedCount),
			"synced":      strconv.Itoa(sourceSynced),
			"failed":      strconv.Itoa(sourceFailed),
			"duration":    sourceDuration.String(),
		}

		if sourceFailed > 0 {
			observability.Error(ctx, "External source sync completed with failures", lastSourceSyncErr, fields)
		} else {
			observability.Info(ctx, "External source sync completed", fields)
		}
	}

	totalDuration := time.Since(startedAt).Round(time.Millisecond)

	log.Printf(
		"External alert sync completed: fetched=%d skippedOld=%d processed=%d synced=%d failed=%d duration=%s\n",
		totalFetched,
		totalSkippedOld,
		totalProcessed,
		totalSynced,
		totalFailed,
		totalDuration,
	)

	fields := observability.Fields{
		"module":      "external_alert_sync",
		"stage":       "sync_completed",
		"fetched":     strconv.Itoa(totalFetched),
		"skipped_old": strconv.Itoa(totalSkippedOld),
		"processed":   strconv.Itoa(totalProcessed),
		"synced":      strconv.Itoa(totalSynced),
		"failed":      strconv.Itoa(totalFailed),
		"duration":    totalDuration.String(),
	}

	if totalFailed > 0 {
		observability.Warn(ctx, "External alert sync completed with failures", fields)
	} else {
		observability.Info(ctx, "External alert sync completed", fields)
	}

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

func prepareExternalAlertsForSync(alerts []models.Alert) []models.Alert {
	now := time.Now().UTC()

	prepared := make([]models.Alert, 0, len(alerts))

	for _, alert := range alerts {
		alert = enrichAlertPriority(alert, now)

		// Drop weak/low-value external alerts unless they are very recent or verified.
		if shouldDropLowValueAlert(alert, now) {
			continue
		}

		prepared = append(prepared, alert)
	}

	sort.SliceStable(prepared, func(i, j int) bool {
		if prepared[i].PriorityScore != prepared[j].PriorityScore {
			return prepared[i].PriorityScore > prepared[j].PriorityScore
		}

		if prepared[i].Confidence != prepared[j].Confidence {
			return prepared[i].Confidence > prepared[j].Confidence
		}

		return alertTime(prepared[i]).After(alertTime(prepared[j]))
	})

	return prepared
}

func enrichAlertPriority(alert models.Alert, now time.Time) models.Alert {
	alert.Title = strings.TrimSpace(alert.Title)
	alert.Description = strings.TrimSpace(alert.Description)
	alert.Summary = buildAlertSummary(alert)

	if alert.LastSyncedAt == nil {
		alert.LastSyncedAt = &now
	}

	alert.Tags = buildAlertTags(alert)
	alert.IsVerified = isVerifiedAlertSource(alert.SourceName)

	score := 0

	score += severityScore(alert.Severity)
	score += categoryScore(alert.Category)
	score += sourceScore(alert.SourceName)
	score += confidenceScore(alert.Confidence)
	score += recencyScore(alert.EventTime, now)
	score += mediaScore(alert)
	score += keywordScore(alert.Title + " " + alert.Description)

	if alert.Location.Country != "" {
		score += 5
	}

	if alert.EventTime == nil {
		score -= 10
	}

	if strings.EqualFold(alert.Severity, "low") {
		score -= 25
	}

	if score < 0 {
		score = 0
	}

	alert.PriorityScore = score
	alert.PriorityLabel = priorityLabel(score)
	alert.IsBreaking = score >= 85

	return alert
}

func shouldDropLowValueAlert(alert models.Alert, now time.Time) bool {
	if !strings.EqualFold(alert.Severity, "low") {
		return false
	}

	// Keep low alerts if they are recent and from a strong official source.
	if alert.IsVerified && recencyScore(alert.EventTime, now) >= 20 {
		return false
	}

	// Keep low alerts if they still scored high because of strong keywords.
	if alert.PriorityScore >= 55 {
		return false
	}

	return true
}

func severityScore(severity string) int {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		return 55
	case "high":
		return 40
	case "medium":
		return 22
	case "low":
		return 5
	default:
		return 12
	}
}

func categoryScore(category string) int {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "earthquake", "flood", "fire", "weather", "health", "conflict":
		return 15
	case "volcano", "drought":
		return 12
	default:
		return 8
	}
}

func sourceScore(sourceName string) int {
	switch strings.ToLower(strings.TrimSpace(sourceName)) {
	case "gdacs":
		return 20
	case "nasa_firms":
		return 16
	case "reliefweb":
		return 14
	case "gdelt":
		return 8
	default:
		return 5
	}
}

func confidenceScore(confidence float64) int {
	switch {
	case confidence >= 0.90:
		return 15
	case confidence >= 0.75:
		return 12
	case confidence >= 0.60:
		return 8
	case confidence >= 0.40:
		return 4
	default:
		return 0
	}
}

func recencyScore(eventTime *time.Time, now time.Time) int {
	if eventTime == nil {
		return 0
	}

	age := now.Sub(*eventTime)

	switch {
	case age <= 3*time.Hour:
		return 25
	case age <= 12*time.Hour:
		return 20
	case age <= 24*time.Hour:
		return 15
	case age <= 72*time.Hour:
		return 10
	case age <= 7*24*time.Hour:
		return 4
	default:
		return 0
	}
}

func mediaScore(alert models.Alert) int {
	score := 0

	if len(alert.ImageURLs) > 0 {
		score += 6
	}

	if len(alert.VideoURLs) > 0 {
		score += 8
	}

	return score
}

func keywordScore(text string) int {
	text = strings.ToLower(text)

	score := 0

	criticalKeywords := []string{
		"evacuation",
		"state of emergency",
		"deadly",
		"killed",
		"death",
		"deaths",
		"missing",
		"collapsed",
		"outbreak",
		"cholera",
		"war",
		"attack",
		"explosion",
		"landslide",
		"severe",
		"major",
		"catastrophic",
	}

	for _, keyword := range criticalKeywords {
		if strings.Contains(text, keyword) {
			score += 5
		}
	}

	if score > 25 {
		score = 25
	}

	return score
}

func priorityLabel(score int) string {
	switch {
	case score >= 85:
		return "breaking"
	case score >= 65:
		return "serious"
	case score >= 45:
		return "watch"
	default:
		return "low"
	}
}

func isVerifiedAlertSource(sourceName string) bool {
	switch strings.ToLower(strings.TrimSpace(sourceName)) {
	case "gdacs", "nasa_firms", "reliefweb":
		return true
	default:
		return false
	}
}

func alertTime(alert models.Alert) time.Time {
	if alert.EventTime != nil {
		return *alert.EventTime
	}

	if !alert.CreatedAt.IsZero() {
		return alert.CreatedAt
	}

	return time.Time{}
}

func buildAlertSummary(alert models.Alert) string {
	if strings.TrimSpace(alert.Summary) != "" {
		return strings.TrimSpace(alert.Summary)
	}

	description := strings.TrimSpace(alert.Description)
	if description == "" {
		return strings.TrimSpace(alert.Title)
	}

	if len(description) <= 220 {
		return description
	}

	return description[:220] + "..."
}

func buildAlertTags(alert models.Alert) []string {
	seen := map[string]bool{}
	tags := []string{}

	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			return
		}

		seen[value] = true
		tags = append(tags, value)
	}

	add(alert.Category)
	add(alert.Severity)
	add(alert.SourceName)
	add(alert.Location.Country)
	add(alert.PriorityLabel)

	return tags
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
