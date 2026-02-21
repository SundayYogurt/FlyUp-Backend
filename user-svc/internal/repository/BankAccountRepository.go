package repository

import (
	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type BankAccountRepository interface {
	// user actions
	Create(account *domain.UserBankAccount) error
	FindByID(id uint) (*domain.UserBankAccount, error)
	ListByUserID(userID uint) ([]domain.UserBankAccount, error)
	SetDefault(userID uint, accountID uint) error
	Disable(userID uint, accountID uint) error
}

type bankAccountRepository struct {
	db *gorm.DB
}

func NewBankAccountRepository(db *gorm.DB) BankAccountRepository {
	return &bankAccountRepository{db: db}
}

func (b *bankAccountRepository) Create(account *domain.UserBankAccount) error {
	return b.db.Create(account).Error
}

func (s *bankAccountRepository) FindByID(id uint) (*domain.UserBankAccount, error) {
	var account domain.UserBankAccount
	if err := s.db.First(&account, id).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

func (b *bankAccountRepository) ListByUserID(userID uint) ([]domain.UserBankAccount, error) {
	var accounts []domain.UserBankAccount
	if err := b.db.Where("user_id = ?", userID).Order("is_default DESC, created_at ASC").Find(&accounts).Error; err != nil {
		return nil, err
	}

	return accounts, nil
}

func (b *bankAccountRepository) SetDefault(userID uint, accountID uint) error {
	return b.db.Transaction(func(tx *gorm.DB) error {
		// reset default
		if err := tx.Model(&domain.UserBankAccount{}).Where("user_id = ?", userID).Update("is_default", false).Error; err != nil {
			return err
		}
		// ตั้ง default account
		res := tx.Model(&domain.UserBankAccount{}).Where("id = ? AND user_id = ?", accountID, userID).Update("is_default", true)

		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return nil
	})
}

func (b *bankAccountRepository) Disable(userID uint, accountID uint) error {
	res := b.db.Model(&domain.UserBankAccount{}).
		Where("id = ? AND user_id = ?", accountID, userID).
		Updates(map[string]any{
			"status":     "disabled",
			"is_default": false,
		})

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
