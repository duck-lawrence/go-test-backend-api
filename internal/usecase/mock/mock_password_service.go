package mock

import (
	"github.com/stretchr/testify/mock"
)

// --- Mock PasswordService ---
type MockPasswordService struct{ mock.Mock }

// HashPassword implements externalservice.PasswordService.
func (m *MockPasswordService) HashPassword(password string) (string, error) {
	panic("unimplemented")
}

func (m *MockPasswordService) ComparePasswords(hashed string, plain []byte) bool {
	return m.Called(hashed, plain).Bool(0)
}
