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
	UserID      uint       `gorm:"not null;index;uniqueIndex:uidx_user_consents_user_code" json:"user_id"`
	ConsentCode string     `gorm:"type:varchar(50);not null;uniqueIndex:uidx_user_consents_user_code" json:"consent_code"`
	Accepted    bool       `gorm:"not null;default:true" json:"accepted"`
	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	gorm.Model
}
