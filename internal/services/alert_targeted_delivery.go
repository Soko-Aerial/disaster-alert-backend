package services

import (
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type alertTargetRecipient struct {
	UserID     primitive.ObjectID
	Zone       string
	DistanceKm float64
}

type alertTargetSelection struct {
	DangerRecipients    []alertTargetRecipient
	AwarenessRecipients []alertTargetRecipient

	ScannedUsers int

	ExcludedNoLocation    int
	ExcludedOldLocation   int
	ExcludedPreferenceOff int
	ExcludedNoFCMToken    int
	ExcludedOutsideTarget int
	SkippedDuplicate      int
}

func (s *AlertService) notifyUsersInAlertTarget(
	alert *models.Alert,
	deliveryType string,
) {
	if alert == nil {
		return
	}

	if s.userRepo == nil {
		return
	}

	selection, err := s.selectAlertTargetRecipients(alert, deliveryType)
	if err != nil {
		return
	}

	totalRecipients :=
		len(selection.DangerRecipients) +
			len(selection.AwarenessRecipients)

	if totalRecipients == 0 {
		return
	}

	batch := s.createAlertDeliveryBatch(alert, deliveryType, selection)

	title := strings.TrimSpace(alert.Title)
	if title == "" {
		title = "Emergency Alert"
	}

	dangerBody := strings.TrimSpace(alert.Description)
	if dangerBody == "" {
		dangerBody = "An emergency alert has been issued near you."
	}

	awarenessBody := buildAwarenessAlertBody(alert)

	baseData := buildAlertNotificationData(alert, deliveryType)

	dangerUserIDs := make([]primitive.ObjectID, 0, len(selection.DangerRecipients))
	awarenessUserIDs := make([]primitive.ObjectID, 0, len(selection.AwarenessRecipients))

	successCount := 0
	failureCount := 0
	skippedCount := selection.SkippedDuplicate

	for _, recipient := range selection.DangerRecipients {
		if recipient.UserID.IsZero() {
			continue
		}

		dangerUserIDs = append(dangerUserIDs, recipient.UserID)

		data := copyStringMap(baseData)
		data["zone"] = models.AlertTargetZoneDanger
		data["distanceKm"] = floatToString(recipient.DistanceKm)

		if s.notificationDispatcher != nil {
			s.notificationDispatcher.Dispatch(jobs.NotificationJob{
				TargetType: jobs.TargetUser,
				UserID:     recipient.UserID.Hex(),
				Title:      title,
				Body:       dangerBody,
				Data:       data,
			})
		}

		successCount++

		s.recordAlertRecipientDelivery(
			alert,
			batch,
			recipient,
			deliveryType,
			models.AlertDeliveryStatusSent,
			"",
		)
	}

	for _, recipient := range selection.AwarenessRecipients {
		if recipient.UserID.IsZero() {
			continue
		}

		awarenessUserIDs = append(awarenessUserIDs, recipient.UserID)

		data := copyStringMap(baseData)
		data["zone"] = models.AlertTargetZoneAwareness
		data["distanceKm"] = floatToString(recipient.DistanceKm)

		if s.notificationDispatcher != nil {
			s.notificationDispatcher.Dispatch(jobs.NotificationJob{
				TargetType: jobs.TargetUser,
				UserID:     recipient.UserID.Hex(),
				Title:      title,
				Body:       awarenessBody,
				Data:       data,
			})
		}

		successCount++

		s.recordAlertRecipientDelivery(
			alert,
			batch,
			recipient,
			deliveryType,
			models.AlertDeliveryStatusSent,
			"",
		)
	}

	if s.appNotificationService != nil {
		if len(dangerUserIDs) > 0 {
			data := copyStringMap(baseData)
			data["zone"] = models.AlertTargetZoneDanger

			if err := s.appNotificationService.CreateManyForUsers(
				dangerUserIDs,
				title,
				dangerBody,
				"alert",
				alert.ID.Hex(),
				data,
			); err != nil {
				failureCount += len(dangerUserIDs)
			}
		}

		if len(awarenessUserIDs) > 0 {
			data := copyStringMap(baseData)
			data["zone"] = models.AlertTargetZoneAwareness

			if err := s.appNotificationService.CreateManyForUsers(
				awarenessUserIDs,
				title,
				awarenessBody,
				"alert",
				alert.ID.Hex(),
				data,
			); err != nil {
				failureCount += len(awarenessUserIDs)
			}
		}
	}

	if s.alertDeliveryRepo != nil && batch != nil {
		_ = s.alertDeliveryRepo.CompleteBatch(
			batch.ID,
			successCount,
			failureCount,
			skippedCount,
		)
	}
}

func (s *AlertService) selectAlertTargetRecipients(
	alert *models.Alert,
	deliveryType string,
) (*alertTargetSelection, error) {
	selection := &alertTargetSelection{
		DangerRecipients:    []alertTargetRecipient{},
		AwarenessRecipients: []alertTargetRecipient{},
	}

	if alert == nil {
		return selection, nil
	}

	targeting := normalizeAlertTargeting(alert)

	candidates, err := s.userRepo.FindAlertTargetCandidates(
		targeting.Country,
		targeting.Region,
		50000,
	)
	if err != nil {
		return nil, err
	}

	selection.ScannedUsers = len(candidates)

	activeFCMMap, hasAuthoritativeFCMMap := s.activeFCMUserMapForCandidates(candidates)

	now := time.Now().UTC()
	category := strings.ToLower(strings.TrimSpace(alert.Category))
	severity := strings.ToLower(strings.TrimSpace(alert.Severity))

	for _, candidate := range candidates {
		if !candidate.LocationSharingEnabled {
			selection.ExcludedNoLocation++
			continue
		}

		if !candidate.HasLocation {
			selection.ExcludedNoLocation++
			continue
		}

		if !candidateHasActiveFCM(candidate, activeFCMMap, hasAuthoritativeFCMMap) {
			selection.ExcludedNoFCMToken++
			continue
		}

		if isLocationOld(candidate, targeting.LocationFreshnessHours, now) {
			selection.ExcludedOldLocation++
			continue
		}

		zone, insideTarget, distanceKm := classifyCandidateZoneForAlert(
			alert,
			targeting,
			candidate,
		)

		if !insideTarget {
			selection.ExcludedOutsideTarget++
			continue
		}

		if targeting.RespectUserPreferences &&
			!targeting.CriticalOverridePreferences &&
			!candidatePreferenceAllowsCategory(candidate, category) {
			selection.ExcludedPreferenceOff++
			continue
		}

		if targeting.RespectUserPreferences &&
			targeting.CriticalOverridePreferences &&
			severity != "critical" &&
			!candidatePreferenceAllowsCategory(candidate, category) {
			selection.ExcludedPreferenceOff++
			continue
		}

		recipient := alertTargetRecipient{
			UserID:     candidate.UserID,
			Zone:       zone,
			DistanceKm: distanceKm,
		}

		if s.hasAlreadyReceivedAlert(alert, recipient, deliveryType) {
			selection.SkippedDuplicate++
			continue
		}

		switch zone {
		case models.AlertTargetZoneDanger:
			selection.DangerRecipients = append(selection.DangerRecipients, recipient)

		case models.AlertTargetZoneAwareness:
			selection.AwarenessRecipients = append(selection.AwarenessRecipients, recipient)

		default:
			selection.ExcludedOutsideTarget++
		}
	}

	return selection, nil
}

func (s *AlertService) createAlertDeliveryBatch(
	alert *models.Alert,
	deliveryType string,
	selection *alertTargetSelection,
) *models.AlertDeliveryBatch {
	if s.alertDeliveryRepo == nil || alert == nil || selection == nil {
		return nil
	}

	now := time.Now().UTC()

	batch := models.AlertDeliveryBatch{
		ID:           primitive.NewObjectID(),
		AlertID:      alert.ID,
		DeliveryType: deliveryType,

		TargetingMode: alert.Targeting.Mode,

		DangerRecipients:    len(selection.DangerRecipients),
		AwarenessRecipients: len(selection.AwarenessRecipients),
		TotalRecipients: len(selection.DangerRecipients) +
			len(selection.AwarenessRecipients),

		SuccessCount: 0,
		FailureCount: 0,
		SkippedCount: selection.SkippedDuplicate,

		ExcludedNoLocation:    selection.ExcludedNoLocation,
		ExcludedOldLocation:   selection.ExcludedOldLocation,
		ExcludedPreferenceOff: selection.ExcludedPreferenceOff,
		ExcludedNoFCMToken:    selection.ExcludedNoFCMToken,

		StartedAt: now,
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdBatch, err := s.alertDeliveryRepo.CreateBatch(batch)
	if err != nil {
		return nil
	}

	return createdBatch
}

func (s *AlertService) recordAlertRecipientDelivery(
	alert *models.Alert,
	batch *models.AlertDeliveryBatch,
	recipient alertTargetRecipient,
	deliveryType string,
	status string,
	failureReason string,
) {
	if s.alertDeliveryRepo == nil {
		return
	}

	if alert == nil || batch == nil || recipient.UserID.IsZero() {
		return
	}

	now := time.Now().UTC()

	sentAt := &now
	if status != models.AlertDeliveryStatusSent {
		sentAt = nil
	}

	delivery := models.AlertRecipientDelivery{
		ID:              primitive.NewObjectID(),
		AlertID:         alert.ID,
		DeliveryBatchID: batch.ID,
		UserID:          recipient.UserID,

		DeliveryType: deliveryType,
		Zone:         recipient.Zone,
		Status:       status,

		DedupKey: buildAlertDedupKey(
			alert.ID,
			recipient.UserID,
			deliveryType,
			recipient.Zone,
		),

		DistanceKm:    recipient.DistanceKm,
		FailureReason: failureReason,

		SentAt:    sentAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_ = s.alertDeliveryRepo.CreateRecipientDelivery(delivery)
}

func (s *AlertService) hasAlreadyReceivedAlert(
	alert *models.Alert,
	recipient alertTargetRecipient,
	deliveryType string,
) bool {
	if s.alertDeliveryRepo == nil {
		return false
	}

	if alert == nil || recipient.UserID.IsZero() {
		return false
	}

	// Area expansion rule:
	// If this is an escalation caused by area expansion,
	// send only to users who have never received this alert before.
	if deliveryType == models.AlertDeliveryTypeEscalation {
		exists, err := s.alertDeliveryRepo.HasUserReceivedAlert(
			alert.ID,
			recipient.UserID,
		)
		if err != nil {
			return false
		}

		return exists
	}

	dedupKey := buildAlertDedupKey(
		alert.ID,
		recipient.UserID,
		deliveryType,
		recipient.Zone,
	)

	exists, err := s.alertDeliveryRepo.HasDedupKey(dedupKey)
	if err != nil {
		return false
	}

	return exists
}

func buildAlertDedupKey(
	alertID primitive.ObjectID,
	userID primitive.ObjectID,
	deliveryType string,
	zone string,
) string {
	cleanDeliveryType := strings.TrimSpace(deliveryType)
	cleanZone := strings.TrimSpace(zone)

	// Critical update rule:
	// A user should receive one critical update per alert,
	// not one per zone change.
	if cleanDeliveryType == models.AlertDeliveryTypeUpdate {
		cleanZone = "any"
	}

	return alertID.Hex() + ":" +
		userID.Hex() + ":" +
		cleanDeliveryType + ":" +
		cleanZone
}

func normalizeAlertTargeting(alert *models.Alert) models.AlertTargeting {
	targeting := alert.Targeting

	mode := strings.ToLower(strings.TrimSpace(targeting.Mode))
	if mode == "" {
		if len(targeting.Polygon) >= 3 {
			mode = models.AlertTargetingModePolygon
		} else if targeting.RadiusKm > 0 || alert.RadiusKm > 0 {
			mode = models.AlertTargetingModeRadius
		} else if strings.TrimSpace(targeting.Region) != "" || strings.TrimSpace(alert.Location.Region) != "" {
			mode = models.AlertTargetingModeRegion
		} else if strings.TrimSpace(targeting.Country) != "" || strings.TrimSpace(alert.Location.Country) != "" {
			mode = models.AlertTargetingModeCountry
		} else {
			mode = models.AlertTargetingModeRadius
		}
	}

	targeting.Mode = mode

	if targeting.RadiusKm <= 0 {
		targeting.RadiusKm = alert.RadiusKm
	}

	if targeting.RadiusKm <= 0 {
		targeting.RadiusKm = 5
	}

	if targeting.AwarenessRadiusKm <= 0 {
		targeting.AwarenessRadiusKm = targeting.RadiusKm * 2
	}

	if targeting.AwarenessRadiusKm < targeting.RadiusKm {
		targeting.AwarenessRadiusKm = targeting.RadiusKm
	}

	if targeting.LocationFreshnessHours <= 0 {
		targeting.LocationFreshnessHours = 24
	}

	if strings.TrimSpace(targeting.Country) == "" {
		targeting.Country = strings.TrimSpace(alert.Location.Country)
	}

	if strings.TrimSpace(targeting.Region) == "" {
		targeting.Region = strings.TrimSpace(alert.Location.Region)
	}

	targeting.RespectUserPreferences = true
	targeting.CriticalOverridePreferences = true

	return targeting
}

func classifyCandidateZoneForAlert(
	alert *models.Alert,
	targeting models.AlertTargeting,
	candidate repositories.AlertTargetCandidate,
) (string, bool, float64) {
	switch targeting.Mode {
	case models.AlertTargetingModeRadius:
		distance := alertDistanceKm(
			alert.Location.Latitude,
			alert.Location.Longitude,
			candidate.Latitude,
			candidate.Longitude,
		)

		if distance <= targeting.RadiusKm {
			return models.AlertTargetZoneDanger, true, distance
		}

		if distance <= targeting.AwarenessRadiusKm {
			return models.AlertTargetZoneAwareness, true, distance
		}

		return "", false, distance

	case models.AlertTargetingModePolygon:
		if pointInsidePolygon(
			candidate.Latitude,
			candidate.Longitude,
			targeting.Polygon,
		) {
			return models.AlertTargetZoneDanger, true, 0
		}

		centroid := polygonCentroid(targeting.Polygon)

		distanceFromCentroid := alertDistanceKm(
			centroid.Latitude,
			centroid.Longitude,
			candidate.Latitude,
			candidate.Longitude,
		)

		if targeting.AwarenessRadiusKm > 0 &&
			distanceFromCentroid <= targeting.AwarenessRadiusKm {
			return models.AlertTargetZoneAwareness, true, distanceFromCentroid
		}

		return "", false, distanceFromCentroid

	case models.AlertTargetingModeRegion:
		if sameText(candidate.Region, targeting.Region) {
			return models.AlertTargetZoneDanger, true, 0
		}

		return "", false, 0

	case models.AlertTargetingModeCountry:
		if sameText(candidate.Country, targeting.Country) {
			return models.AlertTargetZoneDanger, true, 0
		}

		return "", false, 0

	case models.AlertTargetingModeNational:
		return models.AlertTargetZoneDanger, true, 0

	default:
		return "", false, 0
	}
}

func buildAlertNotificationData(
	alert *models.Alert,
	deliveryType string,
) map[string]string {
	return map[string]string{
		"type":          "alert",
		"referenceId":   alert.ID.Hex(),
		"alertId":       alert.ID.Hex(),
		"deliveryType":  deliveryType,
		"category":      alert.Category,
		"severity":      alert.Severity,
		"source":        alert.SourceName,
		"country":       alert.Location.Country,
		"region":        alert.Location.Region,
		"targetingMode": alert.Targeting.Mode,
		"latitude":      floatToString(alert.Location.Latitude),
		"longitude":     floatToString(alert.Location.Longitude),
	}
}

func buildAwarenessAlertBody(alert *models.Alert) string {
	location := strings.TrimSpace(alert.Location.Address)
	if location == "" {
		location = strings.TrimSpace(alert.Location.Region)
	}

	if location == "" {
		location = "the affected area"
	}

	category := strings.TrimSpace(alert.Category)
	if category == "" {
		category = "emergency"
	}

	return "Emergency nearby: " + category + " reported around " + location + ". Avoid the affected area and follow official instructions."
}

func copyStringMap(input map[string]string) map[string]string {
	output := map[string]string{}

	for key, value := range input {
		output[key] = value
	}

	return output
}

func (s *AlertService) activeFCMUserMapForCandidates(
	candidates []repositories.AlertTargetCandidate,
) (map[primitive.ObjectID]bool, bool) {
	result := map[primitive.ObjectID]bool{}

	if s.fcmTokenRepo == nil {
		return result, false
	}

	userIDs := make([]primitive.ObjectID, 0, len(candidates))
	seen := map[primitive.ObjectID]bool{}

	for _, candidate := range candidates {
		if candidate.UserID.IsZero() {
			continue
		}

		if seen[candidate.UserID] {
			continue
		}

		seen[candidate.UserID] = true
		userIDs = append(userIDs, candidate.UserID)
	}

	activeMap, err := s.fcmTokenRepo.FindActiveUserIDMap(userIDs)
	if err != nil {
		return result, false
	}

	return activeMap, true
}

func candidateHasActiveFCM(
	candidate repositories.AlertTargetCandidate,
	activeFCMMap map[primitive.ObjectID]bool,
	hasAuthoritativeFCMMap bool,
) bool {
	if candidate.UserID.IsZero() {
		return false
	}

	if hasAuthoritativeFCMMap {
		return activeFCMMap[candidate.UserID]
	}

	return candidate.HasFCMToken
}

func buildPreviewRequestFromAlert(alert *models.Alert) dto.AlertTargetingPreviewRequest {
	if alert == nil {
		return dto.AlertTargetingPreviewRequest{}
	}

	targeting := alert.Targeting

	polygon := make([]dto.AlertGeoPointRequest, 0, len(targeting.Polygon))
	for _, point := range targeting.Polygon {
		polygon = append(polygon, dto.AlertGeoPointRequest{
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
		})
	}

	return dto.AlertTargetingPreviewRequest{
		Category:  alert.Category,
		Severity:  alert.Severity,
		Latitude:  alert.Location.Latitude,
		Longitude: alert.Location.Longitude,
		Targeting: dto.AlertTargetingRequest{
			Mode:                        targeting.Mode,
			RadiusKm:                    targeting.RadiusKm,
			AwarenessRadiusKm:           targeting.AwarenessRadiusKm,
			Polygon:                     polygon,
			Country:                     targeting.Country,
			Region:                      targeting.Region,
			District:                    targeting.District,
			LocationFreshnessHours:      targeting.LocationFreshnessHours,
			RespectUserPreferences:      targeting.RespectUserPreferences,
			CriticalOverridePreferences: targeting.CriticalOverridePreferences,
			ConfirmNationalAlert:        targeting.ConfirmNationalAlert,
			NationalAlertReason:         targeting.NationalAlertReason,
		},
	}
}
