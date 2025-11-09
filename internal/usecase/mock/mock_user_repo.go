package mock

import (
	"context"

	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// --- Mock UserRepo ---
type MockUserRepo struct{ mock.Mock }

// Create implements repository.UserRepository.
func (m *MockUserRepo) Create(ctx context.Context, user *entities.User) error {
	panic("unimplemented")
}

// DeleteByID implements repository.UserRepository.
func (m *MockUserRepo) DeleteByID(ctx context.Context, userID uuid.UUID) error {
	panic("unimplemented")
}

// GetByID implements repository.UserRepository.
func (m *MockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	panic("unimplemented")
}

// IsEmailTaken implements repository.UserRepository.
func (m *MockUserRepo) IsEmailTaken(ctx context.Context, email string, excludeUserID uuid.UUID) (bool, error) {
	panic("unimplemented")
}

// IsUserNameTaken implements repository.UserRepository.
func (m *MockUserRepo) IsUserNameTaken(ctx context.Context, userName string, excludeUserID uuid.UUID) (bool, error) {
	panic("unimplemented")
}

// Update implements repository.UserRepository.
func (m *MockUserRepo) Update(ctx context.Context, user *entities.User, fields map[string]any) error {
	panic("unimplemented")
}

func (m *MockUserRepo) GetByUserNameOrEmail(ctx context.Context, u string) (*entities.User, error) {
	args := m.Called(ctx, u)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.User), args.Error(1)
}
