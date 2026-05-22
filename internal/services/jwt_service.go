package services

import (
	"errors"
	"time"

	"disaster_alert_backend/config"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JWTService struct {
	secretKey string
	expiresIn time.Duration
}

type JWTClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func NewJWTService(cfg *config.Config) *JWTService {
	duration, err := time.ParseDuration(cfg.JWTExpiresIn)

	if err != nil {
		duration = 24 * time.Hour
	}

	return &JWTService{
		secretKey: cfg.JWTSecret,
		expiresIn: duration,
	}
}

func (s *JWTService) GenerateToken(userID primitive.ObjectID, email string, role string) (string, error) {
	claims := JWTClaims{
		UserID: userID.Hex(),
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.expiresIn)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.secretKey))
}

func (s *JWTService) ValidateToken(tokenString string) (*JWTClaims, error){
	token, err := jwt.ParseWithClaims(
		tokenString,
		&JWTClaims{},
		func(t *jwt.Token) (interface{}, error) {
			return []byte(s.secretKey), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)

	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}