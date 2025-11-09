package mock

import (
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- Mock JwtService ---
type MockJwtService struct{ mock.Mock }

// GenerateEmailToken implements externalservice.JwtService.
func (m *MockJwtService) GenerateEmailToken(secret []byte, expiresIn time.Duration, email string, purpose jwtpurpose.JWTPurpose) (string, error) {
	panic("unimplemented")
}

func (m *MockJwtService) GenerateAcAndRtTokens(cfg *config.JWT, userID uuid.UUID) (string, string, error) {
	args := m.Called(cfg, userID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJwtService) ValidateToken(secret []byte, token string, purpose jwtpurpose.JWTPurpose) (*externalservice.CustomClaims, error) {
	args := m.Called(secret, token, purpose)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*externalservice.CustomClaims), args.Error(1)
}
