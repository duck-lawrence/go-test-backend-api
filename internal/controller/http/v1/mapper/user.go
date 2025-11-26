package mapper

import (
	"github.com/ducklawrence05/go-test-backend-api/internal/controller/http/v1/contract/response"
	"github.com/ducklawrence05/go-test-backend-api/internal/entities"
)

func ToUserInfoResponse(user *entities.User) *response.UserInfoRes {
	return &response.UserInfoRes{
		ID:        user.ID,
		UserName:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		CreatedAt: user.CreatedAt,
		Role: response.RoleInfoRes{
			Name:        user.Role.Name,
			Description: user.Role.Description,
		},
	}
}
