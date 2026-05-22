package dto

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
	Phone       string                          `json:"phone"`
	MedicalInfo *UpdateUserMedicalInfoRequest  `json:"medicalInfo"`
	Insurance   *UpdateUserInsuranceInfoRequest `json:"insurance"`
	Workplace   *UpdateUserWorkplaceInfoRequest `json:"workplace"`
}