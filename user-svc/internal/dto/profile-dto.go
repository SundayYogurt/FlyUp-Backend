package dto

type UpdateUserProfile struct {
	DisplayName string  `json:"display_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
}

type UserProfileResponse struct {
	ID          uint    `json:"id"`
	Email       string  `json:"email"`
	DisplayName string  `json:"display_name"`
	Phone       *string `json:"phone,omitempty"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}
