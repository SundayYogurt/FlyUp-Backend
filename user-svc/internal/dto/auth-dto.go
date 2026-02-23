package dto

type RegisterRequest struct {
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	FirstName    string   `json:"first_name"`
	LastName     string   `json:"last_name"`
	Phone        string   `json:"phone"`
	Role         string   `json:"role"`
	ConsentCodes []string `json:"consent_codes"`
}

type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	Token string              `json:"token"`
	User  UserProfileResponse `json:"user"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type SetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

type AuthResponse struct {
	UserID int     `json:"user_id"`
	Email  string  `json:"email"`
	Iat    float64 `json:"iat"`
	Expiry float64 `json:"expiry"`
}

type VerifyEmailRequest struct {
	Token string `json:"token"`
}
