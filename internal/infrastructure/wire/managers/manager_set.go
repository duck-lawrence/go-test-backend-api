package managers

import (
	otpUC "github.com/ducklawrence05/go-test-backend-api/internal/usecase/otp"
	roleUC "github.com/ducklawrence05/go-test-backend-api/internal/usecase/role"
	userUC "github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
)

type ManagerSet struct {
	UserRegistration userUC.UserRegistrationManager
	UserRestore      userUC.UserRestoreManager
	UserAuth         userUC.UserAuthManager
	UserProfile      userUC.UserProfileManager
	Role             roleUC.RoleManager
	OTPRateLimit     otpUC.OTPRateLimitManager
	OTPVerify        otpUC.OTPVerifyManager
}

type UserManagerSet struct {
	Registration userUC.UserRegistrationManager
	Restore      userUC.UserRestoreManager
	Auth         userUC.UserAuthManager
	Profile      userUC.UserProfileManager
	OTPRateLimit otpUC.OTPRateLimitManager
	OTPVerify    otpUC.OTPVerifyManager
}
