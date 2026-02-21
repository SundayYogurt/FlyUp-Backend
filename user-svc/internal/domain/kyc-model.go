package domain

import "time"

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
	ID          uint          `gorm:"primaryKey" json:"id"`
	UserID      uint          `gorm:"not null;index" json:"user_id"`
	Status      string        `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	SubmittedAt time.Time     `gorm:"autoCreateTime" json:"submitted_at"`
	ReviewedAt  *time.Time    `json:"reviewed_at,omitempty"`
	ReviewedBy  *uint         `gorm:"index" json:"reviewed_by,omitempty"` // admin user_id
	Documents   []KYCDocument `gorm:"foreignKey:KYCID" json:"documents,omitempty"`
	Review      *KYCReview    `gorm:"foreignKey:KYCID" json:"review,omitempty"`
	CreatedAt   time.Time     `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time     `gorm:"autoUpdateTime" json:"updated_at"`
}

type KYCDocument struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	KYCID     uint      `gorm:"not null;index" json:"kyc_id"`
	DocType   string    `gorm:"type:varchar(30);not null" json:"doc_type"`
	FileURL   string    `gorm:"type:text;not null" json:"file_url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type KYCReview struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	KYCID      uint      `gorm:"uniqueIndex;not null" json:"kyc_id"`
	ReviewedBy uint      `gorm:"not null;index" json:"reviewed_by"`
	Decision   string    `gorm:"type:varchar(20);not null" json:"decision"`
	Note       string    `gorm:"type:text" json:"note"`
	ReviewedAt time.Time `gorm:"autoCreateTime" json:"reviewed_at"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
