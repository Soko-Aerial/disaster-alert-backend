package dto

type CreateAccessCategoryRequest struct {
	Name        string `json:"name" validate:"required" example:"Robbery"`
	Slug        string `json:"slug,omitempty" example:"robbery"`
	Description string `json:"description,omitempty" example:"Robbery, theft, armed robbery, and related security cases"`

	// public, access, both
	Kind string `json:"kind,omitempty" example:"both" enums:"public,access,both"`

	// private, shared
	Visibility string `json:"visibility,omitempty" example:"private" enums:"private,shared"`

	OwnerOrganisationID   string `json:"ownerOrganisationId,omitempty" example:"ghana_police_service"`
	OwnerOrganisationName string `json:"ownerOrganisationName,omitempty" example:"Ghana Police Service"`

	PreferenceKey string `json:"preferenceKey,omitempty" example:"robbery"`

	AllowedActions []string `json:"allowedActions,omitempty" example:"reports:read,sos:read,chats:send"`
}

type UpdateAccessCategoryRequest struct {
	Name        string `json:"name,omitempty" example:"Robbery"`
	Slug        string `json:"slug,omitempty" example:"robbery"`
	Description string `json:"description,omitempty" example:"Robbery, theft, armed robbery, and related security cases"`

	Kind       string `json:"kind,omitempty" example:"both" enums:"public,access,both"`
	Visibility string `json:"visibility,omitempty" example:"private" enums:"private,shared"`

	OwnerOrganisationID   string `json:"ownerOrganisationId,omitempty" example:"ghana_police_service"`
	OwnerOrganisationName string `json:"ownerOrganisationName,omitempty" example:"Ghana Police Service"`

	PreferenceKey string `json:"preferenceKey,omitempty" example:"robbery"`

	AllowedActions []string `json:"allowedActions,omitempty" example:"reports:read,sos:read,chats:send"`

	IsActive *bool `json:"isActive,omitempty" example:"true"`
}
