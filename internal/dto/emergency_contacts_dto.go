package dto

type CreateEmergencyContactRequest struct {
	Name         string `json:"name" binding:"required"`
	Phone        string `json:"phone" binding:"required"`
	Email        string `json:"email"`
	Relationship string `json:"relationship"`

	Type         string `json:"type"`
	Organization string `json:"organization"`
	Address      string `json:"address"`

	IsPrimary    bool `json:"isPrimary"`
	IsGovernment bool `json:"isGovernment"`
}

type UpdateEmergencyContactRequest struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	Email        string `json:"email"`
	Relationship string `json:"relationship"`

	Type         string `json:"type"`
	Organization string `json:"organization"`
	Address      string `json:"address"`

	IsPrimary    *bool `json:"isPrimary"`
	IsGovernment *bool `json:"isGovernment"`
	IsActive     *bool `json:"isActive"`
}