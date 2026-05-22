package services

import (
	"errors"
	"strings"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserProfileDetailsService struct {
	userRepo *repositories.UserRepository
}

func NewUserProfileDetailsService(
	userRepo *repositories.UserRepository,
) *UserProfileDetailsService {
	return &UserProfileDetailsService{
		userRepo: userRepo,
	}
}

func (s *UserProfileDetailsService) GetProfileDetails(
	userID primitive.ObjectID,
) (*models.User, error) {
	return s.userRepo.FindProfileDetailsByUserID(userID)
}

func (s *UserProfileDetailsService) UpdateProfileDetails(
	userID primitive.ObjectID,
	req dto.UpdateUserProfileDetailsRequest,
) (*models.User, error) {
	update := bson.M{}

	if strings.TrimSpace(req.Phone) != "" {
		update["phone"] = strings.TrimSpace(req.Phone)
	}

	if req.MedicalInfo != nil {
		update["medicalInfo"] = models.UserMedicalInfo{
			BloodType:         strings.TrimSpace(req.MedicalInfo.BloodType),
			MedicalCondition: strings.TrimSpace(req.MedicalInfo.MedicalCondition),
			Medications:      cleanStringSlice(req.MedicalInfo.Medications),
			Allergies:        cleanStringSlice(req.MedicalInfo.Allergies),
			Disabilities:     strings.TrimSpace(req.MedicalInfo.Disabilities),
			Notes:            strings.TrimSpace(req.MedicalInfo.Notes),
		}
	}

	if req.Insurance != nil {
		update["insurance"] = models.UserInsuranceInfo{
			Provider:     strings.TrimSpace(req.Insurance.Provider),
			PolicyNumber: strings.TrimSpace(req.Insurance.PolicyNumber),
			Phone:        strings.TrimSpace(req.Insurance.Phone),
		}
	}

	if req.Workplace != nil {
		update["workplace"] = models.UserWorkplaceInfo{
			Name:    strings.TrimSpace(req.Workplace.Name),
			Role:    strings.TrimSpace(req.Workplace.Role),
			Address: strings.TrimSpace(req.Workplace.Address),
			Phone:   strings.TrimSpace(req.Workplace.Phone),
		}
	}

	if len(update) == 0 {
		return nil, errors.New("no valid profile details provided")
	}

	return s.userRepo.UpdateProfileDetailsByUserID(userID, update)
}

func cleanStringSlice(values []string) []string {
	cleaned := make([]string, 0)

	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		cleaned = append(cleaned, value)
	}

	return cleaned
}