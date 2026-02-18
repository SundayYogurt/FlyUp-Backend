package domain

import "time"

type AuditLog struct {
	ID        uint   `gorm:"primaryKey"`
	ActorID   uint   `gorm:"not null"` // admin
	Action    string `gorm:"not null"` // approve_kyc
	Entity    string `gorm:"not null"` // kyc_submission
	EntityID  uint   `gorm:"not null"`
	Note      string
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
