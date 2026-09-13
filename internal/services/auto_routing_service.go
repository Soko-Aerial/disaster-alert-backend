package services

import (
	"strings"

	"disaster_alert_backend/internal/authz"
)

const (
	SystemOrganisationID = "system"

	OrganisationNADMO              = "nadmo"
	OrganisationPolice             = "ghana_police_service"
	OrganisationFireService        = "ghana_fire_service"
	OrganisationHealthService      = "ghana_health_service"
	OrganisationAmbulanceService   = "national_ambulance_service"
	OrganisationArmedForces        = "ghana_armed_forces"
	OrganisationNationalSecurity   = "national_security"
	OrganisationMineralsCommission = "minerals_commission"
)

type AutoRouteResult struct {
	AccessCategoryID   string
	AccessCategorySlug string
	AccessCategoryName string

	OwnerOrganisationID string
	LeadOrganisationID  string
	AssignedOrgIDs      []string
	VisibleToOrgIDs     []string

	NeedsManualReview bool
	Reason            string
}

func ResolveAutoRoute(
	accessCategoryID string,
	accessCategorySlug string,
	accessCategoryName string,
	fallbackCategory string,
) AutoRouteResult {
	slug := strings.TrimSpace(accessCategorySlug)
	if slug == "" {
		slug = strings.TrimSpace(fallbackCategory)
	}

	slug = authz.NormalizeScopeSlug(slug)

	name := strings.TrimSpace(accessCategoryName)
	if name == "" {
		name = categoryNameFromSlug(slug)
	}

	route := AutoRouteResult{
		AccessCategoryID:    strings.TrimSpace(accessCategoryID),
		AccessCategorySlug:  slug,
		AccessCategoryName:  name,
		OwnerOrganisationID: SystemOrganisationID,
		LeadOrganisationID:  SystemOrganisationID,
		AssignedOrgIDs:      []string{},
		VisibleToOrgIDs:     []string{},
		NeedsManualReview:   false,
		Reason:              "auto routed by category",
	}

	switch slug {
	case "fire":
		route.LeadOrganisationID = OrganisationFireService
		route.AssignedOrgIDs = []string{
			OrganisationFireService,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationFireService,
			OrganisationNADMO,
		}

	case "flood", "weather", "drought", "earthquake":
		route.LeadOrganisationID = OrganisationNADMO
		route.AssignedOrgIDs = []string{
			OrganisationNADMO,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationNADMO,
			OrganisationFireService,
			OrganisationPolice,
		}

	case "health":
		route.LeadOrganisationID = OrganisationHealthService
		route.AssignedOrgIDs = []string{
			OrganisationHealthService,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationHealthService,
			OrganisationNADMO,
		}

	case "medical":
		route.LeadOrganisationID = OrganisationAmbulanceService
		route.AssignedOrgIDs = []string{
			OrganisationAmbulanceService,
			OrganisationHealthService,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationAmbulanceService,
			OrganisationHealthService,
			OrganisationNADMO,
		}

	case "security", "robbery", "protests":
		route.LeadOrganisationID = OrganisationPolice
		route.AssignedOrgIDs = []string{
			OrganisationPolice,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationPolice,
			OrganisationNADMO,
		}

	case "conflict", "munitions":
		route.LeadOrganisationID = OrganisationNationalSecurity
		route.AssignedOrgIDs = []string{
			OrganisationNationalSecurity,
			OrganisationPolice,
			OrganisationArmedForces,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationNationalSecurity,
			OrganisationPolice,
			OrganisationArmedForces,
			OrganisationNADMO,
		}

	case "galamsey":
		route.LeadOrganisationID = OrganisationPolice
		route.AssignedOrgIDs = []string{
			OrganisationPolice,
			OrganisationMineralsCommission,
			OrganisationNationalSecurity,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationPolice,
			OrganisationMineralsCommission,
			OrganisationNationalSecurity,
			OrganisationNADMO,
		}

	case "accident":
		route.LeadOrganisationID = OrganisationPolice
		route.AssignedOrgIDs = []string{
			OrganisationPolice,
			OrganisationAmbulanceService,
		}
		route.VisibleToOrgIDs = []string{
			OrganisationPolice,
			OrganisationAmbulanceService,
			OrganisationHealthService,
			OrganisationNADMO,
		}

	default:
		route.NeedsManualReview = true
		route.Reason = "category could not be auto routed"
		route.LeadOrganisationID = SystemOrganisationID
		route.AssignedOrgIDs = []string{}
		route.VisibleToOrgIDs = []string{}
	}

	return route
}

func ApplyAlertRoutingOverrides(
	route AutoRouteResult,
	ownerOrganisationID string,
	leadOrganisationID string,
	assignedOrgIDs []string,
	visibleToOrgIDs []string,
) AutoRouteResult {
	if strings.TrimSpace(ownerOrganisationID) != "" {
		route.OwnerOrganisationID = strings.TrimSpace(ownerOrganisationID)
	}

	if strings.TrimSpace(leadOrganisationID) != "" {
		route.LeadOrganisationID = strings.TrimSpace(leadOrganisationID)
	}

	cleanAssigned := cleanStringList(assignedOrgIDs)
	if len(cleanAssigned) > 0 {
		route.AssignedOrgIDs = cleanAssigned
	}

	cleanVisible := cleanStringList(visibleToOrgIDs)
	if len(cleanVisible) > 0 {
		route.VisibleToOrgIDs = cleanVisible
	}

	return route
}

func categoryNameFromSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "Other"
	}

	parts := strings.Split(slug, "_")

	for index, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		lowerPart := strings.ToLower(part)
		parts[index] = strings.ToUpper(lowerPart[:1]) + lowerPart[1:]
	}

	return strings.Join(parts, " ")
}
