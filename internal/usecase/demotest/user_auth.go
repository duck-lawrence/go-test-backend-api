package demotest

import (
	"context"
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/repository"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/uow"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/google/uuid"
)

// implement
type userAuthManager struct {
	config           *config.Config
	uow              uow.UserManagerUow
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jwtService       externalservice.JwtService
	passwordService  externalservice.PasswordService
}

func NewUserAuthManager(
	config *config.Config,
	uow uow.UserManagerUow,
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
	jwtService externalservice.JwtService,
	passwordService externalservice.PasswordService,
) user.UserAuthManager {
	return &userAuthManager{
		config:           config,
		uow:              uow,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
		passwordService:  passwordService,
	}
}

func (m *userAuthManager) Login(ctx context.Context, dto user.LoginUserDto) (string, string, error) {
	// get user from db
	user, err := m.userRepo.GetByUserNameOrEmail(ctx, dto.EmailOrUsername)
	if err != nil {
		return "", "", err
	}

	if !m.passwordService.ComparePasswords(user.Password, []byte(dto.Password)) {
		return "", "", errorcode.ErrInvalidPassword
	}

	// gene ac and rt
	accessToken, refreshToken, err := m.jwtService.GenerateAcAndRtTokens(&m.config.JWT, user.ID)
	if err != nil {
		return "", "", err
	}

	// decode rt to get exp and iat
	claims, err := m.jwtService.ValidateToken([]byte(m.config.JWT.RefreshTokenKey),
		refreshToken, jwtpurpose.Refresh)
	if err != nil {
		return "", "", err
	}

	// insert rt to into db
	err = m.refreshTokenRepo.Create(ctx, &entities.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     refreshToken,
		IssuedAt:  claims.IssuedAt.Time,
		ExpiresAt: claims.ExpiresAt.Time,
		CreatedAt: time.Now(),
		Revoked:   false,
	})
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (m *userAuthManager) Logout(ctx context.Context, dto user.LogoutUserDto) error {
	// _ = dto.RefreshToken == ""
	panic("unimplement")
}

func (m *userAuthManager) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	// _ = refreshToken == ""
	panic("unimplement")
}
