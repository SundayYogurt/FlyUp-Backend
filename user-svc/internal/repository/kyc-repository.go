package repository

import (
	"time"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type KYCRepository interface {
	CreateSubmission(sub *domain.KYCSubmission) error
	AddDocuments(kycID uint, docs []domain.KYCDocument) error

	FindLatestByUserID(userID uint) (*domain.KYCSubmission, error)
	FindByID(kycID uint) (*domain.KYCSubmission, error)
	ListPending(limit, offset int) ([]domain.KYCSubmission, error)

	Approve(kycID uint, adminID uint, note string) error
	Reject(kycID uint, adminID uint, reason string) error
}

type kycRepository struct {
	db *gorm.DB
}

func NewKYCRepository(db *gorm.DB) KYCRepository {
	return &kycRepository{db: db}
}

func (k *kycRepository) CreateSubmission(sub *domain.KYCSubmission) error {
	return k.db.Where("user_id = ? AND status = ?", sub.UserID, domain.KYCStatusPending).FirstOrCreate(sub).Error
}

func (k *kycRepository) AddDocuments(kycID uint, docs []domain.KYCDocument) error {
	if len(docs) == 0 {
		return nil
	}

	for i := range docs {
		docs[i].KYCID = kycID
	}

	return k.db.Create(&docs).Error
}

func (k *kycRepository) FindLatestByUserID(userID uint) (*domain.KYCSubmission, error) {
	var sub domain.KYCSubmission
	err := k.db.
		Preload("Documents").
		Preload("Review").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (k *kycRepository) FindByID(kycID uint) (*domain.KYCSubmission, error) {
	var sub domain.KYCSubmission
	err := k.db.
		Preload("Documents").
		Preload("Review").
		First(&sub, kycID).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (k *kycRepository) ListPending(limit, offset int) ([]domain.KYCSubmission, error) {
	var subs []domain.KYCSubmission

	err := k.db.Where("status = ?", domain.KYCStatusPending).Order("created_at ASC").Limit(limit).Offset(offset).Find(&subs).Error

	if err != nil {
		return nil, err
	}
	return subs, nil
}

func (k *kycRepository) Approve(kycID uint, adminID uint, note string) error {
	now := time.Now()

	return k.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.KYCSubmission{}).
			Where("id = ? AND status = ?", kycID, domain.KYCStatusPending).
			Updates(map[string]any{
				"status":      domain.KYCStatusApproved,
				"reviewed_by": adminID,
				"reviewed_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound // หรือ custom error ว่าไม่ใช่ pending แล้ว
		}

		review := &domain.KYCReview{
			KYCID:      kycID,
			ReviewedBy: adminID,
			Decision:   domain.KYCDecisionApproved,
			Note:       note,
			ReviewedAt: now,
		}
		return tx.Create(review).Error
	})
}

func (k *kycRepository) Reject(kycID uint, adminID uint, reason string) error {
	now := time.Now()

	return k.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.KYCSubmission{}).
			Where("id = ? AND status = ?", kycID, domain.KYCStatusPending).
			Updates(map[string]any{
				"status":      domain.KYCStatusRejected,
				"reviewed_by": adminID,
				"reviewed_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		review := &domain.KYCReview{
			KYCID:      kycID,
			ReviewedBy: adminID,
			Decision:   domain.KYCDecisionRejected,
			Note:       reason,
			ReviewedAt: now,
		}

		return tx.Create(review).Error
	})
}
