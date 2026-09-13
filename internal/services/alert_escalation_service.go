package services

import (
	"errors"
	"strings"

	"disaster_alert_backend/internal/authz"
	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *AlertService) EscalateAlertForPrivilege(
	alertID string,
	req dto.AlertEscalationRequest,
	privilegeCtx *authz.PrivilegeContext,
) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	existingAlert, err := s.alertRepo.FindByID(objectID)
	if err != nil || existingAlert == nil {
		return nil, errors.New("alert not found")
	}

	if !authz.CanAccessRecord(
		privilegeCtx,
		permissions.AlertsEscalate,
		alertScope(existingAlert),
	) {
		return nil, errors.New("you do not have access to escalate this alert")
	}

	newSeverity := strings.ToLower(strings.TrimSpace(req.Severity))
	if newSeverity == "" {
		newSeverity = existingAlert.Severity
	}

	updatedTargeting := mergeEscalationTargeting(existingAlert, req)

	if err := validateEscalationNationalRules(updatedTargeting, privilegeCtx); err != nil {
		return nil, err
	}

	targetingChanged := escalationRequestHasTargetingChange(req)

	criticalUpgrade :=
		!strings.EqualFold(existingAlert.Severity, "critical") &&
			strings.EqualFold(newSeverity, "critical")

	if !targetingChanged && !criticalUpgrade {
		return nil, errors.New("no escalation change detected")
	}

	priorityScore := manualAlertPriorityScore(newSeverity, existingAlert.Category)
	priorityLabel := manualAlertPriorityLabel(priorityScore)

	updatedAlert, err := s.alertRepo.UpdateTargetingAndSeverity(
		objectID,
		newSeverity,
		updatedTargeting,
		updatedTargeting.RadiusKm,
		priorityScore,
		priorityLabel,
	)
	if err != nil {
		return nil, err
	}

	deliveryType := models.AlertDeliveryTypeEscalation

	// Critical rule:
	// If severity is upgraded to critical,
	// previous affected users and newly affected users must receive an update.
	if criticalUpgrade {
		deliveryType = models.AlertDeliveryTypeUpdate
	}

	go s.notifyUsersInAlertTarget(updatedAlert, deliveryType)

	if s.broadcaster != nil {
		s.broadcaster.BroadcastAlertApproved(
			updatedAlert.Location.Country,
			buildAlertPayload(updatedAlert),
		)
	}

	return updatedAlert, nil
}

func mergeEscalationTargeting(
	existingAlert *models.Alert,
	req dto.AlertEscalationRequest,
) models.AlertTargeting {
	targeting := normalizeAlertTargeting(existingAlert)

	mode := strings.ToLower(strings.TrimSpace(req.Targeting.Mode))
	if mode != "" {
		targeting.Mode = mode
	}

	if req.Targeting.RadiusKm > 0 {
		targeting.RadiusKm = req.Targeting.RadiusKm
	} else if req.RadiusKm > 0 {
		targeting.RadiusKm = req.RadiusKm
	}

	if req.Targeting.AwarenessRadiusKm > 0 {
		targeting.AwarenessRadiusKm = req.Targeting.AwarenessRadiusKm
	} else if targeting.RadiusKm > 0 && targeting.AwarenessRadiusKm < targeting.RadiusKm {
		targeting.AwarenessRadiusKm = targeting.RadiusKm * 2
	}

	if len(req.Targeting.Polygon) >= 3 {
		targeting.Polygon = make([]models.AlertGeoPoint, 0, len(req.Targeting.Polygon))

		for _, point := range req.Targeting.Polygon {
			targeting.Polygon = append(targeting.Polygon, models.AlertGeoPoint{
				Latitude:  point.Latitude,
				Longitude: point.Longitude,
			})
		}
	}

	if country := strings.TrimSpace(req.Targeting.Country); country != "" {
		targeting.Country = country
	}

	if region := strings.TrimSpace(req.Targeting.Region); region != "" {
		targeting.Region = region
	}

	if district := strings.TrimSpace(req.Targeting.District); district != "" {
		targeting.District = district
	}

	if req.Targeting.LocationFreshnessHours > 0 {
		targeting.LocationFreshnessHours = req.Targeting.LocationFreshnessHours
	}

	if targeting.LocationFreshnessHours <= 0 {
		targeting.LocationFreshnessHours = 24
	}

	if targeting.RadiusKm <= 0 {
		targeting.RadiusKm = existingAlert.RadiusKm
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

	targeting.RespectUserPreferences = true
	targeting.CriticalOverridePreferences = true

	targeting.ConfirmNationalAlert = req.Targeting.ConfirmNationalAlert
	targeting.NationalAlertReason = strings.TrimSpace(req.Targeting.NationalAlertReason)
	targeting.NotificationStatus = models.AlertNotificationStatusSending

	return targeting
}

func escalationRequestHasTargetingChange(req dto.AlertEscalationRequest) bool {
	if req.RadiusKm > 0 {
		return true
	}

	if strings.TrimSpace(req.Targeting.Mode) != "" {
		return true
	}

	if req.Targeting.RadiusKm > 0 {
		return true
	}

	if req.Targeting.AwarenessRadiusKm > 0 {
		return true
	}

	if len(req.Targeting.Polygon) >= 3 {
		return true
	}

	if strings.TrimSpace(req.Targeting.Country) != "" {
		return true
	}

	if strings.TrimSpace(req.Targeting.Region) != "" {
		return true
	}

	if strings.TrimSpace(req.Targeting.District) != "" {
		return true
	}

	return false
}

func validateEscalationNationalRules(
	targeting models.AlertTargeting,
	privilegeCtx *authz.PrivilegeContext,
) error {
	if targeting.Mode != models.AlertTargetingModeNational {
		return nil
	}

	if privilegeCtx == nil {
		return errors.New("national alerts require a valid privilege context")
	}

	if !permissions.ContainsPermission(
		privilegeCtx.Permissions,
		permissions.AlertsSendNational,
	) {
		return errors.New("national alerts require alerts:send_national permission")
	}

	if !targeting.ConfirmNationalAlert {
		return errors.New("confirmNationalAlert must be true for national alerts")
	}

	if strings.TrimSpace(targeting.NationalAlertReason) == "" {
		return errors.New("nationalAlertReason is required for national alerts")
	}

	return nil
}
