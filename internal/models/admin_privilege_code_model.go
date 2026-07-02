package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PrivilegeCodeStatusActive  = "active"
	PrivilegeCodeStatusRevoked = "revoked"
	PrivilegeCodeStatusExpired = "expired"
)

type AdminPrivilegeCode struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// Raw UUID is not stored. The backend stores a hash for security.
	CodeHash   string `bson:"codeHash" json:"-"`
	CodePrefix string `bson:"codePrefix" json:"codePrefix"`

	Label   string `bson:"label" json:"label"`
	Purpose string `bson:"purpose,omitempty" json:"purpose,omitempty"`

	OrganisationID   string `bson:"organisationId" json:"organisationId"`
	OrganisationName string `bson:"organisationName" json:"organisationName"`

	LevelID   string `bson:"levelId,omitempty" json:"levelId,omitempty"`
	LevelName string `bson:"levelName,omitempty" json:"levelName,omitempty"`

	Permissions []string `bson:"permissions" json:"permissions"`

	Status string `bson:"status" json:"status"`

	UsageCount int        `bson:"usageCount" json:"usageCount"`
	LastUsedAt *time.Time `bson:"lastUsedAt,omitempty" json:"lastUsedAt,omitempty"`

	ExpiresAt *time.Time `bson:"expiresAt,omitempty" json:"expiresAt,omitempty"`

	CreatedBy string `bson:"createdBy,omitempty" json:"createdBy,omitempty"`

	RevokedAt     *time.Time `bson:"revokedAt,omitempty" json:"revokedAt,omitempty"`
	RevokedBy     string     `bson:"revokedBy,omitempty" json:"revokedBy,omitempty"`
	RevokeReason  string     `bson:"revokeReason,omitempty" json:"revokeReason,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}