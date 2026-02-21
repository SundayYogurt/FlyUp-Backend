package repository

import (
	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type UniversityRepository interface {
	FindByID(id uint) (*domain.University, error)
	FindByDomain(domain string) (*domain.University, error)
	List(limit, offset int) ([]domain.University, error)
	AddUniversity(university *domain.University) error
}

type universityRepository struct {
	db *gorm.DB
}

func NewUniversityRepository(db *gorm.DB) UniversityRepository {
	return &universityRepository{db: db}
}

func (u *universityRepository) FindByID(id uint) (*domain.University, error) {
	var university domain.University
	if err := u.db.First(&university, id).Error; err != nil {
		return nil, err
	}
	return &university, nil
}

func (u *universityRepository) FindByDomain(emailDomain string) (*domain.University, error) {
	var university domain.University
	err := u.db.First(&university, "domain = ?", emailDomain).Error

	if err != nil {
		return nil, err
	}

	return &university, nil
}

func (u *universityRepository) List(limit, offset int) ([]domain.University, error) {
	var universities []domain.University

	err := u.db.Order("name_th ASC").Limit(limit).Offset(offset).Find(&universities).Error

	if err != nil {
		return nil, err
	}
	return universities, nil
}

func (u *universityRepository) AddUniversity(university *domain.University) error {
	return u.db.Create(university).Error
}
