package mock

import (
	"context"

	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- Mock RefreshTokenRepo ---
type MockRefreshTokenRepo struct{ mock.Mock }

// DeleteByUserID implements repository.RefreshTokenRepository.
func (m *MockRefreshTokenRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}

// GetByTokenAndUserID implements repository.RefreshTokenRepository.
func (m *MockRefreshTokenRepo) GetByTokenAndUserID(ctx context.Context, token string, userID uuid.UUID) (*entities.RefreshToken, error) {
	panic("unimplemented")
}

// Revoke implements repository.RefreshTokenRepository.
func (m *MockRefreshTokenRepo) Revoke(ctx context.Context, token string, userID uuid.UUID) error {
	panic("unimplemented")
}

func (m *MockRefreshTokenRepo) Create(ctx context.Context, rt *entities.RefreshToken) error {
	return m.Called(ctx, rt).Error(0)
}
