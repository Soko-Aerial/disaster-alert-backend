package dto

import "time"

type CreateAdminPrivilegeCodeRequest struct {
	Label string `json:"label" validate:"required" example:"Police Traffic Unit Access"`

	Purpose string `json:"purpose,omitempty" example:"Allow Police Traffic Unit to view SOS and update SOS status"`

	OrganisationID string `json:"organisationId" validate:"required" example:"firebase_police_org_id"`

	OrganisationName string `json:"organisationName" validate:"required" example:"Ghana Police Service"`

	LevelID string `json:"levelId,omitempty" example:"firebase_traffic_unit_id"`

	LevelName string `json:"levelName,omitempty" example:"Traffic Unit"`

	Permissions []string `json:"permissions" validate:"required,min=1" example:"sos:read"`

	ExpiresAt string `json:"expiresAt,omitempty" example:"2026-07-28T12:00:00Z"`
}

type ValidateAdminPrivilegeCodeRequest struct {
	UUID string `json:"uuid" validate:"required" example:"4e1b5a0a-71d7-40ad-9f30-9f1c4cbb1d9e"`
}

type RevokeAdminPrivilegeCodeRequest struct {
	Reason string `json:"reason,omitempty" example:"Access no longer needed"`
}

type AdminPrivilegeCodeResponse struct {
	ID string `json:"id"`

	// Returned only when the code is first generated.
	UUID string `json:"uuid,omitempty"`

	CodePrefix string `json:"codePrefix"`

	Label   string `json:"label"`
	Purpose string `json:"purpose,omitempty"`

	OrganisationID   string `json:"organisationId"`
	OrganisationName string `json:"organisationName"`

	LevelID   string `json:"levelId,omitempty"`
	LevelName string `json:"levelName,omitempty"`

	Permissions []string `json:"permissions"`

	Status string `json:"status"`

	UsageCount int        `json:"usageCount"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`

	ExpiresAt *time.Time `json:"expiresAt,omitempty"`

	CreatedBy string `json:"createdBy,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}