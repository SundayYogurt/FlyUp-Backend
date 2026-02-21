package domain

import "time"

type AuditLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ActorID   uint      `gorm:"not null;index" json:"actor_id"` // admin/user
	Action    string    `gorm:"type:varchar(100);not null" json:"action"`
	Entity    string    `gorm:"type:varchar(100);not null" json:"entity"`
	EntityID  uint      `gorm:"not null;index" json:"entity_id"`
	Note      *string   `gorm:"type:text" json:"note,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}
