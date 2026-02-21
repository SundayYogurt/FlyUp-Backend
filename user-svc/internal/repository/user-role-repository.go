package repository

import (
	"errors"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type UserRoleRepository interface {
	ReplaceUserRoles(userID uint, roleIDs []uint) error
	GetRolesByUserID(userID uint) ([]domain.Role, error)
	UserHasRole(userID uint, roleCode string) (bool, error)
}

type userRoleRepository struct {
	db *gorm.DB
}

func NewUserRoleRepository(db *gorm.DB) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (ur *userRoleRepository) ReplaceUserRoles(userID uint, roleIDs []uint) error {
	if userID == 0 {
		return errors.New("invalid user_id")
	}

	return ur.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&domain.UserRole{}).Error; err != nil {
			return err
		}

		if len(roleIDs) == 0 {
			return nil
		}

		links := make([]domain.UserRole, 0, len(roleIDs))
		for _, rid := range roleIDs {
			links = append(links, domain.UserRole{
				UserID: userID,
				RoleID: rid,
			})
		}
		return tx.Create(&links).Error
	})
}

func (ur *userRoleRepository) GetRolesByUserID(userID uint) ([]domain.Role, error) {
	var roles []domain.Role
	err := ur.db.
		Model(&domain.Role{}).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	return roles, nil
}

func (ur *userRoleRepository) UserHasRole(userID uint, roleCode string) (bool, error) {
	var count int64
	err := ur.db.
		Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.code = ?", userID, roleCode).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
