package services

import (
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"
)

func (s *AlertService) PreviewAlertTargeting(
	req dto.AlertTargetingPreviewRequest,
) (*dto.AlertTargetingPreviewResponse, error) {
	if s.userRepo == nil {
		return nil, errors.New("user repository is not available")
	}

	category := strings.ToLower(strings.TrimSpace(req.Category))
	severity := strings.ToLower(strings.TrimSpace(req.Severity))

	targeting := buildPreviewTargeting(req)

	if err := validatePreviewTargeting(req, targeting); err != nil {
		return nil, err
	}

	candidates, err := s.userRepo.FindAlertTargetCandidates(
		targeting.Country,
		targeting.Region,
		10000,
	)
	if err != nil {
		return nil, err
	}

	response := &dto.AlertTargetingPreviewResponse{
		TargetingMode:          targeting.Mode,
		LocationFreshnessHours: targeting.LocationFreshnessHours,
		RespectUserPreferences: targeting.RespectUserPreferences,
		Warnings:               []string{},
		DangerRecipients:       0,
		AwarenessRecipients:    0,
		TotalPushRecipients:    0,
		ExcludedNoLocation:     0,
		ExcludedOldLocation:    0,
		ExcludedPreferenceOff:  0,
		ExcludedNoFCMToken:     0,
		ExcludedOutsideTarget:  0,
		ScannedUsers:           len(candidates),
	}

	if len(candidates) == 0 {
		response.Warnings = append(
			response.Warnings,
			"No users matched the broad country/region filter.",
		)

		return response, nil
	}

	activeFCMMap, hasAuthoritativeFCMMap := s.activeFCMUserMapForCandidates(candidates)

	now := time.Now().UTC()

	for _, candidate := range candidates {
		if !candidate.LocationSharingEnabled {
			response.ExcludedNoLocation++
			continue
		}

		if !candidate.HasLocation {
			response.ExcludedNoLocation++
			continue
		}

		if !candidateHasActiveFCM(candidate, activeFCMMap, hasAuthoritativeFCMMap) {
			response.ExcludedNoFCMToken++
			continue
		}

		if isLocationOld(candidate, targeting.LocationFreshnessHours, now) {
			response.ExcludedOldLocation++
			continue
		}

		previewAlert := &models.Alert{
			Category: req.Category,
			Severity: req.Severity,
			Location: models.AlertLocation{
				Latitude:  req.Latitude,
				Longitude: req.Longitude,
				Country:   targeting.Country,
				Region:    targeting.Region,
			},
			RadiusKm:  targeting.RadiusKm,
			Targeting: targeting,
		}

		zone, insideTarget, _ := classifyCandidateZoneForAlert(
			previewAlert,
			targeting,
			candidate,
		)

		if !insideTarget {
			response.ExcludedOutsideTarget++
			continue
		}

		if targeting.RespectUserPreferences &&
			!targeting.CriticalOverridePreferences &&
			!candidatePreferenceAllowsCategory(candidate, category) {
			response.ExcludedPreferenceOff++
			continue
		}

		if targeting.RespectUserPreferences &&
			targeting.CriticalOverridePreferences &&
			severity != "critical" &&
			!candidatePreferenceAllowsCategory(candidate, category) {
			response.ExcludedPreferenceOff++
			continue
		}

		if zone == models.AlertTargetZoneDanger {
			response.DangerRecipients++
			continue
		}

		if zone == models.AlertTargetZoneAwareness {
			response.AwarenessRecipients++
			continue
		}

		response.ExcludedOutsideTarget++
	}

	response.TotalPushRecipients =
		response.DangerRecipients +
			response.AwarenessRecipients

	response.Warnings = appendPreviewWarnings(response, targeting)

	return response, nil
}

func buildPreviewTargeting(
	req dto.AlertTargetingPreviewRequest,
) models.AlertTargeting {
	mode := strings.ToLower(strings.TrimSpace(req.Targeting.Mode))

	if mode == "" {
		if len(req.Targeting.Polygon) >= 3 {
			mode = models.AlertTargetingModePolygon
		} else if req.Targeting.RadiusKm > 0 {
			mode = models.AlertTargetingModeRadius
		} else if strings.TrimSpace(req.Targeting.Region) != "" {
			mode = models.AlertTargetingModeRegion
		} else if strings.TrimSpace(req.Targeting.Country) != "" {
			mode = models.AlertTargetingModeCountry
		} else {
			mode = models.AlertTargetingModeRadius
		}
	}

	radiusKm := req.Targeting.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 5
	}

	awarenessRadiusKm := req.Targeting.AwarenessRadiusKm
	if awarenessRadiusKm <= 0 {
		awarenessRadiusKm = radiusKm * 2
	}

	if awarenessRadiusKm < radiusKm {
		awarenessRadiusKm = radiusKm
	}

	locationFreshnessHours := req.Targeting.LocationFreshnessHours
	if locationFreshnessHours <= 0 {
		locationFreshnessHours = 24
	}

	polygon := make([]models.AlertGeoPoint, 0, len(req.Targeting.Polygon))
	for _, point := range req.Targeting.Polygon {
		polygon = append(polygon, models.AlertGeoPoint{
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
		})
	}

	return models.AlertTargeting{
		Mode: mode,

		RadiusKm:          radiusKm,
		AwarenessRadiusKm: awarenessRadiusKm,
		Polygon:           polygon,

		Country:  strings.TrimSpace(req.Targeting.Country),
		Region:   strings.TrimSpace(req.Targeting.Region),
		District: strings.TrimSpace(req.Targeting.District),

		LocationFreshnessHours:      locationFreshnessHours,
		RespectUserPreferences:      true,
		CriticalOverridePreferences: true,

		ConfirmNationalAlert: req.Targeting.ConfirmNationalAlert,
		NationalAlertReason:  strings.TrimSpace(req.Targeting.NationalAlertReason),
	}
}

func validatePreviewTargeting(
	req dto.AlertTargetingPreviewRequest,
	targeting models.AlertTargeting,
) error {
	switch targeting.Mode {
	case models.AlertTargetingModeRadius:
		if req.Latitude == 0 || req.Longitude == 0 {
			return errors.New("latitude and longitude are required for radius targeting")
		}

		if targeting.RadiusKm <= 0 {
			return errors.New("radiusKm must be greater than 0 for radius targeting")
		}

	case models.AlertTargetingModePolygon:
		if len(targeting.Polygon) < 3 {
			return errors.New("at least 3 polygon points are required for polygon targeting")
		}

	case models.AlertTargetingModeRegion:
		if strings.TrimSpace(targeting.Region) == "" {
			return errors.New("region is required for region targeting")
		}

	case models.AlertTargetingModeCountry:
		if strings.TrimSpace(targeting.Country) == "" {
			return errors.New("country is required for country targeting")
		}

	case models.AlertTargetingModeNational:
		if !targeting.ConfirmNationalAlert {
			return errors.New("confirmNationalAlert must be true for national targeting")
		}

		if strings.TrimSpace(targeting.NationalAlertReason) == "" {
			return errors.New("nationalAlertReason is required for national targeting")
		}

	default:
		return errors.New("unsupported targeting mode")
	}

	return nil
}

func classifyCandidateZone(
	req dto.AlertTargetingPreviewRequest,
	targeting models.AlertTargeting,
	candidate repositories.AlertTargetCandidate,
) (string, bool) {
	switch targeting.Mode {
	case models.AlertTargetingModeRadius:
		distance := alertDistanceKm(
			req.Latitude,
			req.Longitude,
			candidate.Latitude,
			candidate.Longitude,
		)

		if distance <= targeting.RadiusKm {
			return models.AlertTargetZoneDanger, true
		}

		if distance <= targeting.AwarenessRadiusKm {
			return models.AlertTargetZoneAwareness, true
		}

		return "", false

	case models.AlertTargetingModePolygon:
		if pointInsidePolygon(
			candidate.Latitude,
			candidate.Longitude,
			targeting.Polygon,
		) {
			return models.AlertTargetZoneDanger, true
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
			return models.AlertTargetZoneAwareness, true
		}

		return "", false

	case models.AlertTargetingModeRegion:
		if sameText(candidate.Region, targeting.Region) {
			return models.AlertTargetZoneDanger, true
		}

		return "", false

	case models.AlertTargetingModeCountry:
		if sameText(candidate.Country, targeting.Country) {
			return models.AlertTargetZoneDanger, true
		}

		return "", false

	case models.AlertTargetingModeNational:
		return models.AlertTargetZoneDanger, true

	default:
		return "", false
	}
}

func isLocationOld(
	candidate repositories.AlertTargetCandidate,
	freshnessHours int,
	now time.Time,
) bool {
	if freshnessHours <= 0 {
		freshnessHours = 24
	}

	if candidate.LocationUpdatedAt == nil {
		// Temporary rule:
		// If your current user location model does not store locationUpdatedAt yet,
		// do not exclude the user in preview.
		// Later we will enforce this once UpdateLocation saves the timestamp.
		return false
	}

	threshold := now.Add(-time.Duration(freshnessHours) * time.Hour)

	return candidate.LocationUpdatedAt.Before(threshold)
}

func candidatePreferenceAllowsCategory(
	candidate repositories.AlertTargetCandidate,
	category string,
) bool {
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return true
	}

	if len(candidate.AlertPreferences) == 0 {
		return true
	}

	if value, exists := candidate.AlertPreferences[category]; exists {
		return value
	}

	if value, exists := candidate.AlertPreferences[category+"_alerts"]; exists {
		return value
	}

	if value, exists := candidate.AlertPreferences["enable_"+category]; exists {
		return value
	}

	if value, exists := candidate.AlertPreferences["critical_alerts"]; exists {
		return value
	}

	return true
}

func appendPreviewWarnings(
	response *dto.AlertTargetingPreviewResponse,
	targeting models.AlertTargeting,
) []string {
	warnings := response.Warnings

	if targeting.Mode == models.AlertTargetingModeCountry {
		warnings = append(
			warnings,
			"Country targeting may notify too many people. Use radius or polygon for local incidents.",
		)
	}

	if targeting.Mode == models.AlertTargetingModeNational {
		warnings = append(
			warnings,
			"National targeting should be used only for national emergencies.",
		)
	}

	if response.TotalPushRecipients == 0 {
		warnings = append(
			warnings,
			"No users will receive push notifications with the current targeting settings.",
		)
	}

	if response.ExcludedOldLocation > 0 {
		warnings = append(
			warnings,
			"Some users were excluded because their location is older than the freshness limit.",
		)
	}

	return warnings
}

func sameText(a string, b string) bool {
	return strings.EqualFold(
		strings.TrimSpace(a),
		strings.TrimSpace(b),
	)
}
