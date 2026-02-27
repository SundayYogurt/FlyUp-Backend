package dto

type VerifyEmailEvent struct {
	UserID    uint   `json:"user_id"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}
