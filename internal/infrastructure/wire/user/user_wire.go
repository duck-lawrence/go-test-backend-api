//go:build wireinject

package user

import (
	"github.com/ducklawrence05/go-test-backend-api/config"
	externalServiceImpl "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/externalservice"
	"github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/repository/postgres"
	rdRepo "github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/repository/redis"
	userInterface "github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	userImpl "github.com/ducklawrence05/go-test-backend-api/internal/usecase/user/implement"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"
	"github.com/redis/go-redis/v9"

	"github.com/google/wire"
	"gorm.io/gorm"
)

func NewUserAuthManager(
	config *config.Config,
	db *gorm.DB,
	// jwtService externalServiceInterface.JwtService,
	// passwordService externalServiceInterface.PasswordService,
) userInterface.UserAuthManager {
	wire.Build(
		externalServiceImpl.NewJwtService,
		externalServiceImpl.NewPasswordService,
		postgres.NewUserRepo,
		postgres.NewRefreshTokenRepo,
		postgres.NewUserManagerUow,
		userImpl.NewUserAuthManager,
	)
	return nil
}

func NewUserRegistrationManager(
	config *config.Config,
	db *gorm.DB,
	rdb *redis.Client,
	l logger.Interface,
) userInterface.UserRegistrationManager {
	wire.Build(
		rdRepo.NewOtpRepo,
		postgres.NewUserRepo,
		postgres.NewRefreshTokenRepo,
		postgres.NewUserManagerUow,
		userImpl.NewUserRegistrationManager,
	)
	return nil
}

func NewUserProfileManager(
	config *config.Config,
	db *gorm.DB,
) userInterface.UserProfileManager {
	wire.Build(
		postgres.NewUserRepo,
		postgres.NewUserManagerUow,
		userImpl.NewUserProfileManager,
	)
	return nil
}

func NewUserRestoreManager(
	config *config.Config,
	db *gorm.DB,
	rdb *redis.Client,
	l logger.Interface,
) userInterface.UserRestoreManager {
	wire.Build(
		rdRepo.NewOtpRepo,
		postgres.NewUserRepo,
		postgres.NewRefreshTokenRepo,
		postgres.NewUserManagerUow,
		userImpl.NewUserRestoreManager,
	)
	return nil
}
