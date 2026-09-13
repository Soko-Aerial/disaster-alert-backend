package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
)

type AccessCategoryService struct {
	repo *repositories.AccessCategoryRepository
}

func NewAccessCategoryService(repo *repositories.AccessCategoryRepository) *AccessCategoryService {
	return &AccessCategoryService{
		repo: repo,
	}
}

func (s *AccessCategoryService) EnsureDefaultSystemCategories(ctx context.Context) error {
	for _, category := range defaultSystemCategories() {
		if err := s.repo.UpsertSystemCategory(ctx, &category); err != nil {
			return err
		}
	}

	return nil
}

func (s *AccessCategoryService) CreateCategory(
	ctx context.Context,
	req dto.CreateAccessCategoryRequest,
) (*models.AccessCategory, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("name is required")
	}

	kind := normalizeCategoryKind(req.Kind)
	visibility := normalizeCategoryVisibility(req.Visibility)
	slug := normalizeAccessCategorySlug(req.Slug)

	if slug == "" {
		slug = normalizeAccessCategorySlug(name)
	}

	allowedActions := permissions.RemoveDuplicatePermissions(req.AllowedActions)
	invalid := permissions.InvalidPermissions(allowedActions)
	if len(invalid) > 0 {
		return nil, errors.New("allowedActions contains invalid permissions: " + strings.Join(invalid, ", "))
	}

	now := time.Now().UTC()

	category := &models.AccessCategory{
		Name:                  name,
		Slug:                  slug,
		Description:           strings.TrimSpace(req.Description),
		Kind:                  kind,
		Visibility:            visibility,
		OwnerOrganisationID:   strings.TrimSpace(req.OwnerOrganisationID),
		OwnerOrganisationName: strings.TrimSpace(req.OwnerOrganisationName),
		PreferenceKey:         strings.TrimSpace(req.PreferenceKey),
		AllowedActions:        allowedActions,
		IsSystem:              false,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.repo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *AccessCategoryService) GetCategories(
	ctx context.Context,
	kind string,
	ownerOrganisationID string,
	includeInactive bool,
	limit int64,
) ([]models.AccessCategory, error) {
	return s.repo.GetAll(
		ctx,
		normalizeOptionalCategoryKind(kind),
		strings.TrimSpace(ownerOrganisationID),
		includeInactive,
		limit,
	)
}

func (s *AccessCategoryService) GetCategoryByID(
	ctx context.Context,
	id string,
) (*models.AccessCategory, error) {
	category, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if category == nil {
		return nil, errors.New("access category not found")
	}

	return category, nil
}

func (s *AccessCategoryService) UpdateCategory(
	ctx context.Context,
	id string,
	req dto.UpdateAccessCategoryRequest,
) (*models.AccessCategory, error) {
	update := bson.M{}

	if strings.TrimSpace(req.Name) != "" {
		update["name"] = strings.TrimSpace(req.Name)
	}

	if strings.TrimSpace(req.Slug) != "" {
		update["slug"] = normalizeAccessCategorySlug(req.Slug)
	}

	if strings.TrimSpace(req.Description) != "" {
		update["description"] = strings.TrimSpace(req.Description)
	}

	if strings.TrimSpace(req.Kind) != "" {
		update["kind"] = normalizeCategoryKind(req.Kind)
	}

	if strings.TrimSpace(req.Visibility) != "" {
		update["visibility"] = normalizeCategoryVisibility(req.Visibility)
	}

	if strings.TrimSpace(req.OwnerOrganisationID) != "" {
		update["ownerOrganisationId"] = strings.TrimSpace(req.OwnerOrganisationID)
	}

	if strings.TrimSpace(req.OwnerOrganisationName) != "" {
		update["ownerOrganisationName"] = strings.TrimSpace(req.OwnerOrganisationName)
	}

	if strings.TrimSpace(req.PreferenceKey) != "" {
		update["preferenceKey"] = strings.TrimSpace(req.PreferenceKey)
	}

	if req.AllowedActions != nil {
		allowedActions := permissions.RemoveDuplicatePermissions(req.AllowedActions)
		invalid := permissions.InvalidPermissions(allowedActions)
		if len(invalid) > 0 {
			return nil, errors.New("allowedActions contains invalid permissions: " + strings.Join(invalid, ", "))
		}

		update["allowedActions"] = allowedActions
	}

	if req.IsActive != nil {
		update["isActive"] = *req.IsActive
	}

	if len(update) == 0 {
		return s.GetCategoryByID(ctx, id)
	}

	return s.repo.Update(ctx, id, update)
}

func (s *AccessCategoryService) DeactivateCategory(
	ctx context.Context,
	id string,
) (*models.AccessCategory, error) {
	return s.repo.Deactivate(ctx, id)
}

func defaultSystemCategories() []models.AccessCategory {
	allActions := permissions.AllPermissions

	return []models.AccessCategory{
		systemCategory("Fire", "fire", "Fire outbreaks, smoke, explosions, and fire emergencies", "fire", allActions),
		systemCategory("Flood", "flood", "Flooding, heavy water, blocked roads, and flood emergencies", "flood", allActions),
		systemCategory("Weather", "weather", "Storms, heavy rain, wind, heat, and severe weather alerts", "weather", allActions),
		systemCategory("Earthquake", "earthquake", "Earthquake and tremor-related alerts", "earthquake", allActions),
		systemCategory("Health", "health", "Disease outbreak, contamination, and public health alerts", "health", allActions),
		systemCategory("Conflict", "conflict", "Conflict, violence, and instability alerts", "conflict", allActions),
		systemCategory("Drought", "drought", "Drought and water shortage alerts", "drought", allActions),
		systemCategory("Protests", "protests", "Protests, demonstrations, and civil unrest alerts", "protests", allActions),
		systemCategory("Robbery", "robbery", "Robbery, theft, and security incidents", "robbery", allActions),
		systemCategory("Munitions", "munitions", "Munitions, explosives, and dangerous ordnance alerts", "munitions", allActions),
		systemCategory("Galamsey", "galamsey", "Illegal mining and galamsey-related alerts", "galamsey", allActions),
		systemCategory("Unverified Activity", "unverified_activity", "Unverified activity or suspicious incident reports", "unverifiedActivity", allActions),
		systemCategory("Critical Alerts", "critical_alerts", "Critical or highest-priority alerts", "criticalAlerts", allActions),

		// Extra public/admin categories that are useful but not currently in AlertPreference.
		systemCategory("Medical", "medical", "Medical emergency and ambulance-related cases", "", allActions),
		systemCategory("Security", "security", "Security threats and public safety cases", "", allActions),
		systemCategory("Accident", "accident", "Road, transport, workplace, or public accidents", "", allActions),
		systemCategory("Other", "other", "Other incidents that do not fit a known category", "", allActions),
	}
}

func systemCategory(
	name string,
	slug string,
	description string,
	preferenceKey string,
	allowedActions []string,
) models.AccessCategory {
	now := time.Now().UTC()

	return models.AccessCategory{
		Name:                  name,
		Slug:                  slug,
		Description:           description,
		Kind:                  models.AccessCategoryKindBoth,
		Visibility:            models.AccessCategoryVisibilitySystem,
		OwnerOrganisationID:   "system",
		OwnerOrganisationName: "System",
		PreferenceKey:         preferenceKey,
		AllowedActions:        allowedActions,
		IsSystem:              true,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
}

func normalizeCategoryKind(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case models.AccessCategoryKindPublic:
		return models.AccessCategoryKindPublic
	case models.AccessCategoryKindAccess:
		return models.AccessCategoryKindAccess
	case models.AccessCategoryKindBoth:
		return models.AccessCategoryKindBoth
	default:
		return models.AccessCategoryKindBoth
	}
}

func normalizeOptionalCategoryKind(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case models.AccessCategoryKindPublic:
		return models.AccessCategoryKindPublic
	case models.AccessCategoryKindAccess:
		return models.AccessCategoryKindAccess
	case models.AccessCategoryKindBoth:
		return models.AccessCategoryKindBoth
	default:
		return ""
	}
}

func normalizeCategoryVisibility(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case models.AccessCategoryVisibilityShared:
		return models.AccessCategoryVisibilityShared
	case models.AccessCategoryVisibilityPrivate:
		return models.AccessCategoryVisibilityPrivate
	case models.AccessCategoryVisibilitySystem:
		return models.AccessCategoryVisibilitySystem
	default:
		return models.AccessCategoryVisibilityPrivate
	}
}

func normalizeAccessCategorySlug(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}
