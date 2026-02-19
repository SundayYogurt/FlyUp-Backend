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
	UniversityID   uint       `gorm:"not null" json:"university_id"`
	StudentCode    string     `gorm:"column:student_code" json:"student_code"`
	Faculty        string     `json:"faculty"`
	Major          string     `json:"major"`
	YearLevel      string     `gorm:"column:year_level" json:"year_level"`
	ReviewedBy     string     `gorm:"column:reviewed_by" json:"reviewed_by"`
	Note           string     `gorm:"column:note" json:"note"`
	StudentCardURL string     `gorm:"column:student_card_url" json:"student_card_url"`
	VerifyStatus   string     `gorm:"column:verify_status;type:varchar(20);not null;default:pending" json:"verify_status"`
	VerifiedAt     *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`

	gorm.Model
}
