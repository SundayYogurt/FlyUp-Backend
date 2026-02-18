package domain

import (
	"time"

	"gorm.io/gorm"
)

const (
	KYCStatusPending  = "pending"
	KYCStatusApproved = "approved"
	KYCStatusRejected = "rejected"
)

const (
	KYCDecisionApproved = "approved"
	KYCDecisionRejected = "rejected"
)

const (
	KYCDocTypeIDCard      = "id_card"
	KYCDocTypeStudentCard = "student_card"
	KYCDocTypeSelfie      = "selfie"
	KYCDocTypeOther       = "other"
)

type KYCSubmission struct {
	ID          uint          `gorm:"primary_key;autoIncrement" json:"id"`
	UserID      uint          `gorm:"not null;index" json:"user_id"`
	Status      string        `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending/approved/rejected
	SubmittedAt time.Time     `gorm:"not null;autoCreateTime" json:"submitted_at"`
	ReviewedAt  *time.Time    `json:"reviewed_at,omitempty"`
	ReviewedBy  *uint         `json:"reviewed_by,omitempty"`                       // admin user_id
	Documents   []KYCDocument `gorm:"foreignKey:KYCID" json:"documents,omitempty"` // Relations
	Review      *KYCReview    `gorm:"foreignKey:KYCID" json:"review,omitempty"`    // Relations
	gorm.Model
}

type KYCDocument struct {
	ID      uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	KYCID   uint   `gorm:"not null;index" json:"kyc_id"`
	DocType string `gorm:"type:varchar(30);not null" json:"doc_type"` // id_card/student_card/selfie/other
	FileURL string `gorm:"type:text;not null" json:"file_url"`
	gorm.Model
}

type KYCReview struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	KYCID      uint      `gorm:"uniqueIndex;not null" json:"kyc_id"` // 1 review ต่อ 1 kyc
	ReviewedBy uint      `gorm:"not null;index" json:"reviewed_by"`
	Decision   string    `gorm:"type:varchar(20);not null" json:"decision"` // approved/rejected
	Note       string    `gorm:"type:text" json:"note"`
	ReviewedAt time.Time `gorm:"not null;autoCreateTime" json:"reviewed_at"`
	gorm.Model
}
