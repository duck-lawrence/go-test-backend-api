package implement

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

	manager := NewUserAuthManager(cfg, uowMock, userRepo, rtRepo, jwtSvc, pwSvc)
	return manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx
}

// -------------------- TEST LOGIN SUCCESS --------------------
func TestLogin_ValidInput_ReturnsAccessAndRefreshToken(t *testing.T) {
	// ----- ARRANGE: chuẩn bị test setup -----
	manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx := setupManager()

	// Test nhiều user khác nhau nhưng đều valid
	users := []struct {
		dto user.LoginUserDto
		id  uuid.UUID
		hpw string
	}{
		{user.LoginUserDto{EmailOrUsername: "john", Password: "plain"}, uuid.New(), "hashed"},
		{user.LoginUserDto{EmailOrUsername: "jane", Password: "123456"}, uuid.New(), "hashed2"},
	}

	for _, u := range users {
		t.Run(u.dto.EmailOrUsername, func(t *testing.T) {
			// ----- ARRANGE -----
			userEntity := &entities.User{ID: u.id, Password: u.hpw}
			claims := &externalservice.CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					IssuedAt:  jwt.NewNumericDate(time.Now()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}

			// setup behavior cho các mock
			userRepo.On("GetByUserNameOrEmail", ctx, u.dto.EmailOrUsername).Return(userEntity, nil)
			pwSvc.On("ComparePasswords", u.hpw, []byte(u.dto.Password)).Return(true)
			jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, u.id).Return("ac", "rt", nil)
			jwtSvc.On("ValidateToken", mock.Anything, "rt", jwtpurpose.Refresh).Return(claims, nil)
			rtRepo.On("Create", ctx, mock.Anything).Return(nil)

			// ----- ACT: gọi hàm cần test -----
			ac, rt, err := manager.Login(ctx, u.dto)

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

	inputs := []user.LoginUserDto{
		{EmailOrUsername: "unknown", Password: "123"},
		{EmailOrUsername: "ghost", Password: "abc"},
	}

	for _, dto := range inputs {
		t.Run(dto.EmailOrUsername, func(t *testing.T) {
			userRepo.On("GetByUserNameOrEmail", ctx, dto.EmailOrUsername).Return(nil, errorcode.ErrUserNotFound)

			ac, rt, err := manager.Login(ctx, dto)
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
	inputs := []user.LoginUserDto{
		{EmailOrUsername: "john", Password: "wrong"},
		{EmailOrUsername: "jane", Password: "badpass"},
	}

	for _, dto := range inputs {
		t.Run(dto.EmailOrUsername, func(t *testing.T) {
			userRepo.On("GetByUserNameOrEmail", ctx, dto.EmailOrUsername).Return(userEntity, nil)
			pwSvc.On("ComparePasswords", userEntity.Password, []byte(dto.Password)).Return(false)

			ac, rt, err := manager.Login(ctx, dto)
			require.Error(t, err)
			require.Equal(t, errorcode.ErrInvalidPassword, err)
			require.Empty(t, ac)
			require.Empty(t, rt)

			userRepo.AssertExpectations(t)
			pwSvc.AssertExpectations(t)
		})
	}
}

// -------------------- TEST JWT GENERATE OR VALIDATE TOKEN ERROR --------------------
func TestLogin_JwtGenerationOrValidateFails_ReturnsError(t *testing.T) {
	manager, userRepo, _, jwtSvc, pwSvc, ctx := setupManager()

	userID := uuid.New()
	userEntity := &entities.User{ID: userID, Password: "hashed"}
	dto := user.LoginUserDto{EmailOrUsername: "john", Password: "plain"}

	claims := &externalservice.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	tests := []struct {
		name             string
		mockGenerateErr  error
		mockValidateErr  error
		expectedErrorMsg string
	}{
		{
			name:             "JwtGenerationFails",
			mockGenerateErr:  errors.New("jwt error"),
			mockValidateErr:  nil,
			expectedErrorMsg: "jwt error",
		},
		{
			name:             "ValidateTokenFails",
			mockGenerateErr:  nil,
			mockValidateErr:  errors.New("token invalid"),
			expectedErrorMsg: "token invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// reset mocks
			userRepo.ExpectedCalls = nil
			jwtSvc.ExpectedCalls = nil
			pwSvc.ExpectedCalls = nil

			// setup mocks
			userRepo.On("GetByUserNameOrEmail", ctx, dto.EmailOrUsername).Return(userEntity, nil)
			pwSvc.On("ComparePasswords", userEntity.Password, []byte(dto.Password)).Return(true)
			jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, userID).Return("ac", "rt", tt.mockGenerateErr)
			if tt.mockValidateErr != nil {
				jwtSvc.On("ValidateToken", mock.Anything, "rt", jwtpurpose.Refresh).Return(claims, tt.mockValidateErr)
			}

			ac, rt, err := manager.Login(ctx, dto)
			require.Error(t, err)
			require.Equal(t, tt.expectedErrorMsg, err.Error())
			require.Empty(t, ac)
			require.Empty(t, rt)

			userRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
			pwSvc.AssertExpectations(t)
		})
	}
}

// -------------------- TEST REFRESH TOKEN REPO CREATE ERROR --------------------
func TestLogin_RefreshTokenCreateFails_ReturnsError(t *testing.T) {
	manager, userRepo, rtRepo, jwtSvc, pwSvc, ctx := setupManager()

	userID := uuid.New()
	userEntity := &entities.User{ID: userID, Password: "hashed"}
	dto := user.LoginUserDto{EmailOrUsername: "john", Password: "plain"}
	claims := &externalservice.CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	userRepo.On("GetByUserNameOrEmail", ctx, dto.EmailOrUsername).Return(userEntity, nil)
	pwSvc.On("ComparePasswords", userEntity.Password, []byte(dto.Password)).Return(true)
	jwtSvc.On("GenerateAcAndRtTokens", mock.Anything, userID).Return("ac", "rt", nil)
	jwtSvc.On("ValidateToken", mock.Anything, "rt", jwtpurpose.Refresh).Return(claims, nil)
	rtRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

	ac, rt, err := manager.Login(ctx, dto)
	require.Error(t, err)
	require.Equal(t, "db error", err.Error())
	require.Empty(t, ac)
	require.Empty(t, rt)

	userRepo.AssertExpectations(t)
	rtRepo.AssertExpectations(t)
	jwtSvc.AssertExpectations(t)
	pwSvc.AssertExpectations(t)
}

// -------------------- TEST PANIC UNIMPLEMENT --------------------
// func TestLogout_Panic_BranchCoverage(t *testing.T) {
// 	manager, _, _, _, _, ctx := setupManager()

// 	require.Panics(t, func() {
// 		_ = manager.Logout(ctx, user.LogoutUserDto{RefreshToken: ""})
// 	})
// 	require.Panics(t, func() {
// 		_ = manager.Logout(ctx, user.LogoutUserDto{RefreshToken: "something"})
// 	})
// }

// func TestRefreshToken_Panic_BranchCoverage(t *testing.T) {
// 	manager, _, _, _, _, ctx := setupManager()

// 	require.Panics(t, func() {
// 		_, _, _ = manager.RefreshToken(ctx, "")
// 	})
// 	require.Panics(t, func() {
// 		_, _, _ = manager.RefreshToken(ctx, "token")
// 	})
// }
