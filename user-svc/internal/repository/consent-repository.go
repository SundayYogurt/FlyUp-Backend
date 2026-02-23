package repository

import (
	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type ConsentRepository interface {
	CreateConsent(consent *domain.UserConsent) error
}

type consentRepository struct {
	db *gorm.DB
}

func NewConsentRepository(db *gorm.DB) ConsentRepository {
	return &consentRepository{db: db}
}

func (c consentRepository) CreateConsent(consent *domain.UserConsent) error {
	return c.db.Create(consent).Error
}
