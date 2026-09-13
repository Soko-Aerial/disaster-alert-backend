package authz

import (
	"strings"

	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/permissions"

	"github.com/gin-gonic/gin"
)

const PrivilegeContextKey = "privilegeContext"

type PrivilegeContext struct {
	PrivilegeCodeID string

	OrganisationID   string
	OrganisationName string
	OrganisationType string

	LevelID   string
	LevelName string

	Permissions []string
	Grants      []models.PrivilegeGrant
	AccessMode  string
}

func NewPrivilegeContext(code *models.AdminPrivilegeCode) *PrivilegeContext {
	if code == nil {
		return nil
	}

	return &PrivilegeContext{
		PrivilegeCodeID:  code.ID.Hex(),
		OrganisationID:   code.OrganisationID,
		OrganisationName: code.OrganisationName,
		OrganisationType: code.OrganisationType,
		LevelID:          code.LevelID,
		LevelName:        code.LevelName,
		Permissions:      code.Permissions,
		Grants:           code.Grants,
		AccessMode:       code.AccessMode,
	}
}

func FromGin(c *gin.Context) (*PrivilegeContext, bool) {
	value, exists := c.Get(PrivilegeContextKey)
	if !exists {
		return nil, false
	}

	ctx, ok := value.(*PrivilegeContext)
	if !ok || ctx == nil {
		return nil, false
	}

	return ctx, true
}

func IsGlobal(ctx *PrivilegeContext) bool {
	if ctx == nil {
		return false
	}

	return strings.EqualFold(ctx.AccessMode, models.PrivilegeAccessModeGlobal)
}

func HasPermission(ctx *PrivilegeContext, required string) bool {
	if ctx == nil {
		return false
	}

	return permissions.ContainsPermission(ctx.Permissions, required)
}

func HasGrantAction(ctx *PrivilegeContext, categorySlug string, requiredAction string) bool {
	if ctx == nil {
		return false
	}

	if IsGlobal(ctx) {
		return HasPermission(ctx, requiredAction)
	}

	categorySlug = strings.ToLower(strings.TrimSpace(categorySlug))

	for _, grant := range ctx.Grants {
		if !permissions.ContainsPermission(grant.Actions, requiredAction) {
			continue
		}

		grantSlug := strings.ToLower(strings.TrimSpace(grant.CategorySlug))

		if grantSlug == "" && categorySlug == "" {
			return true
		}

		if grantSlug != "" && grantSlug == categorySlug {
			return true
		}
	}

	return false
}
