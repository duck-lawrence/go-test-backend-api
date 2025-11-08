package externalservice

import (
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/jwt"
	"github.com/google/uuid"
)

type jwtService struct{}

func NewJwtService() externalservice.JwtService {
	return &jwtService{}
}

// GenerateAcAndRtTokens implements JwtService.
func (*jwtService) GenerateAcAndRtTokens(cfg *config.JWT, userID uuid.UUID) (string, string, error) {
	return jwt.GenerateAcAndRtTokens(cfg, userID)
}

// ValidateToken implements JwtService.
func (*jwtService) ValidateToken(secret []byte, tokenString string, purpose jwtpurpose.JWTPurpose) (*externalservice.CustomClaims, error) {
	return jwt.ValidateToken(secret, tokenString, purpose)
}

// GenerateEmailToken implements JwtService.
func (*jwtService) GenerateEmailToken(secret []byte, expiresIn time.Duration, email string, purpose jwtpurpose.JWTPurpose) (string, error) {
	return jwt.GenerateEmailToken(secret, expiresIn, email, purpose)
}
