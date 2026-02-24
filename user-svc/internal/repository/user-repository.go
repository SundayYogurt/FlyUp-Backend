package repository

import (
	"errors"
	"log"

	"github.com/SundayYogurt/user_service/internal/domain"
	"gorm.io/gorm"
)

type UserRepository interface {
	CreateUser(user *domain.User) (*domain.User, error)
	FindUserByEmail(email string) (*domain.User, error)
	SaveUser(user *domain.User) error
	FindUserByResetToken(token string) (*domain.User, error)
	FindUserById(userID uint) (*domain.User, error)
	FindUserByVerificationTokenHash(hash string) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, errors.New("nil user")
	}

	if err := r.db.Create(user).Error; err != nil {
		log.Printf("create user error: %v", err)
		return nil, errors.New("failed to create user")
	}

	return user, nil
}

func (r *userRepository) FindUserByEmail(email string) (*domain.User, error) {
	user := &domain.User{}

	if err := r.db.First(user, "email = ?", email).Error; err != nil {
		// แนะนำ: แยก not found ออก ถ้า service อยากรู้
		// if errors.Is(err, gorm.ErrRecordNotFound) { return nil, err }
		log.Printf("find user by email error: %v", err)
		return nil, errors.New("failed to find user by email")
	}

	return user, nil
}

func (r *userRepository) SaveUser(user *domain.User) error {
	if user == nil {
		return errors.New("nil user")
	}

	if err := r.db.Save(user).Error; err != nil {
		log.Printf("save user error: %v", err)
		return errors.New("failed to save user")
	}
	return nil
}

func (r *userRepository) FindUserByResetToken(token string) (*domain.User, error) {
	user := &domain.User{}

	if err := r.db.Where("reset_token_hash = ?", token).First(user).Error; err != nil {
		log.Printf("find user by reset token error: %v", err)
		return nil, errors.New("failed to find user by reset token")
	}

	return user, nil
}

func (r *userRepository) FindUserById(userID uint) (*domain.User, error) {
	user := &domain.User{}

	if err := r.db.First(user, userID).Error; err != nil {
		log.Printf("find user by id error: %v", err)
		return nil, errors.New("failed to find user by ID")
	}

	return user, nil
}

func (r *userRepository) FindUserByVerificationTokenHash(hash string) (*domain.User, error) {
	user := &domain.User{}

	if err := r.db.Where("verification_token = ?", hash).First(user).Error; err != nil {
		log.Printf("find user by verification token error: %v", err)
		return nil, errors.New("failed to find user by verification token")
	}

	return user, nil
}
