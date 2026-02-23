package repository

import (
	"errors"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByCode(code string) (*domain.Role, error)
	FindByCodes(codes []string) ([]domain.Role, error)
	List(limit, offset int) ([]domain.Role, error)
	GetRoleCodeByUserID(userID uint) (string, error)
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByCode(code string) (*domain.Role, error) {
	var role domain.Role
	if err := r.db.Where("code = ?", code).First(&role).Error; err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByCodes(codes []string) ([]domain.Role, error) {
	var roles []domain.Role
	if len(codes) == 0 {
		return roles, nil
	}
	if err := r.db.Where("code IN ?", codes).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) List(limit, offset int) ([]domain.Role, error) {
	var roles []domain.Role
	if err := r.db.Order("id ASC").Limit(limit).Offset(offset).Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *roleRepository) GetRoleCodeByUserID(userID uint) (string, error) {
	var roleCode string

	err := r.db.
		Table("user_roles").
		Select("roles.code").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Limit(1).
		Scan(&roleCode).Error

	if err != nil {
		return "", err
	}
	if roleCode == "" {
		return "", errors.New("role not found for user")
	}
	return roleCode, nil
}
