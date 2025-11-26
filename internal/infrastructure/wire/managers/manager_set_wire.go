//go:build wireinject

package managers

import (
	"github.com/ducklawrence05/go-test-backend-api/config"
	otpWire "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/otp"
	roleWire "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/role"
	userWire "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func InitializeManagers(
	config *config.Config,
	db *gorm.DB,
	rdb *redis.Client,
	l logger.Interface,
) (*ManagerSet, error) {
	wire.Build(
		userWire.NewUserRegistrationManager,
		userWire.NewUserAuthManager,
		userWire.NewUserRestoreManager,
		userWire.NewUserProfileManager,
		roleWire.NewRoleManager,
		otpWire.NewOTPRateLimitManager,
		otpWire.NewOTPVerifyManager,
		wire.Struct(new(ManagerSet), "*"),
	)
	return nil, nil
}

func ProvideUserManagerSet(m *ManagerSet) *UserManagerSet {
	return &UserManagerSet{
		Registration: m.UserRegistration,
		Restore:      m.UserRestore,
		Auth:         m.UserAuth,
		Profile:      m.UserProfile,
		OTPRateLimit: m.OTPRateLimit,
		OTPVerify:    m.OTPVerify,
	}
}
