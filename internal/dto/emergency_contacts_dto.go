package dto

// CreateEmergencyContactRequest is used by a mobile user to add a new emergency contact.
//
// Emergency contacts can be personal contacts, family members, medical contacts,
// responders, government contacts, or organizations that can be notified during emergencies.
type CreateEmergencyContactRequest struct {
	// Name is the contact person's or organization's name.
	Name string `json:"name" binding:"required" example:"Ama Mensah"`

	// Phone is the contact phone number.
	Phone string `json:"phone" binding:"required" example:"+233241234567"`

	// Email is the contact email address.
	Email string `json:"email" example:"ama@example.com"`

	// Relationship describes how this contact is related to the user.
	Relationship string `json:"relationship" example:"Mother"`

	// Type describes the contact category.
	//
	// Examples:
	// personal, family, medical, police, fire, ambulance, government, organization, other
	Type string `json:"type" example:"family" enums:"personal,family,medical,police,fire,ambulance,government,organization,other"`

	// Organization is the organization name if this contact belongs to an institution.
	Organization string `json:"organization" example:"National Ambulance Service"`

	// Address is the contact's address or office location.
	Address string `json:"address" example:"Accra, Ghana"`

	// IsPrimary marks this as the user's main emergency contact.
	IsPrimary bool `json:"isPrimary" example:"true"`

	// IsGovernment marks this contact as a government or official response contact.
	IsGovernment bool `json:"isGovernment" example:"false"`
}

// UpdateEmergencyContactRequest is used to update an existing emergency contact.
//
// All fields are optional. Send only the fields you want to update.
type UpdateEmergencyContactRequest struct {
	// Name is the contact person's or organization's name.
	Name string `json:"name" example:"Ama Mensah"`

	// Phone is the contact phone number.
	Phone string `json:"phone" example:"+233241234567"`

	// Email is the contact email address.
	Email string `json:"email" example:"ama@example.com"`

	// Relationship describes how this contact is related to the user.
	Relationship string `json:"relationship" example:"Mother"`

	// Type describes the contact category.
	Type string `json:"type" example:"family" enums:"personal,family,medical,police,fire,ambulance,government,organization,other"`

	// Organization is the organization name if this contact belongs to an institution.
	Organization string `json:"organization" example:"National Ambulance Service"`

	// Address is the contact's address or office location.
	Address string `json:"address" example:"Accra, Ghana"`

	// IsPrimary marks this as the user's main emergency contact.
	IsPrimary *bool `json:"isPrimary" example:"true"`

	// IsGovernment marks this contact as a government or official response contact.
	IsGovernment *bool `json:"isGovernment" example:"false"`

	// IsActive controls whether this contact is currently active.
	IsActive *bool `json:"isActive" example:"true"`
}