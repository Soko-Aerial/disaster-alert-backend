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

// cleanStringSlice trims strings and removes empty entries
func cleanStringSlice(in []string) []string {
	if in == nil {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		t := strings.TrimSpace(s)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

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

	if strings.TrimSpace(req.Name) != "" {
		update["name"] = strings.TrimSpace(req.Name)
	}

	if strings.TrimSpace(req.Email) != "" {
		update["email"] = strings.TrimSpace(req.Email)
	}

	if strings.TrimSpace(req.Phone) != "" {
		update["phone"] = strings.TrimSpace(req.Phone)
	}

	if strings.TrimSpace(req.Gender) != "" {
		update["gender"] = strings.TrimSpace(req.Gender)
	}

	if strings.TrimSpace(req.DateOfBirth) != "" {
		update["dateOfBirth"] = strings.TrimSpace(req.DateOfBirth)
	}

	if req.Location != nil {
		update["location"] = models.UserLocation{
			Name:      strings.TrimSpace(req.Location.Name),
			Country:   strings.TrimSpace(req.Location.Country),
			Region:    strings.TrimSpace(req.Location.Region),
			Address:   strings.TrimSpace(req.Location.Address),
			Latitude:  req.Location.Latitude,
			Longitude: req.Location.Longitude,
			Source:    strings.TrimSpace(req.Location.Source),
			IsDefault: req.Location.IsDefault,
		}
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