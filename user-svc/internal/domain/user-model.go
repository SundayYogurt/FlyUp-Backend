package domain

import "gorm.io/gorm"

type User struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	Email        string  `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string  `json:"-"`
	GoogleSub    *string `gorm:"uniqueIndex" json:"google_sub,omitempty"`
	DisplayName  string  `json:"display_name"`
	Phone        *string `json:"phone,omitempty"`
	Status       string  `gorm:"type:varchar(20);not null;default:active" json:"status"`
	ResetToken   string  `json:"reset_token"`
	gorm.Model
}
