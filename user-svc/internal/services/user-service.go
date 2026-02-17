package services

import (
	"context"
	"go-flyup/internal/domain"
	"go-flyup/internal/dto"
)

package services

import (
"context"

"go-flyup/internal/domain"
"go-flyup/internal/dto"
)

type UserService interface {
	// =====================
	// Auth
	// =====================
	Register(ctx context.Context, in dto.UserSignup) (*domain.User, error)

	// Authenticate = ยืนยันตัวตน (verify email+password) คืน user
	Authenticate(ctx context.Context, email, password string) (*domain.User, error)

	// Login = Authenticate + ออก token
	Login(ctx context.Context, email, password string) (*dto.AuthTokens, *domain.User, error)

	// password reset
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// optional (แนะนำถ้าทำ refresh token)
	RefreshToken(ctx context.Context, refreshToken string) (*dto.AuthTokens, error)
	Logout(ctx context.Context, refreshToken string) error

	// =====================
	// Profile (Identity)
	// =====================
	CreateProfile(ctx context.Context, userID string, in dto.UserProfile) (*domain.User, error)
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
	UpdateProfile(ctx context.Context, userID string, in dto.UpdateProfile) (*domain.User, error)

	// =====================
	// Admin
	// =====================
	SetStatus(ctx context.Context, userID string, status string) error           // active/suspended/deleted
	SetRoles(ctx context.Context, userID string, roles []string) error          // ADMIN/STUDENT/BOOSTER
}
