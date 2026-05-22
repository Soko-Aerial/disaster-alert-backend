package models

import(
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserLocation struct {
	Name      string  `bson:"name,omitempty" json:"name,omitempty"`
	Country   string  `bson:"country,omitempty" json:"country,omitempty"`
	Region    string  `bson:"region,omitempty" json:"region,omitempty"`
	Address   string  `bson:"address,omitempty" json:"address,omitempty"`
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
	Source    string  `bson:"source,omitempty" json:"source,omitempty"` // gps, manual
	IsDefault bool    `bson:"isDefault" json:"isDefault"`
}

type UserMedicalInfo struct {
	BloodType         string   `bson:"bloodType,omitempty" json:"bloodType,omitempty"`
	MedicalCondition string   `bson:"medicalCondition,omitempty" json:"medicalCondition,omitempty"`
	Medications      []string `bson:"medications,omitempty" json:"medications,omitempty"`
	Allergies        []string `bson:"allergies,omitempty" json:"allergies,omitempty"`
	Disabilities     string   `bson:"disabilities,omitempty" json:"disabilities,omitempty"`
	Notes            string   `bson:"notes,omitempty" json:"notes,omitempty"`
}

type UserInsuranceInfo struct {
	Provider     string `bson:"provider,omitempty" json:"provider,omitempty"`
	PolicyNumber string `bson:"policyNumber,omitempty" json:"policyNumber,omitempty"`
	Phone        string `bson:"phone,omitempty" json:"phone,omitempty"`
}

type UserWorkplaceInfo struct {
	Name    string `bson:"name,omitempty" json:"name,omitempty"`
	Role    string `bson:"role,omitempty" json:"role,omitempty"`
	Address string `bson:"address,omitempty" json:"address,omitempty"`
	Phone   string `bson:"phone,omitempty" json:"phone,omitempty"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Phone        string             `bson:"phone,omitempty" json:"phone,omitempty"`
	Role         string             `bson:"role" json:"role"`
	Location     *UserLocation      `bson:"location,omitempty" json:"location,omitempty"`
	MedicalInfo *UserMedicalInfo    `bson:"medicalInfo,omitempty" json:"medicalInfo,omitempty"`
	Insurance   *UserInsuranceInfo  `bson:"insurance,omitempty" json:"insurance,omitempty"`
	Workplace   *UserWorkplaceInfo  `bson:"workplace,omitempty" json:"workplace,omitempty"`
	FCMTokens    []string           `bson:"fcmTokens,omitempty" json:"fcmTokens,omitempty"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
	
}