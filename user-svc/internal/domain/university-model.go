package domain

import "time"

type University struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	NameTH    string    `gorm:"type:varchar(255);column:name_th" json:"name_th"`
	NameEN    string    `gorm:"type:varchar(255);column:name_en" json:"name_en"`
	Province  string    `gorm:"type:varchar(255);column:province" json:"province"`
	Domain    string    `gorm:"type:varchar(255);uniqueIndex;column:domain" json:"domain,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
