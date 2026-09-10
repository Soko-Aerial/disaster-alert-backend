package dto

// UpdateUserLocationRequest updates the user's saved location.
//
// This can be used for default location, GPS location, or manually selected location.
type UpdateUserLocationRequest struct {
	// Name is a short name for the location.
	Name string `json:"name" example:"Home"`

	// Country is the user's country.
	Country string `json:"country" example:"Ghana"`

	// Region is the user's region/state/province.
	Region string `json:"region" example:"Greater Accra"`

	// Address is the readable address, landmark, or area.
	Address string `json:"address" example:"Circle, Accra"`

	// Latitude is the GPS latitude.
	Latitude float64 `json:"latitude" example:"5.6037"`

	// Longitude is the GPS longitude.
	Longitude float64 `json:"longitude" example:"-0.1870"`

	// Source describes where the location came from.
	//
	// Common values:
	// gps, manual
	Source string `json:"source" example:"gps" enums:"gps,manual"`

	// IsDefault marks this as the user's default saved location.
	IsDefault bool `json:"isDefault" example:"true"`
}

// UpdateUserMedicalInfoRequest updates emergency medical information.
type UpdateUserMedicalInfoRequest struct {
	// BloodType is the user's blood group.
	BloodType string `json:"bloodType" example:"O+"`

	// MedicalCondition contains known medical conditions.
	MedicalCondition string `json:"medicalCondition" example:"Asthma"`

	// Medications contains medicines the user currently takes.
	Medications []string `json:"medications" example:"Salbutamol inhaler"`

	// Allergies contains known allergies.
	Allergies []string `json:"allergies" example:"Penicillin"`

	// Disabilities contains disability or mobility information.
	Disabilities string `json:"disabilities" example:"None"`

	// Notes contains extra emergency medical notes.
	Notes string `json:"notes" example:"Uses an inhaler during asthma attacks."`
}

// UpdateUserInsuranceInfoRequest updates insurance details.
type UpdateUserInsuranceInfoRequest struct {
	// Provider is the insurance provider name.
	Provider string `json:"provider" example:"NHIS"`

	// PolicyNumber is the insurance policy or membership number.
	PolicyNumber string `json:"policyNumber" example:"NHIS-123456789"`

	// Phone is the insurance provider contact phone.
	Phone string `json:"phone" example:"+233302123456"`
}

// UpdateUserWorkplaceInfoRequest updates workplace or school information.
type UpdateUserWorkplaceInfoRequest struct {
	// Name is the workplace, school, or organization name.
	Name string `json:"name" example:"SokoAerial Robotics"`

	// Role is the user's role or position.
	Role string `json:"role" example:"Operations Officer"`

	// Address is the workplace address.
	Address string `json:"address" example:"Accra, Ghana"`

	// Phone is the workplace contact number.
	Phone string `json:"phone" example:"+233302123456"`
}

// UpdateUserProfileDetailsRequest updates extended user profile details.
//
// All fields are optional. Send only what you want to update.
type UpdateUserProfileDetailsRequest struct {
	// Name is the user's full name.
	Name string `json:"name" example:"Kwame Mensah"`

	// Email is the user's email address.
	Email string `json:"email" example:"kwame@example.com"`

	// Phone is the user's phone number.
	Phone string `json:"phone" example:"+233241234567"`

	// Gender is the user's gender.
	Gender string `json:"gender" example:"male" enums:"male,female,other,prefer_not_to_say"`

	// DateOfBirth is the user's date of birth.
	//
	// Recommended format:
	// YYYY-MM-DD
	DateOfBirth string `json:"dateOfBirth" example:"1995-05-12"`

	// Location contains the user's saved/default location.
	Location *UpdateUserLocationRequest `json:"location"`

	// MedicalInfo contains emergency medical information.
	MedicalInfo *UpdateUserMedicalInfoRequest `json:"medicalInfo"`

	// Insurance contains insurance information.
	Insurance *UpdateUserInsuranceInfoRequest `json:"insurance"`

	// Workplace contains workplace or school information.
	Workplace *UpdateUserWorkplaceInfoRequest `json:"workplace"`
}
