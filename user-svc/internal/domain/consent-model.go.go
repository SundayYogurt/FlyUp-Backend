package domain

import (
	"time"

	"gorm.io/gorm"
)

const (
	ConsentInfoTrue     = "INFO_TRUE"
	ConsentPioneerTerms = "PIONEER_TERMS"
	ConsentTerms        = "TERMS"
)

type UserConsent struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null;index:idx_user_consents_user" json:"user_id"`
	ConsentCode string     `gorm:"type:varchar(50);not null;index:uidx_user_consents_code_ver,unique" json:"consent_code"`
	Accepted    bool       `gorm:"not null;default:true" json:"accepted"`
	AcceptedAt  *time.Time `gorm:"" json:"accepted_at,omitempty"`
	gorm.Model
}
