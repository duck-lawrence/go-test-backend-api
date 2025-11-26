package implement

import (
	"context"
	"errors"
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/otptype"
	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/repository"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/uow"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/jwt"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/otputils"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/password"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/sendto"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/stringutils"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type userRestoreManager struct {
	config           *config.Config
	logger           logger.Interface
	uow              uow.UserManagerUow
	otpRepo          repository.OTPRepository
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
}

func NewUserRestoreManager(
	config *config.Config,
	logger logger.Interface,
	uow uow.UserManagerUow,
	otpRepo repository.OTPRepository,
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
) user.UserRestoreManager {
	return &userRestoreManager{
		config:           config,
		logger:           logger,
		uow:              uow,
		otpRepo:          otpRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

// SendRestoreOTP implements user.UserRestoreManager.
func (m *userRestoreManager) SendRestoreOTP(ctx context.Context, email string) error {
	// gene otp
	otp := otputils.GenerateSecureOTP()
	// hash email
	hashedEmail := stringutils.HashString(email, []byte(m.config.OTP.RestoreAccountKey))
	// save otp to redis
	if err := m.otpRepo.SetOTP(ctx, hashedEmail, otp, otptype.RestoreAccount,
		m.config.OTP.RestoreAccountTTL); err != nil {
		return err
	}

	// send otp to email
	go func() {
		err := sendto.SendTemplateEmailOtp(&m.config.SMTP, []string{email},
			"otp-restore-account.html", map[string]any{"otp": otp})
		if err != nil {
			m.logger.Error("Send email error", zap.Error(err))
		}
	}()
	return nil
}

// VerifyRestoreOTP implements user.UserRestoreManager.
func (m *userRestoreManager) VerifyRestoreOTP(ctx context.Context, email string, otp string) (string, error) {
	// hash email
	hashedEmail := stringutils.HashString(email, []byte(m.config.OTP.RestoreAccountKey))

	// check otp in redis
	storedOtp, err := m.otpRepo.GetOTP(ctx, hashedEmail, otptype.RestoreAccount)
	if err != nil {
		return "", err
	}
	if storedOtp != otp {
		return "", errorcode.ErrInvalidOTP
	}

	// delete otp in redis
	if err := m.otpRepo.DeleteOTP(ctx, hashedEmail, otptype.RestoreAccount); err != nil {
		m.logger.Warn("Cannot delete old otp from Redis", zap.Error(err))
	}

	// gene jwt token
	token, err := jwt.GenerateEmailToken([]byte(m.config.JWT.RestoreAccountTokenKey),
		m.config.JWT.RestoreAccountTokenExpiresIn,
		email, jwtpurpose.Restore)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Restore implements user.UserRestoreManager.
func (m *userRestoreManager) Restore(ctx context.Context, dto user.RestoreUserDto) (string, string, error) {
	g, gCtx := errgroup.WithContext(ctx)

	userChan := make(chan *entities.User, 1)
	hpChan := make(chan string, 1)

	// get user by email
	g.Go(func() error {
		user, err := m.userRepo.GetByUserNameOrEmail(gCtx, dto.Email)
		if err != nil && !errors.Is(err, errorcode.ErrDeletedAccount) {
			return err
		}
		userChan <- user
		return nil
	})

	// hash pass
	g.Go(func() error {
		hp, err := password.HashPassword(dto.NewPassword)
		if err != nil {
			return err
		}
		hpChan <- hp
		return nil
	})

	if err := g.Wait(); err != nil {
		return "", "", err
	}

	// update user password and deleted at field
	user := <-userChan
	user.Password = <-hpChan

	var accessToken, refreshToken string
	// begin transaction
	err := m.uow.Do(ctx, func(r uow.UserManagerRepoProvider) error {
		// update user in db
		err := r.UserRepository().Update(ctx, user, map[string]any{
			"password":   user.Password,
			"deleted_at": nil,
		})
		if err != nil {
			return err
		}

		// gene ac and rt
		accessToken, refreshToken, err = jwt.GenerateAcAndRtTokens(&m.config.JWT, user.ID)
		if err != nil {
			return err
		}

		// decode rt to get iat and exp
		claims, err := jwt.ValidateToken([]byte(m.config.JWT.RefreshTokenKey),
			refreshToken, jwtpurpose.Refresh)
		if err != nil {
			return err
		}

		// insert rt to db
		if err := r.RefreshTokenRepository().Create(ctx, &entities.RefreshToken{
			ID:        uuid.New(),
			UserID:    user.ID,
			Token:     refreshToken,
			IssuedAt:  claims.IssuedAt.Time,
			ExpiresAt: claims.ExpiresAt.Time,
			CreatedAt: time.Now(),
			Revoked:   false,
		}); err != nil {
			return err
		}

		// commit
		return nil
	})

	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}
