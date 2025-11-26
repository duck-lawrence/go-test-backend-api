package user

import "github.com/google/uuid"

type CreateUserDto struct {
	Email     string
	UserName  string
	FirstName string
	LastName  string
	Password  string
}

type RestoreUserDto struct {
	Email       string
	NewPassword string
}

type LoginUserDto struct {
	EmailOrUsername string
	Password        string
}

type LogoutUserDto struct {
	UserID       uuid.UUID
	RefreshToken string
}

type UpdateMeDto struct {
	UserID    uuid.UUID
	UserName  string
	FirstName string
	LastName  string
}

type ChangePasswordDto struct {
	UserID      uuid.UUID
	OldPassword string
	NewPassword string
}
