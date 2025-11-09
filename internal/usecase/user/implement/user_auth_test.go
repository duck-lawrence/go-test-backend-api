package implement_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/externalservice"
	useCaseMock "github.com/ducklawrence05/go-test-backend-api/internal/usecase/mock"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user/implement"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupManager() (user.UserAuthManager,
	*useCaseMock.MockUserRepo,
	*useCaseMock.MockRefreshTokenRepo,
	*useCaseMock.MockJwtService,
	*useCaseMock.MockPasswordService,
	context.Context) {

	ctx := context.Background()
	cfg := &config.Config{
		JWT: config.JWT{
			AccessTokenKey:        "access",
			RefreshTokenKey:       "refresh",
			AccessTokenExpiresIn:  time.Hour,
			RefreshTokenExpiresIn: 24 * time.Hour,
		},
	}

	userRepo := new(useCaseMock.MockUserRepo)
	rtRepo := new(useCaseMock.MockRefreshTokenRepo)
	jwtSvc := new(useCaseMock.MockJwtService)
	pwSvc := new(useCaseMock.MockPasswordService)
	uowMock := new(useCaseMock.MockUserManagerUow)

	manager := implement.NewUserAuthManager(cfg, uowMock, userRepo, rtRepo, jwtSvc, pwSvc)
	return manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx
}

// -------------------- TEST LOGIN SUCCESS --------------------
func TestLogin_ValidInput_ReturnsAccessAndRefreshToken(t *testing.T) {
	// ----- ARRANGE: chuẩn bị test setup -----
	manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx := setupManager()

	// Test nhiều user khác nhau nhưng đều valid
	users := []struct {
		vo  user.LoginUserVO
		id  uuid.UUID
		hpw string
	}{
		{user.LoginUserVO{EmailOrUsername: "john", Password: "plain"}, uuid.New(), "hashed"},
		{user.LoginUserVO{EmailOrUsername: "jane", Password: "123456"}, uuid.New(), "hashed2"},
	}

	for _, u := range users {
		t.Run(u.vo.EmailOrUsername, func(t *testing.T) {
			// ----- ARRANGE -----
			userEntity := &entities.User{ID: u.id, Password: u.hpw}
			claims := &externalservice.CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}

			// setup behavior cho các mock
			userRepo.On("GetByUserNameOrEmail", ctx, u.vo.EmailOrUsername).Return(userEntity, nil)
			pwSvc.On("ComparePasswords", u.hpw, []byte(u.vo.Password)).Return(true)
			jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, u.id).Return("ac", "rt", nil)
			jwtSvc.On("ValidateToken", mock.Anything, "rt", jwtpurpose.Refresh).Return(claims, nil)
			rtRepo.On("Create", ctx, mock.Anything).Return(nil)

			// ----- ACT: gọi hàm cần test -----
			ac, rt, err := manager.Login(ctx, u.vo)

			// ----- ASSERT: kiểm tra kết quả -----
			require.NoError(t, err)    // không có lỗi
			require.Equal(t, "ac", ac) // access token đúng
			require.Equal(t, "rt", rt) // refresh token đúng

			// kiểm tra mock được gọi đúng như setup
			userRepo.AssertExpectations(t)
			rtRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
			pwSvc.AssertExpectations(t)
		})
	}
}

// -------------------- TEST USER NOT FOUND --------------------
func TestLogin_UserNotFound_ReturnsError(t *testing.T) {
	manager, userRepo, _, _, _, ctx := setupManager()

	inputs := []user.LoginUserVO{
		{EmailOrUsername: "unknown", Password: "123"},
		{EmailOrUsername: "ghost", Password: "abc"},
	}

	for _, vo := range inputs {
		t.Run(vo.EmailOrUsername, func(t *testing.T) {
			userRepo.On("GetByUserNameOrEmail", ctx, vo.EmailOrUsername).Return(nil, errorcode.ErrUserNotFound)

			ac, rt, err := manager.Login(ctx, vo)
			require.Error(t, err)
			require.Equal(t, errorcode.ErrUserNotFound, err)
			require.Empty(t, ac)
			require.Empty(t, rt)

			userRepo.AssertExpectations(t)
		})
	}
}

// -------------------- TEST INVALID PASSWORD --------------------
func TestLogin_InvalidPassword_ReturnsError(t *testing.T) {
	manager, userRepo, _, _, pwSvc, ctx := setupManager()

	userID := uuid.New()
	userEntity := &entities.User{ID: userID, Password: "hashed"}
	inputs := []user.LoginUserVO{
		{EmailOrUsername: "john", Password: "wrong"},
		{EmailOrUsername: "jane", Password: "badpass"},
	}

	for _, vo := range inputs {
		t.Run(vo.EmailOrUsername, func(t *testing.T) {
			userRepo.On("GetByUserNameOrEmail", ctx, vo.EmailOrUsername).Return(userEntity, nil)
			pwSvc.On("ComparePasswords", userEntity.Password, []byte(vo.Password)).Return(false)

			ac, rt, err := manager.Login(ctx, vo)
			require.Error(t, err)
			require.Equal(t, errorcode.ErrInvalidPassword, err)
			require.Empty(t, ac)
			require.Empty(t, rt)

			userRepo.AssertExpectations(t)
			pwSvc.AssertExpectations(t)
		})
	}
}

// -------------------- TEST JWT GENERATE ERROR --------------------
func TestLogin_JwtGenerationFails_ReturnsError(t *testing.T) {
	manager, userRepo, _, jwtSvc, pwSvc, ctx := setupManager()

	userID := uuid.New()
	userEntity := &entities.User{ID: userID, Password: "hashed"}
	vo := user.LoginUserVO{EmailOrUsername: "john", Password: "plain"}

	userRepo.On("GetByUserNameOrEmail", ctx, vo.EmailOrUsername).Return(userEntity, nil)
	pwSvc.On("ComparePasswords", userEntity.Password, []byte(vo.Password)).Return(true)
	jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, userID).Return("", "", errors.New("jwt error"))

	ac, rt, err := manager.Login(ctx, vo)
	require.Error(t, err)
	require.Equal(t, "jwt error", err.Error())
	require.Empty(t, ac)
	require.Empty(t, rt)

	userRepo.AssertExpectations(t)
	jwtSvc.AssertExpectations(t)
	pwSvc.AssertExpectations(t)
}

// -------------------- TEST REFRESH TOKEN REPO CREATE ERROR --------------------
func TestLogin_RefreshTokenCreateFails_ReturnsError(t *testing.T) {
	manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx := setupManager()

	userID := uuid.New()
	userEntity := &entities.User{ID: userID, Password: "hashed"}
	vo := user.LoginUserVO{EmailOrUsername: "john", Password: "plain"}
	claims := &externalservice.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	userRepo.On("GetByUserNameOrEmail", ctx, vo.EmailOrUsername).Return(userEntity, nil)
	pwSvc.On("ComparePasswords", userEntity.Password, []byte(vo.Password)).Return(true)
	jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, userID).Return("ac", "rt", nil)
	jwtSvc.On("ValidateToken", mock.Anything, "rt", jwtpurpose.Refresh).Return(claims, nil)
	rtRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

	ac, rt, err := manager.Login(ctx, vo)
	require.Error(t, err)
	require.Equal(t, "db error", err.Error())
	require.Empty(t, ac)
	require.Empty(t, rt)

	userRepo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
	jwtSvc.AssertExpectations(t)
	pwSvc.AssertExpectations(t)
}
