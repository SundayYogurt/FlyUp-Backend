package domain

import "gorm.io/gorm"

type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Code string `gorm:"type:varchar(50);uniqueIndex;not null" json:"code"` // BOOSTER | PIONEER | ADMIN
	Name string `gorm:"type:varchar(100);not null" json:"name"`
	gorm.Model
}

type UserRole struct {
	ID     uint `gorm:"primaryKey" json:"id"`
	UserID uint `gorm:"index;not null" json:"user_id"`
	RoleID uint `gorm:"index;not null" json:"role_id"`
	gorm.Model
}
