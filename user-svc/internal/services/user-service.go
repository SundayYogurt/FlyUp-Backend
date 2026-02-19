package services

import (
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/interfaces"
	"github.com/SundayYogurt/user_service/internal/repository"
	"github.com/SundayYogurt/user_service/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type UserService interface {
	// Auth
	Register(input dto.UserSignup) error
	Login(email, password string) (domain.User, error)
	Authenticate(c *fiber.Ctx) (*domain.User, error)
	ForgotPassword(email string) error
	SetPassword(token, newPassword string) error

	// Profile
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error)

	// Role & Status
	SetStatus(userID uint, status string) error
	SetRoles(userID uint, roles []string) error

	// Pioneer Verification
	SubmitPioneerVerification(userID uint, input dto.PioneerVerificationRequest) error
	ApprovePioneer(userID uint, adminID uint, note string) error
	RejectPioneer(userID uint, adminID uint, reason string) error
	ListPendingPioneerVerifications(limit, offset int) ([]dto.PioneerVerificationResponse, error)

	// Booster Verification
	SubmitKYC(userID uint, input dto.KYCRequest) error
	GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error)

	// Authorization for investment
	EnsureUserCanInvest(userID uint) error
}

type userService struct {
	repo           repository.UserRepository
	producer       *interfaces.ProducerHandler
	kycRepo        repository.KYCRepository
	studentRepo    repository.StudentProfileRepository
	universityRepo repository.UniversityRepository
}

func NewUserService(repo repository.UserRepository, producer *interfaces.ProducerHandler, kycRepo repository.KYCRepository, studentRepo repository.StudentProfileRepository, universityRepo repository.UniversityRepository) UserService {
	return &userService{repo: repo, producer: producer, kycRepo: kycRepo, studentRepo: studentRepo, universityRepo: universityRepo}
}

func (u userService) Register(input dto.UserSignup) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) Login(email, password string) (domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u userService) Authenticate(c *fiber.Ctx) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u userService) ForgotPassword(email string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) SetPassword(token, newPassword string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) GetProfile(userID uint) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u userService) UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (u userService) SetStatus(userID uint, status string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) SetRoles(userID uint, roles []string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) SubmitPioneerVerification(userID uint, input dto.PioneerVerificationRequest) error {

	// หา user
	user, err := u.repo.FindUserById(userID)
	if err != nil {
		return err
	}

	// ดึง domain จาก email
	emailDomain, err := utils.ExtractEmailDomain(user.Email)
	if err != nil {
		return err
	}

	// หา university จาก domain
	university, uniErr := u.universityRepo.FindByDomain(emailDomain)

	// สร้าง/อัปเดต student profile เป็น pending เสมอ
	profile := &domain.StudentProfile{
		UserID:         userID,
		StudentCode:    input.StudentCode,
		Faculty:        input.Faculty,
		Major:          input.Major,
		YearLevel:      input.YearLevel,
		StudentCardURL: input.StudentCardURL,
		VerifyStatus:   domain.StudentVerifyPending, // default = pending
	}

	if uniErr == nil && university != nil {
		profile.UniversityID = university.ID
	}

	return u.studentRepo.Upsert(profile)
}

func (u userService) ApprovePioneer(userID uint, adminID uint, note string) error {
	return u.studentRepo.Approve(userID, adminID, note)
}

func (u userService) RejectPioneer(userID uint, adminID uint, reason string) error {
	return u.studentRepo.Reject(userID, adminID, reason)
}

func (u userService) ListPendingPioneerVerifications(limit, offset int) ([]dto.PioneerVerificationResponse, error) {
	profiles, err := u.studentRepo.ListPending(limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]dto.PioneerVerificationResponse, 0, len(profiles))
	for _, p := range profiles {
		out = append(out, dto.PioneerVerificationResponse{
			UserID:         p.UserID,
			UniversityID:   p.UniversityID,
			StudentCode:    p.StudentCode,
			Faculty:        p.Faculty,
			Major:          p.Major,
			YearLevel:      p.YearLevel,
			StudentCardURL: p.StudentCardURL,
			VerifyStatus:   p.VerifyStatus,
		})
	}
	return out, nil

}

func (u userService) SubmitKYC(userID uint, input dto.KYCRequest) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u userService) EnsureUserCanInvest(userID uint) error {
	//TODO implement me
	panic("implement me")
}
