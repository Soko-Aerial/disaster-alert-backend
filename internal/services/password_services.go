package services

import(
	"golang.org/x/crypto/bcrypt"
)

type passwordService struct{}

func NewPasswordService() *passwordService {
	return &passwordService{}
}

func (s *passwordService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(password), 
		bcrypt.DefaultCost,
	)

	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (s *passwordService) CheckPassword(password string, hashPassword string) bool{
	err := bcrypt.CompareHashAndPassword(
		[]byte(hashPassword),
		[]byte(password),
	)

	
	return err == nil
}