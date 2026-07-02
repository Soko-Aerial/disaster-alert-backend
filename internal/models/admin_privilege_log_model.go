package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	PrivilegeLogCodeCreated     = "PRIVILEGE_CODE_CREATED"
	PrivilegeLogCodeValidated   = "PRIVILEGE_CODE_VALIDATED"
	PrivilegeLogCodeRevoked     = "PRIVILEGE_CODE_REVOKED"
	PrivilegeLogPermissionCheck = "PERMISSION_CHECK"
)

type AdminPrivilegeLog struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	PrivilegeCodeID string `bson:"privilegeCodeId,omitempty" json:"privilegeCodeId,omitempty"`
	CodePrefix      string `bson:"codePrefix,omitempty" json:"codePrefix,omitempty"`

	OrganisationID   string `bson:"organisationId,omitempty" json:"organisationId,omitempty"`
	OrganisationName string `bson:"organisationName,omitempty" json:"organisationName,omitempty"`

	LevelID   string `bson:"levelId,omitempty" json:"levelId,omitempty"`
	LevelName string `bson:"levelName,omitempty" json:"levelName,omitempty"`

	Action string `bson:"action" json:"action"`

	Endpoint string `bson:"endpoint,omitempty" json:"endpoint,omitempty"`
	Method   string `bson:"method,omitempty" json:"method,omitempty"`

	RequiredPermission string `bson:"requiredPermission,omitempty" json:"requiredPermission,omitempty"`
	Allowed            bool   `bson:"allowed" json:"allowed"`

	IPAddress string `bson:"ipAddress,omitempty" json:"ipAddress,omitempty"`
	UserAgent string `bson:"userAgent,omitempty" json:"userAgent,omitempty"`

	Message string `bson:"message,omitempty" json:"message,omitempty"`

	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}