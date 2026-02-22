package dto

type RegisterRequest struct {
	Email       string  `json:"email" validate:"required,email"`
	Password    string  `json:"password" validate:"required,min=6"`
	DisplayName string  `json:"display_name" validate:"required"`
	Phone       *string `json:"phone,omitempty"`

	Role string `json:"role" validate:"required,oneof=BOOSTER PIONEER"`

	Pioneer *PioneerInput         `json:"pioneer,omitempty"` // ต้องมีถ้า role=PIONEER
	Booster *BoosterRegisterInput `json:"booster,omitempty"`
}

type BoosterRegisterInput struct {

	// ถ้าส่งมา แปลว่า "ผูกบัญชีธนาคารตอนสมัคร"
	BankAccount *BankAccountInput `json:"bank_account,omitempty"`
}

type BankAccountInput struct {
	BankCode    string `json:"bank_code" validate:"required,oneof=SCB KTB KBANK BBL BAY TTB GSB BAAC"`
	AccountNo   string `json:"account_no" validate:"required"`
	AccountName string `json:"account_name" validate:"required"`
	IsDefault   bool   `json:"is_default"`
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
