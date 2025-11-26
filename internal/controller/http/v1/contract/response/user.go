package response

import (
	"time"

	"github.com/google/uuid"
)

type UserInfoRes struct {
	ID        uuid.UUID   `json:"id"`
	UserName  string      `json:"user_name"`
	FirstName string      `json:"first_name"`
	LastName  string      `json:"last_name"`
	CreatedAt time.Time   `json:"created_at"`
	Role      RoleInfoRes `json:"role"`
}
