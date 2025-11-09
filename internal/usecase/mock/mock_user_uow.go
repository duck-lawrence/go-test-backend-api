package mock

import (
	"context"

	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/uow"
	"github.com/stretchr/testify/mock"
)

// --- Mock UoW ---
type MockUserManagerUow struct{ mock.Mock }

// Do implements uow.UserManagerUow.
func (m *MockUserManagerUow) Do(ctx context.Context, fn func(r uow.UserManagerRepoProvider) error) error {
	panic("unimplemented")
}
