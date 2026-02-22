package repository

import (
	"errors"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type KYCRepository interface {
	CreateSubmission(sub *domain.KYCSubmission) error
	AddDocuments(kycID uint, docs []domain.KYCDocument) error

	// แนะนำเพิ่ม: ทำให้ transaction-safe
	CreateSubmissionWithDocuments(sub *domain.KYCSubmission, docs []domain.KYCDocument) error

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

// สร้าง submission "ใหม่จริง" (ไม่ใช้ FirstOrCreate)
func (k *kycRepository) CreateSubmission(sub *domain.KYCSubmission) error {
	if sub == nil {
		return errors.New("nil submission")
	}
	return k.db.Create(sub).Error
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

// สำคัญ: กัน orphan (create + docs ใน tx เดียว)
func (k *kycRepository) CreateSubmissionWithDocuments(sub *domain.KYCSubmission, docs []domain.KYCDocument) error {
	if sub == nil {
		return errors.New("nil submission")
	}

	return k.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(sub).Error; err != nil {
			return err
		}

		if len(docs) == 0 {
			return errors.New("documents are required")
		}
		for i := range docs {
			docs[i].KYCID = sub.ID
		}
		return tx.Create(&docs).Error
	})
}

func (k *kycRepository) FindLatestByUserID(userID uint) (*domain.KYCSubmission, error) {
	var sub domain.KYCSubmission
	err := k.db.
		Preload("Documents").
		Preload("Review").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&sub).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
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
	q := k.db.
		Where("status = ?", domain.KYCStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Offset(offset).
		Preload("Documents").
		Preload("Review")

	if err := q.Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}

func (k *kycRepository) Approve(kycID uint, adminID uint, note string) error {
	now := time.Now()
	note = strings.TrimSpace(note)
	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	return k.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.KYCSubmission{}).
			Where("id = ? AND status = ?", kycID, domain.KYCStatusPending).
			Updates(map[string]any{
				"status":      domain.KYCStatusApproved,
				"reviewed_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		// upsert review (kyc_id unique)
		var existing domain.KYCReview
		err := tx.Where("kyc_id = ?", kycID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&domain.KYCReview{
				KYCID:      kycID,
				ReviewedBy: adminID,
				Decision:   domain.KYCDecisionApproved,
				Note:       notePtr,
				ReviewedAt: now,
			}).Error
		}
		if err != nil {
			return err
		}

		return tx.Model(&existing).Updates(map[string]any{
			"reviewed_by": adminID,
			"decision":    domain.KYCDecisionApproved,
			"note":        notePtr, // set nil ได้จริง
			"reviewed_at": now,
		}).Error
	})
}

func (k *kycRepository) Reject(kycID uint, adminID uint, reason string) error {
	now := time.Now()
	reason = strings.TrimSpace(reason)
	var reasonPtr *string
	if reason != "" {
		reasonPtr = &reason
	}

	return k.db.Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&domain.KYCSubmission{}).
			Where("id = ? AND status = ?", kycID, domain.KYCStatusPending).
			Updates(map[string]any{
				"status":      domain.KYCStatusRejected,
				"reviewed_at": now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var existing domain.KYCReview
		err := tx.Where("kyc_id = ?", kycID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&domain.KYCReview{
				KYCID:      kycID,
				ReviewedBy: adminID,
				Decision:   domain.KYCDecisionRejected,
				Note:       reasonPtr,
				ReviewedAt: now,
			}).Error
		}
		if err != nil {
			return err
		}

		return tx.Model(&existing).Updates(map[string]any{
			"reviewed_by": adminID,
			"decision":    domain.KYCDecisionRejected,
			"note":        reasonPtr,
			"reviewed_at": now,
		}).Error
	})
}
