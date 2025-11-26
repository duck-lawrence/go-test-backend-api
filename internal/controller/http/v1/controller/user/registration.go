package user

import (
	"net/http"

	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/contract/request"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/validation"
	"github.com/gin-gonic/gin"
)

type UserRegistrationController struct {
	registration user.UserRegistrationManager
}

func NewUserRegistrationController(
	registration user.UserRegistrationManager,
) *UserRegistrationController {
	return &UserRegistrationController{
		registration: registration,
	}
}

func (uc *UserRegistrationController) SendRegistrationOTP(c *gin.Context) {
	var req request.SendEmailOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()
	if err := uc.registration.SendRegistrationOTP(ctx, req.Email); err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Please check your email to get OTP",
	})
}

func (uc *UserRegistrationController) VerifyRegistrationOTP(c *gin.Context) {
	var req request.VerifyEmailOTPReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()
	emailVerifiedToken, err := uc.registration.VerifyRegistrationOTP(ctx, req.Email, req.OTP)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Verify email success",
		"email_verified_token": emailVerifiedToken,
	})
}

func (uc *UserRegistrationController) Register(c *gin.Context) {
	var req request.CreateUserReq
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
	dto := user.CreateUserDto{
		Email:     email.(string),
		UserName:  req.UserName,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Password:  req.Password,
	}

	accessToken, refreshToken, err := uc.registration.Register(ctx, dto)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "register success",
		"token": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}
