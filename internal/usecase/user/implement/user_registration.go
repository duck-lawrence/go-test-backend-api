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
	"github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/rolecache"
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

type userRegistrationManager struct {
	config           *config.Config
	logger           logger.Interface
	uow              uow.UserManagerUow
	otpRepo          repository.OTPRepository
	userRepo         repository.UserRepository
	refreshTokenRepo repository.RefreshTokenRepository
}

func NewUserRegistrationManager(
	config *config.Config,
	logger logger.Interface,
	uow uow.UserManagerUow,
	otpRepo repository.OTPRepository,
	userRepo repository.UserRepository,
	refreshTokenRepo repository.RefreshTokenRepository,
) user.UserRegistrationManager {
	return &userRegistrationManager{
		config:           config,
		uow:              uow,
		logger:           logger,
		otpRepo:          otpRepo,
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
	}
}

func (m *userRegistrationManager) SendRegistrationOTP(ctx context.Context, email string) error {
	// check email exists
	exists, err := m.userRepo.IsEmailTaken(ctx, email, uuid.Nil)
	if err != nil && !errors.Is(err, errorcode.ErrUserNotFound) {
		return err
	}
	if exists {
		return errorcode.ErrExistedEmail
	}

	// gene otp
	otp := otputils.GenerateSecureOTP()
	// hash email
	hashedEmail := stringutils.HashString(email, []byte(m.config.OTP.RegisterKey))
	// check rate limit
	count, err := m.otpRepo.CountRateLimit(ctx, hashedEmail, otptype.Register,
		m.config.OTP.RegisterRateLimitTTL)
	if err != nil {
		return err
	}
	if count > int64(m.config.OTP.RegisterRateLimit) {
		return errorcode.ErrOTPRateLimit
	}

	// check old otp if exists then delete
	if _, err := m.otpRepo.GetOTP(ctx, hashedEmail, otptype.Register); err == nil {
		go func() {
			if err := m.otpRepo.DeleteOTP(ctx, hashedEmail, otptype.Register); err != nil {
				m.logger.Warn("Cannot delete old otp", zap.Error(err))
			}
		}()
	}

	// save otp to redis
	if err := m.otpRepo.SetOTP(ctx, hashedEmail, otp, otptype.Register,
		m.config.OTP.RegisterTTL); err != nil {
		return err
	}

	// send otp to email
	go func() {
		err := sendto.SendTemplateEmailOtp(&m.config.SMTP, []string{email},
			"otp-register-verify.html", map[string]any{"otp": otp})
		if err != nil {
			m.logger.Error("Send email error", zap.Error(err))
		}
	}()

	return nil
}

func (m *userRegistrationManager) VerifyRegistrationOTP(ctx context.Context, email, otp string) (string, error) {
	// hash email
	hashedEmail := stringutils.HashString(email, []byte(m.config.OTP.RegisterKey))

	// check otp in redis
	storedOtp, err := m.otpRepo.GetOTP(ctx, hashedEmail, otptype.Register)
	if err != nil {
		return "", err
	}
	checkExceedAttempts := false
	if storedOtp != otp {
		// incr attempt
		attempts, err := m.otpRepo.IncrementAttempt(ctx, hashedEmail, otptype.Register,
			m.config.OTP.RegisterAttemptsTTL)
		if err != nil {
			return "", err
		}

		if attempts <= int64(m.config.OTP.RegisterAttempts) {
			return "", errorcode.ErrInvalidOTP
		}
		checkExceedAttempts = true
	}

	// delete otp in redis
	if err := m.otpRepo.DeleteOTP(ctx, hashedEmail, otptype.Register); err != nil {
		m.logger.Warn("Cannot delete old otp", zap.Error(err))
	}
	if err := m.otpRepo.ResetAttempt(ctx, hashedEmail, otptype.Register); err != nil {
		m.logger.Warn("Cannot reset attempt old otp", zap.Error(err))
	}

	if checkExceedAttempts {
		return "", errorcode.ErrOTPTooManyAttempts
	}

	// gene jwt token
	token, err := jwt.GenerateEmailToken([]byte(m.config.JWT.RegisterTokenKey), m.config.JWT.RegisterTokenExpiresIn,
		email, jwtpurpose.Register)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (m *userRegistrationManager) Register(ctx context.Context, dto user.CreateUserDto) (string, string, error) {
	g, gCtx := errgroup.WithContext(ctx)

	hpChan := make(chan string, 1)

	// check if username exists
	g.Go(func() error {
		exists, err := m.userRepo.IsUserNameTaken(gCtx, dto.UserName, uuid.Nil)
		if err != nil {
			return err
		}
		if exists {
			return errorcode.ErrInvalidUserName
		}
		return nil
	})

	// re-check if email exists
	g.Go(func() error {
		exists, err := m.userRepo.IsEmailTaken(gCtx, dto.Email, uuid.Nil)
		if err != nil && !errors.Is(err, errorcode.ErrUserNotFound) {
			return err
		}
		if exists {
			return errorcode.ErrExistedEmail
		}
		return nil
	})

	// hash pass
	g.Go(func() error {
		hp, err := password.HashPassword(dto.Password)
		if err != nil {
			return err
		}
		hpChan <- hp
		return nil
	})

	if err := g.Wait(); err != nil {
		return "", "", err
	}

	// get user role id
	defaultRole, ok := rolecache.Get("user")
	if !ok {
		return "", "", errorcode.ErrUnexpectedCreatingUser
	}

	// create user
	userID := uuid.New()
	user := &entities.User{
		ID:        userID,
		Email:     dto.Email,
		UserName:  dto.UserName,
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		Password:  <-hpChan,
		IsActive:  true,

		RoleID: defaultRole.ID,
	}

	var accessToken, refreshToken string
	// begin transaction
	err := m.uow.Do(ctx, func(r uow.UserManagerRepoProvider) error {
		// insert user into db
		err := r.UserRepository().Create(ctx, user)
		if err != nil {
			return err
		}

		// gene ac and rt
		accessToken, refreshToken, err = jwt.GenerateAcAndRtTokens(&m.config.JWT, user.ID)
		if err != nil {
			return err
		}

		// decode rt to get exp and iat
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
