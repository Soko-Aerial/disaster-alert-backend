package dto

import "time"

// CreateAdminPrivilegeCodeRequest is used to generate a new admin privilege UUID.
//
// The generated UUID is used in the X-Privilege-Code header when calling
// protected admin endpoints such as reports, alerts, SOS, assistance, chats,
// and notifications.
type CreateAdminPrivilegeCodeRequest struct {
	// Label is the readable name of this privilege code.
	//
	// Optional.
	// Example: Police Traffic Unit Access
	Label string `json:"label,omitempty" example:"Police Traffic Unit Access"`

	// Purpose explains why this privilege code exists.
	//
	// Optional.
	Purpose string `json:"purpose,omitempty" example:"Allow Police Traffic Unit to view reports, respond to SOS, and send chat replies"`

	// OrganisationID is the external/internal organisation identifier.
	//
	// Optional.
	OrganisationID string `json:"organisationId,omitempty" example:"firebase_police_org_id"`

	// OrganisationName is the readable organisation name.
	//
	// Optional.
	OrganisationName string `json:"organisationName,omitempty" example:"Ghana Police Service"`

	// LevelID is the optional department, unit, role, or level identifier.
	//
	// Optional.
	LevelID string `json:"levelId,omitempty" example:"firebase_traffic_unit_id"`

	// LevelName is the readable department, unit, role, or level name.
	//
	// Optional.
	LevelName string `json:"levelName,omitempty" example:"Traffic Unit"`

	// Permissions is the list of allowed admin actions for this privilege code.
	//
	// Required.
	Permissions []string `json:"permissions" validate:"required,min=1" example:"dashboard:read,reports:read,sos:read"`

	// ExpiresAt is the optional expiry date/time for the privilege code.
	//
	// Recommended format:
	// 2026-09-30T23:59:00Z
	ExpiresAt string `json:"expiresAt,omitempty" example:"2026-09-30T23:59:00Z"`
}

// ValidateAdminPrivilegeCodeRequest is used to check if a privilege UUID is valid.
type ValidateAdminPrivilegeCodeRequest struct {
	// UUID is the full privilege code returned during creation.
	//
	// Do not use codePrefix here.
	UUID string `json:"uuid" validate:"required" example:"4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"`
}

// RevokeAdminPrivilegeCodeRequest is used to revoke a privilege code.
type RevokeAdminPrivilegeCodeRequest struct {
	// Reason explains why the privilege code is being revoked.
	Reason string `json:"reason,omitempty" example:"Access no longer needed"`
}

// AdminPrivilegeCodeResponse is returned when privilege code records are fetched.
//
// Important:
// The full UUID is only returned once during creation.
// List and detail endpoints return codePrefix only, not the full UUID.
type AdminPrivilegeCodeResponse struct {
	ID string `json:"id" example:"66e19b71c8f2a2b4d1234567"`

	// UUID is returned only when the code is first generated.
	UUID string `json:"uuid,omitempty" example:"4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"`

	CodePrefix string `json:"codePrefix" example:"4e1b5a0a"`

	Label string `json:"label" example:"Police Traffic Unit Access"`

	Purpose string `json:"purpose,omitempty" example:"Allow Police Traffic Unit to view SOS and update SOS status"`

	OrganisationID string `json:"organisationId" example:"firebase_police_org_id"`

	OrganisationName string `json:"organisationName" example:"Ghana Police Service"`

	LevelID string `json:"levelId,omitempty" example:"firebase_traffic_unit_id"`

	LevelName string `json:"levelName,omitempty" example:"Traffic Unit"`

	Permissions []string `json:"permissions" example:"reports:read,reports:approve,alerts:read"`

	Status string `json:"status" example:"active" enums:"active,revoked,expired"`

	UsageCount int `json:"usageCount" example:"0"`

	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`

	ExpiresAt *time.Time `json:"expiresAt,omitempty"`

	CreatedBy string `json:"createdBy,omitempty" example:"admin_api_key"`

	CreatedAt time.Time `json:"createdAt"`

	UpdatedAt time.Time `json:"updatedAt"`
}
