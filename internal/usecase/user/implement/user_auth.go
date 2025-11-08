package implement

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
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/jwt"
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

func (m *userAuthManager) Login(ctx context.Context, vo user.LoginUserVO) (string, string, error) {
	// get user from db
	user, err := m.userRepo.GetByUserNameOrEmail(ctx, vo.EmailOrUsername)
	if err != nil {
		return "", "", err
	}

	if !m.passwordService.ComparePasswords(user.Password, []byte(vo.Password)) {
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

func (m *userAuthManager) Logout(ctx context.Context, vo user.LogoutUserVO) error {
	// decode rt
	claims, err := jwt.ValidateToken([]byte(m.config.JWT.RefreshTokenKey),
		vo.RefreshToken, jwtpurpose.Refresh)
	if err != nil {
		return err
	}

	// compare userID from ac and rt
	if claims.Subject != vo.UserID.String() {
		return errorcode.ErrInvalidToken
	}

	// check if revoked or not
	if _, err := m.refreshTokenRepo.GetByTokenAndUserID(ctx, vo.RefreshToken, vo.UserID); err != nil {
		return errorcode.ErrInvalidToken
	}

	// revoke
	err = m.refreshTokenRepo.Revoke(ctx, vo.RefreshToken, vo.UserID)
	if err != nil {
		return err
	}

	return nil
}

func (m *userAuthManager) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	var accessToken, newRefreshToken string
	err := m.uow.Do(ctx, func(r uow.UserManagerRepoProvider) error {
		// validate token
		claims, err := jwt.ValidateToken([]byte(m.config.JWT.RefreshTokenKey),
			refreshToken, jwtpurpose.Refresh)
		if err != nil {
			return errorcode.ErrInvalidToken
		}

		// parse sub to uuid
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			return errorcode.ErrInvalidToken
		}

		// check token in db
		if _, err := r.RefreshTokenRepository().GetByTokenAndUserID(
			ctx, refreshToken, userID,
		); err != nil {
			return errorcode.ErrInvalidToken
		}

		// gene ac and rt
		accessToken, newRefreshToken, err = jwt.GenerateAcAndRtTokens(&m.config.JWT, userID)
		if err != nil {
			return err
		}

		// decode rt to get exp and iat
		newClaims, err := jwt.ValidateToken([]byte(m.config.JWT.RefreshTokenKey),
			newRefreshToken, jwtpurpose.Refresh)
		if err != nil {
			return err
		}

		// insert rt to into db
		err = r.RefreshTokenRepository().Create(ctx, &entities.RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			Token:     newRefreshToken,
			IssuedAt:  newClaims.IssuedAt.Time,
			ExpiresAt: newClaims.ExpiresAt.Time,
			CreatedAt: time.Now(),
			Revoked:   false,
		})
		if err != nil {
			return err
		}

		// revoke old rt
		err = r.RefreshTokenRepository().Revoke(ctx, refreshToken, userID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}
	return accessToken, newRefreshToken, nil
}
