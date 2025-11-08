package externalservice

import (
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/password"
)

type passwordService struct{}

func NewPasswordService() externalservice.PasswordService {
	return &passwordService{}
}

// ComparePasswords implements externalservice.PasswordService.
func (p *passwordService) ComparePasswords(hashedPassword string, plainPassword []byte) bool {
	return password.ComparePasswords(hashedPassword, plainPassword)
}

// HashPassword implements externalservice.PasswordService.
func (p *passwordService) HashPassword(rawPassword string) (string, error) {
	return password.HashPassword(rawPassword)
}
