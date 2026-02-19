package domain

import (
	"gorm.io/gorm"
)

type University struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	NameTH   string `gorm:"column:name_th" json:"name_th"`
	NameEN   string `gorm:"column:name_en" json:"name_en"`
	Province string `gorm:"column:province" json:"province"`
	Domain   string `gorm:"column:domain" json:"domain"`

	gorm.Model
}
