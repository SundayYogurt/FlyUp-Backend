package domain

import "gorm.io/gorm"

type UserBankAccount struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	UserID      uint   `gorm:"index;not null" json:"user_id"`
	BankCode    string `gorm:"type:varchar(20);not null" json:"bank_code"`
	AccountNo   string `gorm:"type:varchar(50);not null" json:"account_no"`
	AccountName string `gorm:"type:varchar(100);not null" json:"account_name"`

	IsDefault bool   `gorm:"default:false" json:"is_default"`
	Status    string `gorm:"type:varchar(20);default:active" json:"status"` // active | disabled

	gorm.Model
}
