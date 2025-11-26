package user

import (
	"net/http"

	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/contract/request"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/validation"
	"github.com/gin-gonic/gin"
)

type UserRestoreController struct {
	restore user.UserRestoreManager
}

func NewUserRestoreController(
	restore user.UserRestoreManager,
) *UserRestoreController {
	return &UserRestoreController{
		restore: restore,
	}
}

func (uc *UserRestoreController) SendRestoreOTP(c *gin.Context) {
	var req request.SendEmailOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()
	if err := uc.restore.SendRestoreOTP(ctx, req.Email); err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Please check your email to get OTP",
	})
}

func (uc *UserRestoreController) VerifyRestoreOTP(c *gin.Context) {
	var req request.VerifyEmailOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()
	restoreToken, err := uc.restore.VerifyRestoreOTP(ctx, req.Email, req.OTP)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Verify email success",
		"restore_token": restoreToken,
	})
}

func (uc *UserRestoreController) Restore(c *gin.Context) {
	var req request.RestoreUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing email in email verified token"})
		return
	}
	dto := user.RestoreUserDto{
		Email:       email.(string),
		NewPassword: req.NewPassword,
	}

	accessToken, refreshToken, err := uc.restore.Restore(ctx, dto)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "restore success",
		"token": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}
