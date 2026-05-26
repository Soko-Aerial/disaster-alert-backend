package dto

type UpdateUserLocationRequest struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Region    string  `json:"region"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Source    string  `json:"source"`
	IsDefault bool    `json:"isDefault"`
}

type UpdateUserMedicalInfoRequest struct {
	BloodType         string   `json:"bloodType"`
	MedicalCondition string   `json:"medicalCondition"`
	Medications      []string `json:"medications"`
	Allergies        []string `json:"allergies"`
	Disabilities     string   `json:"disabilities"`
	Notes            string   `json:"notes"`
}

type UpdateUserInsuranceInfoRequest struct {
	Provider     string `json:"provider"`
	PolicyNumber string `json:"policyNumber"`
	Phone        string `json:"phone"`
}

type UpdateUserWorkplaceInfoRequest struct {
	Name    string `json:"name"`
	Role    string `json:"role"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type UpdateUserProfileDetailsRequest struct {
	Name     string                     `json:"name"`
	Email    string                     `json:"email"`
	Phone    string                     `json:"phone"`
	Gender       string                     `json:"gender"`
	DateOfBirth  string                     `json:"dateOfBirth"`
	Location *UpdateUserLocationRequest `json:"location"`

	MedicalInfo *UpdateUserMedicalInfoRequest   `json:"medicalInfo"`
	Insurance   *UpdateUserInsuranceInfoRequest `json:"insurance"`
	Workplace   *UpdateUserWorkplaceInfoRequest `json:"workplace"`
}