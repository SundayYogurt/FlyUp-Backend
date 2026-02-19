package repository

import (
	"time"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type StudentProfileRepository interface {
	Upsert(profile *domain.StudentProfile) error
	FindByUserID(userID uint) (*domain.StudentProfile, error)
	ListPending(limit, offset int) ([]domain.StudentProfile, error)

	Approve(userID uint, adminID uint, note string) error
	Reject(userID uint, adminID uint, reason string) error
}

type studentProfileRepository struct {
	db *gorm.DB
}

func NewStudentProfileRepository(db *gorm.DB) StudentProfileRepository {
	return &studentProfileRepository{db: db}
}

func (s *studentProfileRepository) Upsert(profile *domain.StudentProfile) error {
	return s.db.Where("user_id = ?", profile.UserID).Assign(profile).FirstOrCreate(profile).Error
}

func (s *studentProfileRepository) FindByUserID(userID uint) (*domain.StudentProfile, error) {
	var profile domain.StudentProfile
	if err := s.db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *studentProfileRepository) ListPending(limit, offset int) ([]domain.StudentProfile, error) {
	var profiles []domain.StudentProfile

	err := s.db.Where("verify_status = ?", domain.StudentVerifyPending).Order("created_at ASC").Limit(limit).Offset(offset).Find(&profiles).Error

	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (s *studentProfileRepository) Approve(userID uint, adminID uint, note string) error {
	now := time.Now()

	return s.db.Model(&domain.StudentProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"verify_status": domain.StudentVerifyApproved,
			"verified_at":   now,
			"reviewed_by":   adminID,
			"note":          note,
		}).Error
}

func (s *studentProfileRepository) Reject(userID uint, adminID uint, reason string) error {
	now := time.Now()

	return s.db.Model(&domain.StudentProfile{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"verify_status": domain.StudentVerifyRejected,
			"verified_at":   now,
			"reviewed_by":   adminID,
			"note":          reason,
		}).Error
}
