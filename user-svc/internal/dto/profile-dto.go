package dto

type UpdateUserProfile struct {
	FirstName   *string           `json:"first_name,omitempty"`
	LastName    *string           `json:"last_name,omitempty"`
	Phone       *string           `json:"phone,omitempty"`
	BankAccount *BankAccountInput `json:"bank_account,omitempty"`
}

type BankAccountInput struct {
	BankCode    string `json:"bank_code" validate:"required,oneof=SCB KTB KBANK BBL BAY TTB GSB BAAC"`
	AccountNo   string `json:"account_no" validate:"required"`
	AccountName string `json:"account_name" validate:"required"`
	IsDefault   bool   `json:"is_default"`
}

type UserProfileResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
