package user

import (
	"net/http"

	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/contract/request"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserAuthController struct {
	auth user.UserAuthManager
}

func NewUserAuthController(
	auth user.UserAuthManager,
) *UserAuthController {
	return &UserAuthController{
		auth: auth,
	}
}

func (uc *UserAuthController) Login(c *gin.Context) {
	var req request.LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	dto := user.LoginUserDto{
		EmailOrUsername: req.UserName,
		Password:        req.Password,
	}

	ctx := c.Request.Context()

	accessToken, refreshToken, err := uc.auth.Login(ctx, dto)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "login success",
		"token": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}

func (uc *UserAuthController) Logout(c *gin.Context) {
	var req request.LogoutUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	// get userID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID in access token"})
		return
	}

	dto := user.LogoutUserDto{
		UserID:       userID.(uuid.UUID),
		RefreshToken: req.RefreshToken,
	}

	ctx := c.Request.Context()

	err := uc.auth.Logout(ctx, dto)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logout success"})
}

func (uc *UserAuthController) RefreshToken(c *gin.Context) {
	var req request.RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	ctx := c.Request.Context()

	accessToken, refreshToken, err := uc.auth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "refresh token success",
		"token": gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		},
	})
}
