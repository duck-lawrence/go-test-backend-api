package user

import (
	"github.com/ducklawrence05/go-test-backend-api/config"
	"github.com/ducklawrence05/go-test-backend-api/pkg/logger"

	"github.com/ducklawrence05/go-test-backend-api/internal/constants/jwtpurpose"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/middleware"
	controller "github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/controller/user"
	"github.com/ducklawrence05/go-test-backend-api/internal/infrastructure/wire/managers"
	"github.com/gin-gonic/gin"
)

type UserRouterConfig struct {
	Config *config.Config
	Logger logger.Interface
}

type UserRouter struct{}

func (u *UserRouter) NewUserRouter(
	router *gin.RouterGroup,
	cfg *UserRouterConfig,
	mSet *managers.UserManagerSet,
) {
	// New controller
	profileCtrl := controller.NewUserProfileController(mSet.Profile)
	registrationCtrl := controller.NewUserRegistrationController(mSet.Registration)
	restoreCtrl := controller.NewUserRestoreController(mSet.Restore)
	authCtrl := controller.NewUserAuthController(mSet.Auth)

	// ===== Public routes =====
	public := router.Group("/user")
	{
		public.POST("/login", authCtrl.Login)
		public.POST("/refresh-token", authCtrl.RefreshToken)
	}

	// Register route
	register := public.Group("/register")
	{
		register.POST("/send-email-otp", registrationCtrl.SendRegistrationOTP)
		register.POST("/verify-email-otp", registrationCtrl.VerifyRegistrationOTP)
		register.POST("/complete",
			middleware.ValidateToken(cfg.Logger, []byte(cfg.Config.JWT.RegisterTokenKey), jwtpurpose.Register),
			registrationCtrl.Register,
		)
	}

	// Restore
	restore := public.Group("/restore")
	{
		restore.POST("/send-email-otp", restoreCtrl.SendRestoreOTP)
		restore.POST("/verify-email-otp", restoreCtrl.VerifyRestoreOTP)
		restore.POST("/complete",
			middleware.ValidateToken(cfg.Logger, []byte(cfg.Config.JWT.RestoreAccountTokenKey), jwtpurpose.Restore),
			restoreCtrl.Restore,
		)
	}

	// ===== Private routes (need access token) =====
	private := router.Group("/user")
	// middleware
	private.Use(middleware.ValidateToken(cfg.Logger, []byte(cfg.Config.JWT.AccessTokenKey), jwtpurpose.Access))
	// controller
	{
		private.POST("/logout", authCtrl.Logout)
		private.GET("/me", profileCtrl.GetMe)
		private.PATCH("/me", profileCtrl.UpdateMe)
		private.PUT("/change-password", profileCtrl.ChangePassword)
		private.DELETE("/me", profileCtrl.DeleteMe)
	}
}
