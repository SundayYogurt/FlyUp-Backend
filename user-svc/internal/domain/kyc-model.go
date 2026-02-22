package domain

import (
	"time"

	"gorm.io/gorm"
)

type KYCStatus string

const (
	KYCStatusPending      KYCStatus = "pending"
	KYCStatusAutoApproved KYCStatus = "auto_approved" // ผ่าน OCR / Face match
	KYCStatusApproved     KYCStatus = "approved"      // admin อนุมัติ
	KYCStatusRejected     KYCStatus = "rejected"
)

type KYCDecision string

const (
	KYCDecisionApproved KYCDecision = "approved"
	KYCDecisionRejected KYCDecision = "rejected"
)

type KYCDocType string

const (
	KYCDocTypeIDCard      KYCDocType = "id_card"
	KYCDocTypeStudentCard KYCDocType = "student_card"
	KYCDocTypeSelfie      KYCDocType = "selfie"
	KYCDocTypeOther       KYCDocType = "other"
)

type KYCSubmission struct {
	ID     uint      `gorm:"primaryKey" json:"id"`
	UserID uint      `gorm:"not null;index" json:"user_id"`
	Status KYCStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`

	// --- iApp / Auto KYC ---
	OCRProvider     *string  `gorm:"type:varchar(50)" json:"ocr_provider,omitempty"` // iapp
	FaceMatchScore  *float64 `json:"face_match_score,omitempty"`
	FaceMatchPassed *bool    `json:"face_match_passed,omitempty"`
	OCRError        *string  `gorm:"type:text" json:"ocr_error,omitempty"`
	AutoApproved    bool     `gorm:"default:false" json:"auto_approved"`

	// --- Relations ---
	Documents []KYCDocument `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:KYCID" json:"documents,omitempty"`
	Review    *KYCReview    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;foreignKey:KYCID" json:"review,omitempty"`

	// --- Timestamps ---
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	gorm.Model
}

type KYCDocument struct {
	ID      uint       `gorm:"primaryKey" json:"id"`
	KYCID   uint       `gorm:"not null;index" json:"kyc_id"`
	DocType KYCDocType `gorm:"type:varchar(30);not null" json:"doc_type"`
	FileURL string     `gorm:"type:text;not null" json:"file_url"`

	// optional metadata
	MimeType *string `gorm:"type:varchar(100)" json:"mime_type,omitempty"`
	FileSize *int64  `json:"file_size,omitempty"`
	FileHash *string `gorm:"type:varchar(128)" json:"file_hash,omitempty"` // sha256

	gorm.Model
}

type KYCReview struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	KYCID      uint        `gorm:"uniqueIndex;not null" json:"kyc_id"`
	ReviewedBy uint        `gorm:"not null;index" json:"reviewed_by"` // admin user_id
	Decision   KYCDecision `gorm:"type:varchar(20);not null" json:"decision"`
	Note       *string     `gorm:"type:text" json:"note,omitempty"`

	ReviewedAt time.Time `gorm:"autoCreateTime" json:"reviewed_at"`
	gorm.Model
}
