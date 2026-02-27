package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/clients/iapp"
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/helper"
	"github.com/SundayYogurt/user_service/internal/helper/utils"
	"github.com/SundayYogurt/user_service/internal/interfaces"
	"github.com/SundayYogurt/user_service/internal/repository"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	// Auth
	Register(input dto.RegisterRequest) error
	Login(input dto.UserLogin) (*domain.User, error)
	Authenticate(c *fiber.Ctx) (*domain.User, error)
	ForgotPassword(email string) error
	SetPassword(input dto.SetPasswordRequest) error
	VerifyEmail(token string) error

	// Profile
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error)

	// Admin: Role & Status
	SetStatus(userID uint, status string) error
	SetRoles(userID uint, input dto.SetRolesRequest) error
	IsAdmin(userID uint) (bool, error)

	CreateUniversity(adminID uint, input dto.UniversityCreateRequest) error

	// Pioneer Verification
	SubmitPioneerVerification(userID uint, input dto.PioneerInput) error
	ApprovePioneer(userID uint, adminID uint, note string) error
	RejectPioneer(userID uint, adminID uint, reason string) error
	ListPendingPioneerVerifications(limit, offset int) ([]dto.PioneerResponse, error)
	IsPioneer(userID uint) (bool, error)

	// Booster Verification (KYC)
	SubmitKYCMultipart(ctx context.Context, userID uint, input dto.KYCSubmitFiles) (*dto.KYCSubmitResponse, error)
	GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error)
	IsBooster(userID uint) (bool, error)

	// Admin KYC list
	ListPendingKYC(limit, offset int) ([]dto.PendingKYCResponse, error)

	// Authorization for investment
	EnsureUserCanInvest(userID uint, projectOwnerID uint) error
}

type userService struct {
	// user
	repo     repository.UserRepository
	bankRepo repository.BankAccountRepository
	auth     helper.Auth
	// roles
	roleRepo     repository.RoleRepository
	userRoleRepo repository.UserRoleRepository

	// pioneer (student profile)
	studentRepo    repository.StudentProfileRepository
	universityRepo repository.UniversityRepository

	// kyc
	kycRepo     repository.KYCRepository
	iapp        *iapp.Client
	uploader    interfaces.Uploader
	consentRepo repository.ConsentRepository

	// messaging
	producer interfaces.ProducerHandler
}

func NewUserService(
	repo repository.UserRepository,
	producer interfaces.ProducerHandler,
	kycRepo repository.KYCRepository,
	studentRepo repository.StudentProfileRepository,
	universityRepo repository.UniversityRepository,
	roleRepo repository.RoleRepository,
	userRoleRepo repository.UserRoleRepository,
	bankRepo repository.BankAccountRepository,
	iappClient *iapp.Client,
	uploader interfaces.Uploader,
	consentRepo repository.ConsentRepository,
	auth helper.Auth,
) UserService {
	return &userService{
		repo:           repo,
		producer:       producer,
		kycRepo:        kycRepo,
		studentRepo:    studentRepo,
		universityRepo: universityRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		bankRepo:       bankRepo,
		iapp:           iappClient,
		uploader:       uploader,
		consentRepo:    consentRepo,
		auth:           auth,
	}
}

// AUTH
func (u *userService) Register(input dto.RegisterRequest) error {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)
	phone := strings.TrimSpace(input.Phone)
	role := strings.TrimSpace(strings.ToUpper(input.Role))

	if email == "" || strings.TrimSpace(input.Password) == "" || firstName == "" || lastName == "" || phone == "" {
		return errors.New("invalid inputs")
	}
	if role != "BOOSTER" && role != "PIONEER" {
		return errors.New("invalid role")
	}

	if !helper.HasConsent(input.ConsentCodes, domain.ConsentTerms) {
		return errors.New("must accept terms")
	}

	// pioneer: ตรวจ domain ตอนสมัคร (ยืนยันตัวตนค่อยทำทีหลัง)
	if role == "PIONEER" {
		emailDomain, err := utils.ExtractEmailDomain(email)
		if err != nil {
			return err
		}
		if _, err := u.universityRepo.FindByDomain(emailDomain); err != nil {
			return errors.New("email domain is not associated with any university")
		}
	}

	// duplicate email (repo คืน *domain.User)
	existing, err := u.repo.FindUserByEmail(email)
	if err == nil && existing != nil && existing.ID != 0 {
		return errors.New("email already exists")
	}

	if len(input.Password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	// หมายเหตุ: ถ้า err != nil อาจเป็น "not found" ก็ปล่อยผ่านได้
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	newUser := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FirstName:    firstName,
		LastName:     lastName,
		Phone:        phone,
		Status:       "active",
	}

	usr, err := u.repo.CreateUser(newUser)
	if err != nil {
		return err
	}
	if usr == nil || usr.ID == 0 {
		return errors.New("failed to create user")
	}

	now := time.Now()
	for _, code := range input.ConsentCodes {
		c := &domain.UserConsent{
			UserID:      usr.ID,
			ConsentCode: strings.ToUpper(code),
			Accepted:    true,
			AcceptedAt:  &now,
		}
		if err := u.consentRepo.CreateConsent(c); err != nil {
			if helper.IsDuplicateConsent(err) {
				continue
			}
			return err
		}
	}

	// assign role
	roleObj, err := u.roleRepo.FindByCode(role)
	if err != nil {
		return err
	}
	if err := u.userRoleRepo.ReplaceUserRoles(usr.ID, []uint{roleObj.ID}); err != nil {
		return err
	}

	// email verification token
	plainToken, err := utils.RandomToken(32)
	log.Printf("Verification token: %s", plainToken)
	if err != nil {
		return errors.New("failed to generate verification token")
	}
	tokenHash := utils.Sha256Hex(plainToken)
	exp := time.Now().Add(24 * time.Hour)

	usr.VerificationToken = tokenHash
	usr.VerificationTokenExpiresAt = &exp

	if err := u.repo.SaveUser(usr); err != nil {
		return err
	}

	// publish event (optional)
	if u.producer != nil {
		payload := fmt.Sprintf(
			`{"user_id":%d,"email":"%s","token":"%s","expires_at":"%s"}`,
			usr.ID, usr.Email, plainToken, exp.Format(time.RFC3339),
		)
		_ = u.producer.PublishMessage([]byte("user.verify_email"), []byte(payload))
	}

	return nil
}

func (u *userService) Login(input dto.UserLogin) (*domain.User, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	password := strings.TrimSpace(input.Password)

	if email == "" || password == "" {
		return nil, errors.New("invalid email or password")
	}

	user, err := u.repo.FindUserByEmail(email)
	if err != nil || user == nil || user.ID == 0 {
		return nil, errors.New("invalid email or password")
	}

	if user.EmailVerifiedAt == nil {
		return nil, errors.New("please verify email")
	}

	// ถ้าคุณมี status
	if user.Status != "" && user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	if err := u.auth.VerifyPassword(password, user.PasswordHash); err != nil {
		return nil, errors.New("invalid email or password")
	}

	return user, nil
}

func (u *userService) Authenticate(c *fiber.Ctx) (*domain.User, error) {
	v := c.Locals("userID")
	userID, ok := v.(uint)
	if !ok || userID == 0 {
		return nil, errors.New("unauthorized")
	}
	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (u *userService) VerifyEmail(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token is required")
	}

	hash := utils.Sha256Hex(token)
	user, err := u.repo.FindUserByVerificationTokenHash(hash)
	if err != nil || user == nil {
		return errors.New("invalid token")
	}

	if user.VerificationTokenExpiresAt == nil || time.Now().After(*user.VerificationTokenExpiresAt) {
		return errors.New("token expired")
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.VerificationToken = ""
	user.VerificationTokenExpiresAt = nil
	return u.repo.SaveUser(user)
}

func (u *userService) ForgotPassword(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := u.repo.FindUserByEmail(email)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	plain, err := utils.RandomToken(32)
	if err != nil {
		return errors.New("failed to generate reset token")
	}

	hash := utils.Sha256Hex(plain)
	exp := time.Now().Add(30 * time.Minute)

	// อย่า log ค่า user.ResetTokenHash เดิม เพราะทำให้สับสน และไม่ควร log token/hash ใน production
	log.Printf("Reset token (dev only): %s", plain)

	user.ResetTokenHash = hash
	user.ResetTokenExpiresAt = &exp
	if err := u.repo.SaveUser(user); err != nil {
		return errors.New("fail to save user")
	}

	if u.producer != nil {
		payload := fmt.Sprintf(`{"user_id":%d,"email":"%s","token":"%s","expires_at":"%s"}`,
			user.ID, user.Email, plain, exp.Format(time.RFC3339),
		)
		_ = u.producer.PublishMessage([]byte("user.reset_password"), []byte(payload))
	}

	return nil
}

func (u *userService) SetPassword(input dto.SetPasswordRequest) error {
	token := strings.TrimSpace(input.Token)
	newPassword := strings.TrimSpace(input.NewPassword)

	if token == "" || newPassword == "" {
		return errors.New("invalid input")
	}

	if len(input.NewPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	hash := utils.Sha256Hex(token)
	user, err := u.repo.FindUserByResetToken(hash)
	if err != nil || user == nil {
		return errors.New("invalid or expired token")
	}

	if user.ResetTokenExpiresAt == nil || time.Now().After(*user.ResetTokenExpiresAt) {
		return errors.New("invalid or expired token")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("fail to hash password")
	}

	user.PasswordHash = string(hashedPassword)
	user.ResetTokenHash = ""
	user.ResetTokenExpiresAt = nil

	return u.repo.SaveUser(user)
}

// Profile
func (u *userService) UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user_id")
	}

	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	//PATCH: user fields
	if input.FirstName != nil {
		fn := strings.TrimSpace(*input.FirstName)
		if fn == "" {
			return nil, errors.New("first_name cannot be empty")
		}
		user.FirstName = fn
	}

	if input.LastName != nil {
		ln := strings.TrimSpace(*input.LastName)
		if ln == "" {
			return nil, errors.New("last_name cannot be empty")
		}
		user.LastName = ln
	}

	if input.Phone != nil {
		p := strings.TrimSpace(*input.Phone)
		if p == "" {
			return nil, errors.New("phone cannot be empty")
		}
		user.Phone = p
	}

	// ---------- PATCH: bank account (optional) ----------
	if input.BankAccount != nil {
		// allow only BOOSTER
		roleCode, err := u.roleRepo.GetRoleCodeByUserID(userID)
		if err != nil {
			return nil, err
		}
		roleCode = strings.ToUpper(strings.TrimSpace(roleCode))
		if roleCode != "BOOSTER" {
			return nil, errors.New("only booster can update bank account")
		}

		ba := input.BankAccount
		bankCode := strings.ToUpper(strings.TrimSpace(ba.BankCode))
		accountNo := strings.TrimSpace(ba.AccountNo)
		accountName := strings.TrimSpace(ba.AccountName)

		if bankCode == "" || accountNo == "" || accountName == "" {
			return nil, errors.New("missing bank_account fields")
		}

		// สร้างบัญชีใหม่ (เพราะ repo ยังไม่มี update)
		acc := &domain.UserBankAccount{
			UserID:      userID,
			BankCode:    bankCode,
			AccountNo:   accountNo,
			AccountName: accountName,
			Status:      "active",
			IsDefault:   false,
		}

		if err := u.bankRepo.Create(acc); err != nil {
			return nil, err
		}

		// ถ้าขอให้เป็น default
		if ba.IsDefault {
			if err := u.bankRepo.SetDefault(userID, acc.ID); err != nil {
				return nil, err
			}
		}
	}

	// save user
	if err := u.repo.SaveUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

//ADMIN: STATUS / ROLES / UNIVERSITY

func (u *userService) SetStatus(userID uint, status string) error {
	if userID == 0 {
		return errors.New("invalid user_id")
	}

	status = strings.TrimSpace(strings.ToLower(status))
	switch status {
	case "active", "suspended", "deleted":
	default:
		return errors.New("invalid status")
	}

	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	user.Status = status
	return u.repo.SaveUser(user)
}

func (u *userService) SetRoles(userID uint, input dto.SetRolesRequest) error {
	roles := input.Roles

	if userID == 0 {
		return errors.New("invalid user id")
	}
	if len(roles) == 0 {
		return errors.New("roles are required")
	}

	var roleIDs []uint
	for _, code := range roles {
		code = strings.TrimSpace(strings.ToUpper(code))
		if code == "" {
			continue
		}
		role, err := u.roleRepo.FindByCode(code)
		if err != nil {
			return err
		}
		roleIDs = append(roleIDs, role.ID)
	}
	if len(roleIDs) == 0 {
		return errors.New("roles are required")
	}

	return u.userRoleRepo.ReplaceUserRoles(userID, roleIDs)
}

func (u *userService) IsAdmin(userID uint) (bool, error) {
	if userID == 0 {
		return false, errors.New("invalid user id")
	}

	roles, err := u.userRoleRepo.GetRolesByUserID(userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if strings.ToUpper(r.Code) == "ADMIN" {
			return true, nil
		}
	}
	return false, nil
}

func (u *userService) IsPioneer(userID uint) (bool, error) {
	if userID == 0 {
		return false, errors.New("invalid user id")
	}

	roles, err := u.userRoleRepo.GetRolesByUserID(userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if strings.ToUpper(r.Code) == "PIONEER" {
			return true, nil
		}
	}
	return false, nil
}

func (u *userService) IsBooster(userID uint) (bool, error) {
	if userID == 0 {
		return false, errors.New("invalid user id")
	}

	roles, err := u.userRoleRepo.GetRolesByUserID(userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if strings.ToUpper(r.Code) == "BOOSTER" {
			return true, nil
		}
	}
	return false, nil
}

func (u *userService) CreateUniversity(adminID uint, input dto.UniversityCreateRequest) error {
	if adminID == 0 {
		return errors.New("unauthorized")
	}

	nameTh := strings.TrimSpace(input.NameTH)
	nameEn := strings.TrimSpace(input.NameEN)
	domainStr := strings.ToLower(strings.TrimSpace(input.Domain))

	if nameTh == "" || nameEn == "" || domainStr == "" {
		return errors.New("invalid input")
	}

	domainStr = strings.TrimPrefix(domainStr, "@")

	un := &domain.University{
		NameTH: nameTh,
		NameEN: nameEn,
		Domain: domainStr,
	}
	return u.universityRepo.AddUniversity(un)
}

func (u *userService) SubmitPioneerVerification(userID uint, input dto.PioneerInput) error {
	if userID == 0 {
		return errors.New("invalid user id")
	}

	// 1) must be pioneer
	role, err := u.roleRepo.GetRoleCodeByUserID(userID)
	if err != nil {
		return errors.New("cannot get role code")
	}
	role = strings.ToUpper(strings.TrimSpace(role))
	if role != "PIONEER" {
		return errors.New("pioneer only")
	}

	// 2) validate required consents for verification step
	if !helper.HasConsent(input.ConsentCodes,
		domain.ConsentInfoTrue,
	) {
		return errors.New("must accept pioneer Info consent")
	}
	if !helper.HasConsent(input.ConsentCodes,
		domain.ConsentPioneerTerms,
	) {
		return errors.New("must accept pioneer terms consent")
	}

	// 3) load user (need email)
	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// 4) persist consents (idempotent)
	now := time.Now()
	for _, code := range input.ConsentCodes {
		code = strings.ToUpper(strings.TrimSpace(code))
		if code == "" {
			continue
		}

		// บันทึกเฉพาะ 2 อันที่เกี่ยวกับขั้น verify (กันคนส่งมาเยอะเกิน)
		if code != domain.ConsentInfoTrue && code != domain.ConsentPioneerTerms {
			continue
		}

		c := &domain.UserConsent{
			UserID:      userID,
			ConsentCode: code,
			Accepted:    true,
			AcceptedAt:  &now,
		}

		if err := u.consentRepo.CreateConsent(c); err != nil {
			if helper.IsDuplicateConsent(err) {
				continue // เคยกดยอมรับแล้ว
			}
			return err
		}
	}

	// 5) upsert pioneer verification as pending
	return u.upsertPioneerPending(userID, user.Email, input)
}

func (u *userService) ApprovePioneer(userID uint, adminID uint, note string) error {
	return u.studentRepo.Approve(userID, adminID, strings.TrimSpace(note))
}

func (u *userService) RejectPioneer(userID uint, adminID uint, reason string) error {
	return u.studentRepo.Reject(userID, adminID, strings.TrimSpace(reason))
}

func (u *userService) ListPendingPioneerVerifications(limit, offset int) ([]dto.PioneerResponse, error) {
	profiles, err := u.studentRepo.ListPending(limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]dto.PioneerResponse, 0, len(profiles))
	for _, p := range profiles {
		out = append(out, dto.PioneerResponse{
			UserID:         p.UserID,
			UniversityID:   p.UniversityID,
			StudentCode:    p.StudentCode,
			Faculty:        p.Faculty,
			Major:          p.Major,
			StudentCardURL: p.StudentCardURL,
			VerifyStatus:   p.VerifyStatus,
		})
	}
	return out, nil
}

func (u *userService) upsertPioneerPending(userID uint, email string, p dto.PioneerInput) error {
	emailDomain, err := utils.ExtractEmailDomain(email)
	if err != nil {
		return err
	}

	university, err := u.universityRepo.FindByDomain(emailDomain)
	if err != nil || university == nil {
		return errors.New("email domain is not associated with any university")
	}

	profile := &domain.StudentProfile{
		UserID:         userID,
		StudentCode:    strings.TrimSpace(p.StudentCode),
		Faculty:        strings.TrimSpace(p.Faculty),
		Major:          strings.TrimSpace(p.Major),
		StudentCardURL: p.StudentCardURL,
		VerifyStatus:   domain.StudentVerifyPending,
		UniversityID:   &university.ID,
	}

	return u.studentRepo.Upsert(profile)
}

/* =========================
   KYC (BOOSTER)
========================= */

func (u *userService) SubmitKYCMultipart(
	ctx context.Context,
	userID uint,
	input dto.KYCSubmitFiles,
) (*dto.KYCSubmitResponse, error) {
	const (
		faceThreshold = 0.75
		maxWidth      = 1200
		jpgQuality    = 85
		maxFileSize   = 12 * 1024 * 1024 // 12MB
	)

	// deps
	if u.kycRepo == nil {
		return nil, errors.New("kyc repo is not configured")
	}
	if u.iapp == nil {
		return nil, errors.New("iapp is not configured")
	}
	if u.uploader == nil {
		return nil, errors.New("uploader is not configured")
	}

	// validate
	if userID == 0 {
		return nil, errors.New("invalid user_id")
	}
	if strings.TrimSpace(input.IDFront.Filename) == "" || len(input.IDFront.Bytes) == 0 {
		return nil, errors.New("id_front is required")
	}

	role, err := u.roleRepo.GetRoleCodeByUserID(userID)
	if err != nil {
		return nil, errors.New("cannot get role code")
	}

	role = strings.ToUpper(strings.TrimSpace(role))
	if role != "BOOSTER" {
		return nil, errors.New("only booster can submit kyc")
	}

	if len(input.IDFront.Bytes) > maxFileSize {
		return nil, errors.New("id_front size is too large")
	}
	if input.Selfie != nil && len(input.Selfie.Bytes) > maxFileSize {
		return nil, errors.New("selfie size is too large")
	}

	if input.Selfie == nil || len(input.Selfie.Bytes) == 0 {
		return nil, errors.New("selfie is required")
	}

	for i, f := range input.Others {
		if strings.TrimSpace(f.Filename) == "" || len(f.Bytes) == 0 {
			return nil, fmt.Errorf("other file #%d is invalid", i+1)
		}
		if len(f.Bytes) > maxFileSize {
			return nil, fmt.Errorf("other file #%d size is too large", i+1)
		}
		docType := strings.TrimSpace(strings.ToLower(f.DocType))
		if docType == "" {
			return nil, fmt.Errorf("other file #%d doc_type is required", i+1)
		}
		// others ต้องเป็น "other" เท่านั้น
		if docType != string(domain.KYCDocTypeOther) {
			return nil, errors.New("invalid doc_type for others (use 'other')")
		}
	}

	// guard: ห้ามส่งซ้ำถ้ายัง pending อยู่
	if latest, err := u.kycRepo.FindLatestByUserID(userID); err == nil && latest != nil {
		if latest.Status == domain.KYCStatusPending {
			return nil, errors.New("kyc already pending admin review")
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// =========================
	// 1) Normalize images (EXIF + resize + JPG)
	// =========================
	idNorm, err := utils.NormalizeToJPG(input.IDFront.Bytes, maxWidth, jpgQuality)
	if err != nil {
		return nil, fmt.Errorf("normalize id_front failed: %w", err)
	}
	input.IDFront.Bytes = idNorm
	input.IDFront.Filename = "id_front.jpg"

	if input.Selfie != nil && len(input.Selfie.Bytes) > 0 {
		selfieNorm, err := utils.NormalizeToJPG(input.Selfie.Bytes, maxWidth, jpgQuality)
		if err != nil {
			return nil, fmt.Errorf("normalize selfie failed: %w", err)
		}
		input.Selfie.Bytes = selfieNorm
		input.Selfie.Filename = "selfie.jpg"
	}

	for i := range input.Others {
		norm, err := utils.NormalizeToJPG(input.Others[i].Bytes, maxWidth, jpgQuality)
		if err != nil {
			return nil, fmt.Errorf("normalize others #%d failed: %w", i+1, err)
		}
		input.Others[i].Bytes = norm
		input.Others[i].Filename = fmt.Sprintf("other_%d.jpg", i+1)
	}

	// =========================
	// 2) Upload files FIRST
	// =========================
	idFrontURL, err := u.uploader.UploadBytes(ctx, "kyc/id-front", input.IDFront.Filename, input.IDFront.Bytes)
	if err != nil {
		return nil, fmt.Errorf("upload id_front failed: %w", err)
	}

	var selfieURL string
	if input.Selfie != nil && len(input.Selfie.Bytes) > 0 {
		selfieURL, err = u.uploader.UploadBytes(ctx, "kyc/selfie", input.Selfie.Filename, input.Selfie.Bytes)
		if err != nil {
			return nil, fmt.Errorf("upload selfie failed: %w", err)
		}
	}

	otherDocs := make([]domain.KYCDocument, 0, len(input.Others))
	for i, f := range input.Others {
		url, upErr := u.uploader.UploadBytes(ctx, "kyc/other", f.Filename, f.Bytes)
		if upErr != nil {
			return nil, fmt.Errorf("upload other #%d failed: %w", i+1, upErr)
		}
		otherDocs = append(otherDocs, domain.KYCDocument{
			DocType: domain.KYCDocTypeOther,
			FileURL: url,
		})
	}

	// 3) Prepare submission + call iApp face match

	ocrProvider := "iapp"
	sub := &domain.KYCSubmission{
		UserID:      userID,
		Status:      domain.KYCStatusPending,
		OCRProvider: &ocrProvider,
		// FaceMatchScore/Passed/OCRError จะเติมด้านล่าง
		AutoApproved: false,
	}

	// ถ้ามี selfie ค่อยเรียก face match
	if input.Selfie != nil && len(input.Selfie.Bytes) > 0 {
		faceRes, faceErr := u.iapp.VerifyFaceAndIDCard(
			ctx,
			input.IDFront.Filename, bytes.NewReader(input.IDFront.Bytes),
			input.Selfie.Filename, bytes.NewReader(input.Selfie.Bytes),
		)

		if faceRes != nil {
			score := faceRes.Total.Confidence
			passed := strings.ToLower(faceRes.Total.IsSamePerson) == "true"
			sub.FaceMatchScore = &score
			sub.FaceMatchPassed = &passed
		}

		// iApp error -> เก็บไว้ (ไม่ให้ submit ล่ม)
		if faceErr != nil {
			msg := faceErr.Error()
			sub.OCRError = &msg
		}
	} else {
		// ไม่มี selfie -> pending ให้แอดมินตรวจ
		msg := "selfie not provided"
		sub.OCRError = &msg
	}

	// 4) Auto-approve logic

	faceOK := sub.FaceMatchScore != nil &&
		sub.FaceMatchPassed != nil &&
		*sub.FaceMatchScore >= faceThreshold &&
		*sub.FaceMatchPassed

	if faceOK && sub.OCRError == nil {
		sub.Status = domain.KYCStatusAutoApproved
		sub.AutoApproved = true
	} else {
		sub.Status = domain.KYCStatusPending
		sub.AutoApproved = false

		// ถ้าไม่มี error แต่ไม่ผ่าน threshold ใส่เหตุผลช่วย debug
		if sub.OCRError == nil {
			msg := fmt.Sprintf("face match below threshold %.2f", faceThreshold)
			sub.OCRError = &msg
		}
	}

	// 5) Documents + Create (TX)

	docs := make([]domain.KYCDocument, 0, 1+1+len(otherDocs))
	docs = append(docs, domain.KYCDocument{
		DocType: domain.KYCDocTypeIDCard,
		FileURL: idFrontURL,
	})
	if selfieURL != "" {
		docs = append(docs, domain.KYCDocument{
			DocType: domain.KYCDocTypeSelfie,
			FileURL: selfieURL,
		})
	}
	docs = append(docs, otherDocs...)

	if err := u.kycRepo.CreateSubmissionWithDocuments(sub, docs); err != nil {
		return nil, err
	}

	// 6) Response

	return &dto.KYCSubmitResponse{
		KYCID:  sub.ID,
		Status: string(sub.Status),

		OCRProvider:  sub.OCRProvider,
		AutoApproved: sub.AutoApproved,
	}, nil
}

func (u *userService) GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error) {
	if userID == 0 {
		return nil, errors.New("invalid user_id")
	}

	sub, err := u.kycRepo.FindLatestByUserID(userID)
	if err != nil || sub == nil {
		return nil, errors.New("kyc not found")
	}

	docOut := make([]dto.KYCDocumentResponse, 0, len(sub.Documents))
	for _, d := range sub.Documents {
		docOut = append(docOut, dto.KYCDocumentResponse{
			ID:      d.ID,
			DocType: string(d.DocType),
			FileURL: d.FileURL,
		})
	}

	submittedAt := sub.CreatedAt.Format(time.RFC3339)

	var reviewedAt *string
	if sub.ReviewedAt != nil {
		s := sub.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &s
	}

	var reviewNote *string
	if sub.Review != nil && sub.Review.Note != nil && strings.TrimSpace(*sub.Review.Note) != "" {
		n := strings.TrimSpace(*sub.Review.Note)
		reviewNote = &n
	}

	return &dto.KYCStatusResponse{
		KYCID:       sub.ID,
		Status:      string(sub.Status),
		SubmittedAt: submittedAt,
		ReviewedAt:  reviewedAt,
		ReviewNote:  reviewNote,
		Documents:   docOut,
	}, nil
}

func (u *userService) ListPendingKYC(limit, offset int) ([]dto.PendingKYCResponse, error) {
	if u.kycRepo == nil {
		return nil, errors.New("kyc repo is not configured")
	}

	subs, err := u.kycRepo.ListPending(limit, offset)
	if err != nil {
		return nil, err
	}

	out := make([]dto.PendingKYCResponse, 0, len(subs))
	for _, s := range subs {
		out = append(out, dto.PendingKYCResponse{
			KYCID:       s.ID,
			UserID:      s.UserID,
			Status:      string(s.Status),
			SubmittedAt: s.CreatedAt.Format(time.RFC3339),
		})
	}
	return out, nil
}

/* =========================
   INVEST PERMISSION
========================= */

func (u *userService) EnsureUserCanInvest(userID uint, projectOwnerID uint) error {
	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}
	if user.Status != "active" {
		return errors.New("user not active")
	}

	kyc, err := u.kycRepo.FindLatestByUserID(userID)
	if err != nil || kyc == nil {
		return errors.New("kyc not found")
	}
	if kyc.Status != domain.KYCStatusApproved && kyc.Status != domain.KYCStatusAutoApproved {
		return errors.New("kyc not approved")
	}

	roles, err := u.userRoleRepo.GetRolesByUserID(userID)
	if err != nil {
		return err
	}

	isBooster := false
	isPioneer := false
	for _, r := range roles {
		switch strings.ToUpper(r.Code) {
		case "BOOSTER":
			isBooster = true
		case "PIONEER":
			isPioneer = true
		}
	}

	if !isBooster && !isPioneer {
		return errors.New("user cannot invest")
	}
	if isPioneer && userID == projectOwnerID {
		return errors.New("pioneer cannot invest in own project")
	}
	return nil
}

func (u *userService) GetProfile(userID uint) (*domain.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user id")
	}

	user, err := u.repo.FindUserById(userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
