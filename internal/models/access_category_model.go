package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AccessCategoryKindPublic = "public"
	AccessCategoryKindAccess = "access"
	AccessCategoryKindBoth   = "both"

	AccessCategoryVisibilitySystem  = "system"
	AccessCategoryVisibilityPrivate = "private"
	AccessCategoryVisibilityShared  = "shared"
)

type AccessCategory struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	Name        string `bson:"name" json:"name"`
	Slug        string `bson:"slug" json:"slug"`
	Description string `bson:"description,omitempty" json:"description,omitempty"`

	// public, access, both
	Kind string `bson:"kind" json:"kind"`

	// system, private, shared
	Visibility string `bson:"visibility" json:"visibility"`

	OwnerOrganisationID   string `bson:"ownerOrganisationId,omitempty" json:"ownerOrganisationId,omitempty"`
	OwnerOrganisationName string `bson:"ownerOrganisationName,omitempty" json:"ownerOrganisationName,omitempty"`

	// Connects existing AlertPreference fields to system categories.
	// Example: fire, flood, robbery, criticalAlerts, unverifiedActivity
	PreferenceKey string `bson:"preferenceKey,omitempty" json:"preferenceKey,omitempty"`

	// AllowedActions is the list of permissions/actions that can be selected under this category.
	AllowedActions []string `bson:"allowedActions,omitempty" json:"allowedActions,omitempty"`

	IsSystem bool `bson:"isSystem" json:"isSystem"`
	IsActive bool `bson:"isActive" json:"isActive"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
