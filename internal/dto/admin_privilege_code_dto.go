package dto

import "time"

type PrivilegeGrantRequest struct {
	CategoryID   string `json:"categoryId,omitempty" example:"66e19b71c8f2a2b4d1234567"`
	CategorySlug string `json:"categorySlug,omitempty" example:"robbery"`
	CategoryName string `json:"categoryName,omitempty" example:"Robbery"`

	Actions []string `json:"actions" validate:"required,min=1" example:"reports:read,sos:read,chats:send"`

	AccessMode string `json:"accessMode,omitempty" example:"assigned_only" enums:"global,owned_only,assigned_only,scoped"`

	Countries []string `json:"countries,omitempty" example:"Ghana"`
	Regions   []string `json:"regions,omitempty" example:"Greater Accra"`
	Districts []string `json:"districts,omitempty" example:"Accra Metropolitan"`
}

type CreateAdminPrivilegeCodeRequest struct {
	Label   string `json:"label,omitempty" example:"Police Traffic Unit Access"`
	Purpose string `json:"purpose,omitempty" example:"Allow Police Traffic Unit to view reports, respond to SOS, and send chat replies"`

	OrganisationID   string `json:"organisationId,omitempty" example:"ghana_police_service"`
	OrganisationName string `json:"organisationName,omitempty" example:"Ghana Police Service"`
	OrganisationType string `json:"organisationType,omitempty" example:"police"`

	LevelID   string `json:"levelId,omitempty" example:"traffic_unit"`
	LevelName string `json:"levelName,omitempty" example:"Traffic Unit"`

	// Optional.
	// If omitted, backend derives permissions from grants.actions.
	Permissions []string `json:"permissions,omitempty" example:"reports:read,sos:read,chats:send"`

	// Recommended for organisation-specific access.
	Grants []PrivilegeGrantRequest `json:"grants,omitempty"`

	AccessMode string `json:"accessMode,omitempty" example:"assigned_only" enums:"global,owned_only,assigned_only,scoped"`

	ExpiresAt string `json:"expiresAt,omitempty" example:"2026-09-30T23:59:00Z"`
}

type UpdateAdminPrivilegeCodeRequest struct {
	Label   *string `json:"label,omitempty" example:"Police Robbery Access Updated"`
	Purpose *string `json:"purpose,omitempty" example:"Updated purpose for Police robbery access"`

	OrganisationID   *string `json:"organisationId,omitempty" example:"ghana_police_service"`
	OrganisationName *string `json:"organisationName,omitempty" example:"Ghana Police Service"`
	OrganisationType *string `json:"organisationType,omitempty" example:"police"`

	LevelID   *string `json:"levelId,omitempty" example:"traffic_unit"`
	LevelName *string `json:"levelName,omitempty" example:"Traffic Unit"`

	// Optional.
	// If supplied, this replaces the existing flat permissions.
	// If grants are also supplied, backend merges permissions with grants.actions.
	Permissions []string `json:"permissions,omitempty" example:"reports:read,sos:read,chats:send"`

	// Optional.
	// If supplied, this replaces the existing grants.
	Grants []PrivilegeGrantRequest `json:"grants,omitempty"`

	AccessMode *string `json:"accessMode,omitempty" example:"assigned_only" enums:"global,owned_only,assigned_only,scoped"`

	// Optional.
	// Omit to keep existing expiry.
	// Send empty string "" to clear expiry.
	// Send RFC3339 datetime to set expiry.
	ExpiresAt *string `json:"expiresAt,omitempty" example:"2026-09-30T23:59:00Z"`
}

type ValidateAdminPrivilegeCodeRequest struct {
	UUID string `json:"uuid" validate:"required" example:"4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"`
}

type RevokeAdminPrivilegeCodeRequest struct {
	Reason string `json:"reason,omitempty" example:"Access no longer needed"`
}

type PrivilegeGrantResponse struct {
	CategoryID   string `json:"categoryId,omitempty"`
	CategorySlug string `json:"categorySlug,omitempty"`
	CategoryName string `json:"categoryName,omitempty"`

	Actions    []string `json:"actions"`
	AccessMode string   `json:"accessMode"`

	Countries []string `json:"countries,omitempty"`
	Regions   []string `json:"regions,omitempty"`
	Districts []string `json:"districts,omitempty"`
}

type AdminPrivilegeCodeResponse struct {
	ID string `json:"id" example:"66e19b71c8f2a2b4d1234567"`

	UUID string `json:"uuid,omitempty" example:"4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"`

	CodePrefix string `json:"codePrefix" example:"4e1b5a0a"`

	Label   string `json:"label,omitempty" example:"Police Traffic Unit Access"`
	Purpose string `json:"purpose,omitempty" example:"Allow Police Traffic Unit to view SOS and update SOS status"`

	OrganisationID   string `json:"organisationId,omitempty" example:"ghana_police_service"`
	OrganisationName string `json:"organisationName,omitempty" example:"Ghana Police Service"`
	OrganisationType string `json:"organisationType,omitempty" example:"police"`

	LevelID   string `json:"levelId,omitempty" example:"traffic_unit"`
	LevelName string `json:"levelName,omitempty" example:"Traffic Unit"`

	Permissions []string `json:"permissions" example:"reports:read,reports:approve,alerts:read"`

	Grants []PrivilegeGrantResponse `json:"grants,omitempty"`

	AccessMode string `json:"accessMode" example:"assigned_only"`

	Status string `json:"status" example:"active" enums:"active,revoked,expired"`

	UsageCount int `json:"usageCount" example:"0"`

	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`

	CreatedBy string `json:"createdBy,omitempty" example:"admin_api_key"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
