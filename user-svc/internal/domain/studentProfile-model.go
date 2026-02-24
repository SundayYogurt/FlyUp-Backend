package domain

import (
	"time"

	"gorm.io/gorm"
)

const (
	StudentVerifyPending  = "pending"
	StudentVerifyApproved = "approved"
	StudentVerifyRejected = "rejected"
)

type StudentProfile struct {
	ID             uint       `gorm:"primaryKey" json:"id"`
	UserID         uint       `gorm:"uniqueIndex;not null" json:"user_id"`
	UniversityID   *uint      `json:"university_id,omitempty"`
	StudentCode    string     `json:"student_code"`
	Faculty        string     `json:"faculty"`
	Major          string     `json:"major"`
	ReviewedBy     *uint      `json:"reviewed_by,omitempty"` //  admin user_id
	Note           *string    `json:"note,omitempty"`
	StudentCardURL *string    `json:"student_card_url,omitempty"`
	VerifyStatus   string     `gorm:"type:varchar(20);not null;default:pending" json:"verify_status"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`

	gorm.Model
}
