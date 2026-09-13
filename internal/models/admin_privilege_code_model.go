package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PrivilegeCodeStatusActive  = "active"
	PrivilegeCodeStatusRevoked = "revoked"
	PrivilegeCodeStatusExpired = "expired"

	PrivilegeAccessModeGlobal       = "global"
	PrivilegeAccessModeOwnedOnly    = "owned_only"
	PrivilegeAccessModeAssignedOnly = "assigned_only"
	PrivilegeAccessModeScoped       = "scoped"
)

type PrivilegeGrant struct {
	CategoryID   string `bson:"categoryId,omitempty" json:"categoryId,omitempty"`
	CategorySlug string `bson:"categorySlug,omitempty" json:"categorySlug,omitempty"`
	CategoryName string `bson:"categoryName,omitempty" json:"categoryName,omitempty"`

	// Actions are existing permissions under this category.
	// Example: reports:read, sos:update_status, chats:send
	Actions []string `bson:"actions" json:"actions"`

	// global, owned_only, assigned_only, scoped
	AccessMode string `bson:"accessMode" json:"accessMode"`

	Countries []string `bson:"countries,omitempty" json:"countries,omitempty"`
	Regions   []string `bson:"regions,omitempty" json:"regions,omitempty"`
	Districts []string `bson:"districts,omitempty" json:"districts,omitempty"`
}

type AdminPrivilegeCode struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Raw UUID is not stored. The backend stores a hash for security.
	CodeHash   string `bson:"codeHash" json:"-"`
	CodePrefix string `bson:"codePrefix" json:"codePrefix"`

	Label   string `bson:"label,omitempty" json:"label,omitempty"`
	Purpose string `bson:"purpose,omitempty" json:"purpose,omitempty"`

	OrganisationID   string `bson:"organisationId,omitempty" json:"organisationId,omitempty"`
	OrganisationName string `bson:"organisationName,omitempty" json:"organisationName,omitempty"`
	OrganisationType string `bson:"organisationType,omitempty" json:"organisationType,omitempty"`

	LevelID   string `bson:"levelId,omitempty" json:"levelId,omitempty"`
	LevelName string `bson:"levelName,omitempty" json:"levelName,omitempty"`

	// Old flat permission list. Keep this so your existing middleware remains compatible.
	Permissions []string `bson:"permissions" json:"permissions"`

	// New category/action grants.
	Grants []PrivilegeGrant `bson:"grants,omitempty" json:"grants,omitempty"`

	// global, owned_only, assigned_only, scoped
	AccessMode string `bson:"accessMode" json:"accessMode"`

	Status     string `bson:"status" json:"status"`
	UsageCount int    `bson:"usageCount" json:"usageCount"`

	LastUsedAt *time.Time `bson:"lastUsedAt,omitempty" json:"lastUsedAt,omitempty"`
	ExpiresAt  *time.Time `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`

	CreatedBy string `bson:"createdBy,omitempty" json:"createdBy,omitempty"`

	RevokedAt    *time.Time `bson:"revokedAt,omitempty" json:"revokedAt,omitempty"`
	RevokedBy    string     `bson:"revokedBy,omitempty" json:"revokedBy,omitempty"`
	RevokeReason string     `bson:"revokeReason,omitempty" json:"revokeReason,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}
