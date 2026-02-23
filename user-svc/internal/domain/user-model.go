package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                         uint       `gorm:"primaryKey" json:"id"`
	Email                      string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash               string     `gorm:"type:varchar(255);not null" json:"-"`
	GoogleSub                  *string    `gorm:"type:varchar(255)" json:"google_sub,omitempty"`
	FirstName                  string     `gorm:"type:varchar(100)" json:"first_name"`
	LastName                   string     `gorm:"type:varchar(100)" json:"last_name"`
	Phone                      string     `gorm:"type:varchar(50)" json:"phone"`
	Status                     string     `gorm:"type:varchar(20);not null;default:'active'" json:"status"` // active|suspended|deleted
	EmailVerifiedAt            *time.Time `json:"email_verified_at,omitempty"`
	VerificationToken          string     `gorm:"type:varchar(255)" json:"-"`
	VerificationTokenExpiresAt *time.Time `json:"verification_token_expires_at,omitempty"`
	ResetTokenHash             string     `gorm:"type:varchar(255)" json:"-"`
	ResetTokenExpiresAt        *time.Time `json:"reset_token_expires_at,omitempty"`
	gorm.Model
}
