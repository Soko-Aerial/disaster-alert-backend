package services

import (
	"errors"
	"time"

	"disaster_alert_backend/internal/dto"
	"disaster_alert_backend/internal/models"
	"disaster_alert_backend/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthService struct {
	userRepo        *repositories.UserRepository
	passwordService *passwordService
	jwtService      *JWTService
}

func NewAuthService(
	userRepo *repositories.UserRepository,
	passwordService *passwordService,
	jwtService *JWTService,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existingUser, err := s.userRepo.FindUserByEmail(req.Email)

	if err == nil && existingUser != nil {
		return nil, errors.New("email already exists")
	}

	if err != nil && err != mongo.ErrNoDocuments {
		return nil, err
	}

	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	user := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Phone:        req.Phone,
		Role:         "user",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	createdUser, err := s.userRepo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	token, err := s.jwtService.GenerateToken(
		createdUser.ID,
		createdUser.Email,
		createdUser.Role,
	)

	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  createdUser,
	}, nil
}

func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindUserByEmail(req.Email)

	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	passwordIsValid := s.passwordService.CheckPassword(
		req.Password,
		user.PasswordHash,
	)

	if !passwordIsValid {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.jwtService.GenerateToken(
		user.ID,
		user.Email,
		user.Role,
	)

	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GetCurrentUser(userID string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)

	if err != nil {
		return nil, errors.New("Invalid user ID")
	}

	user, err := s.userRepo.FindUserByID(objectID)

	if err != nil {
		return nil, errors.New("User not found")
	}

	return user, nil
}