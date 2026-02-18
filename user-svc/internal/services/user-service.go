package services

import (
	"context"
	"go-flyup/internal/domain"
	"go-flyup/internal/dto"

	"github.com/SundayYogurt/user_service/internal/interfaces"
	"github.com/SundayYogurt/user_service/internal/repository"
	"github.com/gofiber/fiber/v2"
)

package services

import (
"context"

"go-flyup/internal/dto"
)

type UserService interface {
	// Auth
	Register(input dto.UserSignup) error
	Login(email, password string) (domain.User, error)
	Authenticate(c *fiber.Ctx) (*domain.User, error)
	ForgotPassword(email string) error
	ResetPassword(token, newPassword string) error

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
	ListPendingPioneerVerifications() ([]dto.PioneerVerificationResponse, error)

	// Booster Verification
	SubmitKYC(userID uint, input dto.KYCRequest) error
	GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error)

	// Authorization for investment
	EnsureUserCanInvest(userID uint) error
}



type userService struct {
	repo repository.UserRepository
	producer *interfaces.ProducerHandler
}

func NewUserService(repo repository.UserRepository, producer *interfaces.ProducerHandler) UserService {
	return &userService{repo: repo, producer: producer}
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

func (u userService) ResetPassword(token, newPassword string) error {
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
	//TODO implement me
	panic("implement me")
}

func (u userService) ApprovePioneer(userID uint, adminID uint, note string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) RejectPioneer(userID uint, adminID uint, reason string) error {
	//TODO implement me
	panic("implement me")
}

func (u userService) ListPendingPioneerVerifications() ([]dto.PioneerVerificationResponse, error) {
	//TODO implement me
	panic("implement me")
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


