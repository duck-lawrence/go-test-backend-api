package user

import (
	"net/http"

	"github.com/ducklawrence05/go-test-backend-api/internal/constants/errorcode"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/contract/request"
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/mapper"
	"github.com/ducklawrence05/go-test-backend-api/internal/usecase/user"
	"github.com/ducklawrence05/go-test-backend-api/pkg/utils/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserProfileController struct {
	profile user.UserProfileManager
}

func NewUserProfileController(
	profile user.UserProfileManager,
) *UserProfileController {
	return &UserProfileController{
		profile: profile,
	}
}

func (uc *UserProfileController) GetMe(c *gin.Context) {
	// get userID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID in access token"})
		return
	}

	ctx := c.Request.Context()

	user, err := uc.profile.GetMe(ctx, userID.(uuid.UUID))
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.ToUserInfoResponse(user))
}

func (uc *UserProfileController) UpdateMe(c *gin.Context) {
	// get userID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID in access token"})
		return
	}

	var req request.UpdateMeUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	dto := user.UpdateMeDto{
		UserID:    userID.(uuid.UUID),
		UserName:  req.UserName,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	ctx := c.Request.Context()

	user, err := uc.profile.UpdateMe(ctx, dto)
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.ToUserInfoResponse(user))
}

func (uc *UserProfileController) ChangePassword(c *gin.Context) {
	// get userID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID in access token"})
		return
	}

	var req request.ChangePasswordReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": validation.TranslateValidationError(err),
		})
		return
	}

	dto := user.ChangePasswordDto{
		UserID:      userID.(uuid.UUID),
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	}

	ctx := c.Request.Context()

	if err := uc.profile.ChangePassword(ctx, dto); err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "change password success"})
}

func (uc *UserProfileController) DeleteMe(c *gin.Context) {
	// get userID from middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing userID in access token"})
		return
	}

	ctx := c.Request.Context()

	err := uc.profile.DeleteMe(ctx, userID.(uuid.UUID))
	if err != nil {
		errorcode.JSONError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "delete success"})
}
