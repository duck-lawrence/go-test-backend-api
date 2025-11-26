package request

type SendEmailOTPReq struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailOTPReq struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

type CreateUserReq struct {
	UserName        string `json:"user_name" binding:"required,username"`
	FirstName       string `json:"first_name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Password        string `json:"password" binding:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=Password"`
}

type LoginUserReq struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required,min=8,max=30"`
}

type LogoutUserReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateMeUserReq struct {
	UserName  string `json:"user_name"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type ChangePasswordReq struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8,max=30,neqfield=OldPassword"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RestoreUserReq struct {
	NewPassword     string `json:"new_password" binding:"required,min=8,max=30"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}
