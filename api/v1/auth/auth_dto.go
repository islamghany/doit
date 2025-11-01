package auth

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RegisterInput struct {
	Email           string `json:"email" validate:"required,email"`
	Username        string `json:"username" validate:"required,max=50"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
	Role            string `json:"role" validate:"required,oneof=user admin moderator"`
}
