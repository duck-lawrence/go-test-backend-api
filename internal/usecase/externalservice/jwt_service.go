package externalservice

import (
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type CustomClaims struct {
	Purpose jwtpurpose.JWTPurpose `json:"purpose"`
	jwt.RegisteredClaims
}

type JwtService interface {
	GenerateAcAndRtTokens(cfg *config.JWT, userID uuid.UUID) (string, string, error)
	ValidateToken(secret []byte, tokenString string, purpose jwtpurpose.JWTPurpose) (*CustomClaims, error)
	GenerateEmailToken(secret []byte, expiresIn time.Duration, email string, purpose jwtpurpose.JWTPurpose) (string, error)
}
