package services

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"disaster_alert_backend/internal/authz"
	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/jobs"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"
	"disaster_alert_backend/internal/repositories"
	"disaster_alert_backend/internal/utils"
	"disaster_alert_backend/internal/websocket"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AlertService struct {
	alertRepo              *repositories.AlertRepository
	userRepo               *repositories.UserRepository
	fcmTokenRepo           *repositories.FCMTokenRepository
	alertDeliveryRepo      *repositories.AlertDeliveryRepository
	notificationDispatcher NotificationDispatcher
	appNotificationService *AppNotificationService
	broadcaster            *websocket.Broadcaster
}

func NewAlertService(
	alertRepo *repositories.AlertRepository,
	userRepo *repositories.UserRepository,
	notificationDispatcher NotificationDispatcher,
	appNotificationService *AppNotificationService,
	broadcaster *websocket.Broadcaster,
) *AlertService {
	return &AlertService{
		alertRepo:              alertRepo,
		userRepo:               userRepo,
		notificationDispatcher: notificationDispatcher,
		appNotificationService: appNotificationService,
		broadcaster:            broadcaster,
	}
}

func (s *AlertService) SetAlertDeliveryRepository(
	alertDeliveryRepo *repositories.AlertDeliveryRepository,
) {
	s.alertDeliveryRepo = alertDeliveryRepo
}

func (s *AlertService) SetFCMTokenRepository(
	fcmTokenRepo *repositories.FCMTokenRepository,
) {
	s.fcmTokenRepo = fcmTokenRepo
}

func (s *AlertService) CreateAlert(
	userID string,
	req dto.CreateAlertRequest,
) (*models.Alert, error) {
	creatorID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	now := time.Now().UTC()

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = "active"
	}

	sourceType := strings.TrimSpace(req.SourceType)
	if sourceType == "" {
		sourceType = "internal"
	}

	sourceName := strings.TrimSpace(req.SourceName)
	if sourceName == "" {
		sourceName = "manual"
	}

	eventTime := parseOptionalTime(req.EventTime)

	expiresAt := parseOptionalTime(req.ExpiresAt)
	if expiresAt == nil {
		defaultExpiry := now.Add(7 * 24 * time.Hour)
		expiresAt = &defaultExpiry
	}

	radiusKm := req.RadiusKm
	if radiusKm <= 0 {
		radiusKm = 20
	}

	confidence := req.Confidence
	if confidence <= 0 {
		confidence = 1.0
	}

	summary := strings.TrimSpace(req.Summary)
	if summary == "" {
		summary = strings.TrimSpace(req.Description)
	}

	category := strings.TrimSpace(req.Category)

	priorityScore := req.PriorityScore
	if priorityScore <= 0 {
		priorityScore = manualAlertPriorityScore(req.Severity, category)
	}

	priorityLabel := strings.TrimSpace(req.PriorityLabel)
	if priorityLabel == "" {
		priorityLabel = manualAlertPriorityLabel(priorityScore)
	}

	targeting := buildAlertTargetingFromRequest(req)

	route := ResolveAutoRoute(
		req.AccessCategoryID,
		req.AccessCategorySlug,
		req.AccessCategoryName,
		category,
	)

	route = ApplyAlertRoutingOverrides(
		route,
		req.OwnerOrganisationID,
		req.LeadOrganisationID,
		req.AssignedOrgIDs,
		req.VisibleToOrgIDs,
	)

	alert := models.Alert{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		Summary:     summary,

		Category:           category,
		AccessCategoryID:   route.AccessCategoryID,
		AccessCategorySlug: route.AccessCategorySlug,
		AccessCategoryName: route.AccessCategoryName,

		OwnerOrganisationID: route.OwnerOrganisationID,
		LeadOrganisationID:  route.LeadOrganisationID,
		AssignedOrgIDs:      route.AssignedOrgIDs,
		VisibleToOrgIDs:     route.VisibleToOrgIDs,

		Targeting: targeting,

		Severity: strings.TrimSpace(req.Severity),
		Status:   status,
		Location: models.AlertLocation{
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Address:   strings.TrimSpace(req.Address),
			Country:   strings.TrimSpace(req.Country),
			Region:    strings.TrimSpace(req.Region),
		},
		RadiusKm:           radiusKm,
		SafetyInstructions: req.SafetyInstructions,
		SourceType:         sourceType,
		SourceName:         sourceName,
		ExternalID:         strings.TrimSpace(req.ExternalID),
		SourceURL:          strings.TrimSpace(req.SourceURL),
		ImageURLs:          req.ImageURLs,
		VideoURLs:          req.VideoURLs,
		Tags:               normalizeAlertTags(req.Tags, category, req.Severity, sourceName),
		PriorityScore:      priorityScore,
		PriorityLabel:      priorityLabel,
		IsBreaking:         priorityScore >= 85,
		IsVerified:         sourceType == "internal",
		CreatedBy:          &creatorID,
		EventTime:          eventTime,
		ExpiresAt:          expiresAt,
		Confidence:         confidence,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	createdAlert, err := s.alertRepo.Create(alert)
	if err != nil {
		return nil, err
	}

	if createdAlert.Status == "active" {
		go s.notifyUsersInAlertTarget(createdAlert, models.AlertDeliveryTypeInitial)
	}

	if createdAlert.Status == "active" && s.broadcaster != nil {
		s.broadcaster.BroadcastAlertCreated(
			createdAlert.Location.Country,
			buildAlertPayload(createdAlert),
		)
	}

	return createdAlert, nil
}

func (s *AlertService) GetAlertDeliveryHistory(
	alertID string,
) ([]models.AlertDeliveryBatch, error) {
	if s.alertDeliveryRepo == nil {
		return []models.AlertDeliveryBatch{}, nil
	}

	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	return s.alertDeliveryRepo.FindBatchesByAlertID(objectID)
}

func (s *AlertService) GetAlertRecipientDeliveries(
	alertID string,
	limit int64,
) ([]models.AlertRecipientDelivery, error) {
	if s.alertDeliveryRepo == nil {
		return []models.AlertRecipientDelivery{}, nil
	}

	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	return s.alertDeliveryRepo.FindRecipientsByAlertID(objectID, limit)
}

func (s *AlertService) GetAlerts() ([]models.Alert, error) {
	return s.alertRepo.FindAll()
}

func (s *AlertService) GetAlertsForPrivilege(
	privilegeCtx *authz.PrivilegeContext,
) ([]models.Alert, error) {
	if privilegeCtx == nil {
		return nil, errors.New("privilege context not found")
	}

	if !authz.HasPermission(privilegeCtx, permissions.AlertsRead) {
		return nil, errors.New("you do not have permission to read alerts")
	}

	alerts, err := s.alertRepo.FindAll()
	if err != nil {
		return nil, err
	}

	if authz.IsGlobal(privilegeCtx) {
		return alerts, nil
	}

	filteredAlerts := make([]models.Alert, 0)

	for i := range alerts {
		if canPrivilegeAccessAlert(
			privilegeCtx,
			&alerts[i],
			permissions.AlertsRead,
		) {
			filteredAlerts = append(filteredAlerts, alerts[i])
		}
	}

	return filteredAlerts, nil
}

func (s *AlertService) GetActiveAlerts() ([]models.Alert, error) {
	return s.alertRepo.FindActive()
}

func (s *AlertService) GetActiveAlertsWithFilters(
	filter repositories.AlertFilter,
) ([]models.Alert, error) {
	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetLocalAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if strings.TrimSpace(country) == "" {
		return []models.Alert{}, nil
	}

	if limit <= 0 {
		limit = 50
	}

	return s.alertRepo.FindActiveWithFilters(repositories.AlertFilter{
		Country: country,
		Limit:   limit,
	})
}

func (s *AlertService) GetGlobalAlerts(
	userCountry string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	return s.alertRepo.FindActiveWithFilters(repositories.AlertFilter{
		ExcludeCountry: userCountry,
		Limit:          limit,
	})
}

func (s *AlertService) GetWeatherAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := repositories.AlertFilter{
		Category: "weather",
		Limit:    limit,
	}

	if strings.TrimSpace(country) != "" {
		filter.Country = country
	}

	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetHealthAlerts(
	country string,
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 50
	}

	filter := repositories.AlertFilter{
		Category: "health",
		Limit:    limit,
	}

	if strings.TrimSpace(country) != "" {
		filter.Country = country
	}

	return s.alertRepo.FindActiveWithFilters(filter)
}

func (s *AlertService) GetAlertByID(alertID string) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	return s.alertRepo.FindByID(objectID)
}

func (s *AlertService) GetAlertByIDForPrivilege(
	alertID string,
	privilegeCtx *authz.PrivilegeContext,
) (*models.Alert, error) {
	if privilegeCtx == nil {
		return nil, errors.New("privilege context not found")
	}

	if !authz.HasPermission(privilegeCtx, permissions.AlertsRead) {
		return nil, errors.New("you do not have permission to read alerts")
	}

	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	alert, err := s.alertRepo.FindByID(objectID)
	if err != nil || alert == nil {
		return nil, errors.New("alert not found")
	}

	if authz.IsGlobal(privilegeCtx) {
		return alert, nil
	}

	if !canPrivilegeAccessAlert(
		privilegeCtx,
		alert,
		permissions.AlertsRead,
	) {
		return nil, errors.New("you do not have access to this alert")
	}

	return alert, nil
}

func (s *AlertService) UpdateAlertStatus(
	alertID string,
	status string,
) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	existingAlert, err := s.alertRepo.FindByID(objectID)
	if err != nil {
		return nil, err
	}

	updatedAlert, err := s.alertRepo.UpdateStatus(objectID, status)
	if err != nil {
		return nil, err
	}

	wasNotActive := existingAlert == nil || existingAlert.Status != "active"
	isNowActive := updatedAlert.Status == "active"

	if wasNotActive && isNowActive {
		go s.notifyUsersInAlertTarget(updatedAlert, models.AlertDeliveryTypeInitial)
	}

	if updatedAlert.Status == "active" && s.broadcaster != nil {
		s.broadcaster.BroadcastAlertApproved(
			updatedAlert.Location.Country,
			buildAlertPayload(updatedAlert),
		)
	}

	return updatedAlert, nil
}

func (s *AlertService) UpdateAlertStatusForPrivilege(
	alertID string,
	status string,
	privilegeCtx *authz.PrivilegeContext,
) (*models.Alert, error) {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return nil, errors.New("invalid alert id")
	}

	alert, err := s.alertRepo.FindByID(objectID)
	if err != nil || alert == nil {
		return nil, errors.New("alert not found")
	}

	if !canPrivilegeAccessAlert(
		privilegeCtx,
		alert,
		permissions.AlertsUpdate,
	) {
		return nil, errors.New("you do not have access to update this alert")
	}

	return s.UpdateAlertStatus(alertID, status)
}

func (s *AlertService) DeleteAlert(alertID string) error {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return errors.New("invalid alert id")
	}

	return s.alertRepo.Delete(objectID)
}

func (s *AlertService) DeleteAlertForPrivilege(
	alertID string,
	privilegeCtx *authz.PrivilegeContext,
) error {
	objectID, err := primitive.ObjectIDFromHex(alertID)
	if err != nil {
		return errors.New("invalid alert id")
	}

	alert, err := s.alertRepo.FindByID(objectID)
	if err != nil || alert == nil {
		return errors.New("alert not found")
	}

	if !canPrivilegeAccessAlert(
		privilegeCtx,
		alert,
		permissions.AlertsDelete,
	) {
		return errors.New("you do not have access to delete this alert")
	}

	return s.DeleteAlert(alertID)
}

func (s *AlertService) GetNearbyAlerts(
	latitude float64,
	longitude float64,
	radiusKm float64,
) ([]models.Alert, error) {
	if radiusKm <= 0 {
		radiusKm = 100
	}

	return s.alertRepo.FindActiveNearby(latitude, longitude, radiusKm)
}

func (s *AlertService) GetNearbyAlertsWithFilters(
	latitude float64,
	longitude float64,
	radiusKm float64,
	filter repositories.AlertFilter,
) ([]models.Alert, error) {
	if radiusKm <= 0 {
		radiusKm = 100
	}

	return s.alertRepo.FindActiveNearbyWithFilters(
		latitude,
		longitude,
		radiusKm,
		filter,
	)
}

func (s *AlertService) GetCriticalGlobalAlerts(
	limit int,
) ([]models.Alert, error) {
	if limit <= 0 {
		limit = 20
	}

	return s.alertRepo.FindCriticalGlobal(limit)
}

func (s *AlertService) notifyUsersInAlertCountry(alert *models.Alert) {
	if alert == nil {
		return
	}

	if s.userRepo == nil {
		return
	}

	country := strings.TrimSpace(alert.Location.Country)
	if country == "" {
		return
	}

	users, err := s.userRepo.FindUsersByCountry(country)
	if err != nil {
		return
	}

	if len(users) == 0 {
		return
	}

	title := strings.TrimSpace(alert.Title)
	if title == "" {
		title = "New Disaster Alert"
	}

	body := strings.TrimSpace(alert.Description)
	if body == "" {
		body = "A new emergency alert has been issued in your country."
	}

	data := map[string]string{
		"type":        "alert",
		"referenceId": alert.ID.Hex(),
		"alertId":     alert.ID.Hex(),
		"category":    alert.Category,
		"severity":    alert.Severity,
		"source":      alert.SourceName,
		"country":     alert.Location.Country,
		"latitude":    floatToString(alert.Location.Latitude),
		"longitude":   floatToString(alert.Location.Longitude),
	}

	userIDs := make([]primitive.ObjectID, 0, len(users))

	for _, user := range users {
		if user.ID.IsZero() {
			continue
		}

		userIDs = append(userIDs, user.ID)

		if s.notificationDispatcher != nil {
			s.notificationDispatcher.Dispatch(jobs.NotificationJob{
				TargetType: jobs.TargetUser,
				UserID:     user.ID.Hex(),
				Title:      title,
				Body:       body,
				Data:       data,
			})
		}
	}

	if len(userIDs) == 0 {
		return
	}

	if s.appNotificationService != nil {
		_ = s.appNotificationService.CreateManyForUsers(
			userIDs,
			title,
			body,
			"alert",
			alert.ID.Hex(),
			data,
		)
	}
}

func buildAlertTargetingFromRequest(req dto.CreateAlertRequest) models.AlertTargeting {
	mode := strings.TrimSpace(req.Targeting.Mode)

	if mode == "" {
		if len(req.Targeting.Polygon) >= 3 {
			mode = models.AlertTargetingModePolygon
		} else if req.Targeting.RadiusKm > 0 || req.RadiusKm > 0 {
			mode = models.AlertTargetingModeRadius
		} else if strings.TrimSpace(req.Targeting.Region) != "" ||
			strings.TrimSpace(req.Region) != "" {
			mode = models.AlertTargetingModeRegion
		} else if strings.TrimSpace(req.Targeting.Country) != "" ||
			strings.TrimSpace(req.Country) != "" {
			mode = models.AlertTargetingModeCountry
		} else {
			mode = models.AlertTargetingModeRadius
		}
	}

	mode = strings.ToLower(strings.TrimSpace(mode))

	radiusKm := req.Targeting.RadiusKm
	if radiusKm <= 0 {
		radiusKm = req.RadiusKm
	}

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

	country := strings.TrimSpace(req.Targeting.Country)
	if country == "" {
		country = strings.TrimSpace(req.Country)
	}

	region := strings.TrimSpace(req.Targeting.Region)
	if region == "" {
		region = strings.TrimSpace(req.Region)
	}

	targeting := models.AlertTargeting{
		Mode: mode,

		RadiusKm:          radiusKm,
		AwarenessRadiusKm: awarenessRadiusKm,
		Polygon:           polygon,

		Country:  country,
		Region:   region,
		District: strings.TrimSpace(req.Targeting.District),

		LocationFreshnessHours:      locationFreshnessHours,
		RespectUserPreferences:      true,
		CriticalOverridePreferences: true,

		ConfirmNationalAlert: req.Targeting.ConfirmNationalAlert,
		NationalAlertReason:  strings.TrimSpace(req.Targeting.NationalAlertReason),

		NotificationStatus: models.AlertNotificationStatusDraft,
	}

	return targeting
}

func parseOptionalTime(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}

	return &parsed
}

func floatToString(value float64) string {
	return strconv.FormatFloat(value, 'f', 6, 64)
}

func buildAlertPayload(alert *models.Alert) map[string]interface{} {
	if alert == nil {
		return map[string]interface{}{}
	}

	return map[string]interface{}{
		"id":                  alert.ID.Hex(),
		"title":               alert.Title,
		"description":         alert.Description,
		"summary":             alert.Summary,
		"category":            alert.Category,
		"accessCategoryId":    alert.AccessCategoryID,
		"accessCategorySlug":  alert.AccessCategorySlug,
		"accessCategoryName":  alert.AccessCategoryName,
		"ownerOrganisationId": alert.OwnerOrganisationID,
		"leadOrganisationId":  alert.LeadOrganisationID,
		"assignedOrgIds":      alert.AssignedOrgIDs,
		"visibleToOrgIds":     alert.VisibleToOrgIDs,
		"targeting":           alert.Targeting,
		"severity":            alert.Severity,
		"status":              alert.Status,
		"latitude":            alert.Location.Latitude,
		"longitude":           alert.Location.Longitude,
		"address":             alert.Location.Address,
		"country":             alert.Location.Country,
		"region":              alert.Location.Region,
		"radiusKm":            alert.RadiusKm,
		"safetyInstructions":  alert.SafetyInstructions,
		"sourceType":          alert.SourceType,
		"sourceName":          alert.SourceName,
		"externalId":          alert.ExternalID,
		"sourceUrl":           alert.SourceURL,
		"imageUrls":           alert.ImageURLs,
		"videoUrls":           alert.VideoURLs,
		"tags":                alert.Tags,
		"priorityScore":       alert.PriorityScore,
		"priorityLabel":       alert.PriorityLabel,
		"isBreaking":          alert.IsBreaking,
		"isVerified":          alert.IsVerified,
		"eventTime":           alert.EventTime,
		"expiresAt":           alert.ExpiresAt,
		"confidence":          alert.Confidence,
		"lastSyncedAt":        alert.LastSyncedAt,
		"createdAt":           alert.CreatedAt,
		"updatedAt":           alert.UpdatedAt,
	}
}

func manualAlertPriorityScore(severity string, category string) int {
	score := 0

	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical":
		score += 80
	case "high":
		score += 65
	case "medium":
		score += 45
	case "low":
		score += 25
	default:
		score += 35
	}

	switch strings.ToLower(strings.TrimSpace(category)) {
	case "earthquake", "flood", "fire", "weather", "health", "conflict":
		score += 10
	case "volcano", "drought":
		score += 8
	default:
		score += 5
	}

	if score > 100 {
		score = 100
	}

	return score
}

func manualAlertPriorityLabel(score int) string {
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

func normalizeAlertTags(
	inputTags []string,
	category string,
	severity string,
	sourceName string,
) []string {
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

	for _, tag := range inputTags {
		add(tag)
	}

	add(category)
	add(severity)
	add(sourceName)

	return tags
}

func canPrivilegeAccessAlert(
	privilegeCtx *authz.PrivilegeContext,
	alert *models.Alert,
	requiredAction string,
) bool {
	if privilegeCtx == nil || alert == nil {
		return false
	}

	if !authz.HasPermission(privilegeCtx, requiredAction) {
		return false
	}

	if authz.IsGlobal(privilegeCtx) {
		return true
	}

	orgID := normalizeAlertAccessText(privilegeCtx.OrganisationID)
	if orgID == "" {
		return false
	}

	categorySlug := normalizeAlertAccessText(alert.AccessCategorySlug)
	if categorySlug == "" {
		categorySlug = normalizeAlertAccessText(alert.Category)
	}

	if categorySlug == "" {
		return false
	}

	if !privilegeGrantAllowsAlertAction(
		privilegeCtx,
		categorySlug,
		alert,
		requiredAction,
	) {
		return false
	}

	if orgID == normalizeAlertAccessText(alert.OwnerOrganisationID) {
		return true
	}

	if orgID == normalizeAlertAccessText(alert.LeadOrganisationID) {
		return true
	}

	if containsNormalizedAlertAccessText(alert.AssignedOrgIDs, orgID) {
		return true
	}

	if containsNormalizedAlertAccessText(alert.VisibleToOrgIDs, orgID) {
		return true
	}

	return false
}

func privilegeGrantAllowsAlertAction(
	privilegeCtx *authz.PrivilegeContext,
	categorySlug string,
	alert *models.Alert,
	requiredAction string,
) bool {
	if privilegeCtx == nil {
		return false
	}

	// If there are no grants, allow organisation/record scope to decide,
	// as long as the flat permission already exists.
	if len(privilegeCtx.Grants) == 0 {
		return true
	}

	for _, grant := range privilegeCtx.Grants {
		grantCategorySlug := normalizeAlertAccessText(grant.CategorySlug)

		if grantCategorySlug != "" && grantCategorySlug != categorySlug {
			continue
		}

		if !containsNormalizedAlertAccessText(grant.Actions, requiredAction) {
			continue
		}

		if !grantLocationAllowsAlert(grant, alert) {
			continue
		}

		return true
	}

	return false
}

func grantLocationAllowsAlert(
	grant models.PrivilegeGrant,
	alert *models.Alert,
) bool {
	if alert == nil {
		return false
	}

	// If no countries are specified in the grant, do not restrict by country.
	if len(grant.Countries) > 0 {
		alertCountry := utils.NormalizeCountryCode(alert.Location.Country)
		if alertCountry == "" {
			alertCountry = utils.NormalizeCountryCode(alert.Targeting.Country)
		}

		if !grantCountriesContainCountry(grant.Countries, alertCountry) {
			return false
		}
	}

	// If no regions are specified in the grant, do not restrict by region.
	if len(grant.Regions) > 0 {
		alertRegion := normalizeAlertAccessText(alert.Location.Region)
		if alertRegion == "" {
			alertRegion = normalizeAlertAccessText(alert.Targeting.Region)
		}

		if !containsNormalizedAlertAccessText(grant.Regions, alertRegion) {
			return false
		}
	}

	return true
}

func grantCountriesContainCountry(
	grantCountries []string,
	alertCountry string,
) bool {
	alertCountry = utils.NormalizeCountryCode(alertCountry)

	if alertCountry == "" {
		return false
	}

	for _, grantCountry := range grantCountries {
		if utils.NormalizeCountryCode(grantCountry) == alertCountry {
			return true
		}
	}

	return false
}

func containsNormalizedAlertAccessText(
	values []string,
	target string,
) bool {
	target = normalizeAlertAccessText(target)

	if target == "" {
		return false
	}

	for _, value := range values {
		if normalizeAlertAccessText(value) == target {
			return true
		}
	}

	return false
}

func normalizeAlertAccessText(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func alertScope(alert *models.Alert) authz.RecordScope {
	if alert == nil {
		return authz.RecordScope{}
	}

	categorySlug := strings.TrimSpace(alert.AccessCategorySlug)
	if categorySlug == "" {
		categorySlug = strings.TrimSpace(alert.Category)
	}

	country := utils.NormalizeCountryCode(alert.Location.Country)
	if country == "" {
		country = utils.NormalizeCountryCode(alert.Targeting.Country)
	}

	region := strings.TrimSpace(alert.Location.Region)
	if region == "" {
		region = strings.TrimSpace(alert.Targeting.Region)
	}

	return authz.RecordScope{
		AccessCategoryID:    alert.AccessCategoryID,
		AccessCategorySlug:  categorySlug,
		AccessCategoryName:  alert.AccessCategoryName,
		OwnerOrganisationID: alert.OwnerOrganisationID,
		LeadOrganisationID:  alert.LeadOrganisationID,
		AssignedOrgIDs:      alert.AssignedOrgIDs,
		VisibleToOrgIDs:     alert.VisibleToOrgIDs,
		Country:             country,
		Region:              region,
	}
}
