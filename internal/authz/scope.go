package authz

import (
	"strings"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"

	"go.mongodb.org/mongo-driver/bson"
)

type RecordScope struct {
	AccessCategoryID   string
	AccessCategorySlug string
	AccessCategoryName string

	OwnerOrganisationID string
	LeadOrganisationID  string
	AssignedOrgIDs      []string
	VisibleToOrgIDs     []string

	Country  string
	Region   string
	District string
}

func CanAccessRecord(ctx *PrivilegeContext, requiredAction string, record RecordScope) bool {
	if ctx == nil {
		return false
	}

	if !permissions.ContainsPermission(ctx.Permissions, requiredAction) {
		return false
	}

	if IsGlobal(ctx) {
		return true
	}

	orgID := strings.TrimSpace(ctx.OrganisationID)
	if orgID == "" {
		return false
	}

	if !recordVisibleToOrganisation(orgID, record) {
		return false
	}

	if len(ctx.Grants) == 0 {
		return true
	}

	recordSlug := NormalizeScopeSlug(record.AccessCategorySlug)

	for _, grant := range ctx.Grants {
		if !permissions.ContainsPermission(grant.Actions, requiredAction) {
			continue
		}

		grantSlug := NormalizeScopeSlug(grant.CategorySlug)

		if grantSlug != "" && recordSlug != "" && grantSlug != recordSlug {
			continue
		}

		if !grantAllowsGeography(grant, record) {
			continue
		}

		return true
	}

	return false
}

func BuildMongoScopeFilter(
	ctx *PrivilegeContext,
	requiredAction string,
	categoryField string,
	countryField string,
	regionField string,
) bson.M {
	if ctx == nil {
		return impossibleFilter()
	}

	if !permissions.ContainsPermission(ctx.Permissions, requiredAction) {
		return impossibleFilter()
	}

	if IsGlobal(ctx) {
		return bson.M{}
	}

	orgID := strings.TrimSpace(ctx.OrganisationID)
	if orgID == "" {
		return impossibleFilter()
	}

	filter := bson.M{
		"$or": []bson.M{
			{"ownerOrganisationId": orgID},
			{"leadOrganisationId": orgID},
			{"assignedOrgIds": orgID},
			{"visibleToOrgIds": orgID},
		},
	}

	allowedSlugs, hasWildcardGrant := AllowedCategorySlugsForAction(ctx, requiredAction)

	if len(ctx.Grants) > 0 && !hasWildcardGrant {
		if len(allowedSlugs) == 0 {
			return impossibleFilter()
		}

		if categoryField != "" {
			filter[categoryField] = bson.M{
				"$in": allowedSlugs,
			}
		}
	}

	if countryField != "" || regionField != "" {
		geoFilters := BuildGrantGeographyFilter(ctx, requiredAction, countryField, regionField)
		if len(geoFilters) > 0 {
			filter["$and"] = []bson.M{
				{
					"$or": geoFilters,
				},
			}
		}
	}

	return filter
}

func AllowedCategorySlugsForAction(ctx *PrivilegeContext, requiredAction string) ([]string, bool) {
	slugs := []string{}
	seen := map[string]bool{}
	hasWildcardGrant := false

	if ctx == nil {
		return slugs, false
	}

	for _, grant := range ctx.Grants {
		if !permissions.ContainsPermission(grant.Actions, requiredAction) {
			continue
		}

		slug := NormalizeScopeSlug(grant.CategorySlug)

		if slug == "" {
			hasWildcardGrant = true
			continue
		}

		if seen[slug] {
			continue
		}

		seen[slug] = true
		slugs = append(slugs, slug)
	}

	return slugs, hasWildcardGrant
}

func BuildGrantGeographyFilter(
	ctx *PrivilegeContext,
	requiredAction string,
	countryField string,
	regionField string,
) []bson.M {
	filters := []bson.M{}

	if ctx == nil {
		return filters
	}

	for _, grant := range ctx.Grants {
		if !permissions.ContainsPermission(grant.Actions, requiredAction) {
			continue
		}

		filter := bson.M{}

		if countryField != "" && len(grant.Countries) > 0 {
			filter[countryField] = bson.M{"$in": normalizeScopeList(grant.Countries)}
		}

		if regionField != "" && len(grant.Regions) > 0 {
			filter[regionField] = bson.M{"$in": normalizeScopeList(grant.Regions)}
		}

		if len(filter) > 0 {
			filters = append(filters, filter)
		}
	}

	return filters
}

func NormalizeScopeSlug(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func recordVisibleToOrganisation(orgID string, record RecordScope) bool {
	orgID = strings.TrimSpace(orgID)
	if orgID == "" {
		return false
	}

	if strings.EqualFold(record.OwnerOrganisationID, orgID) {
		return true
	}

	if strings.EqualFold(record.LeadOrganisationID, orgID) {
		return true
	}

	if containsStringFold(record.AssignedOrgIDs, orgID) {
		return true
	}

	if containsStringFold(record.VisibleToOrgIDs, orgID) {
		return true
	}

	return false
}

func grantAllowsGeography(grant models.PrivilegeGrant, record RecordScope) bool {
	if len(grant.Countries) > 0 &&
		!containsStringFold(grant.Countries, record.Country) {
		return false
	}

	if len(grant.Regions) > 0 &&
		!containsStringFold(grant.Regions, record.Region) {
		return false
	}

	if len(grant.Districts) > 0 &&
		!containsStringFold(grant.Districts, record.District) {
		return false
	}

	return true
}

func normalizeScopeList(input []string) []string {
	result := []string{}

	for _, value := range input {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		result = append(result, value)
	}

	return result
}

func containsStringFold(values []string, target string) bool {
	target = strings.TrimSpace(target)
	if target == "" {
		return false
	}

	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), target) {
			return true
		}
	}

	return false
}

func impossibleFilter() bson.M {
	return bson.M{
		"_id": bson.M{
			"$exists": false,
		},
	}
}