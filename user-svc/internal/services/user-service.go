package services

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/interfaces"
	"github.com/SundayYogurt/user_service/internal/repository"
	"github.com/SundayYogurt/user_service/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	// Auth
	Register(input dto.RegisterRequest) error
	Login(email, password string) (*domain.User, error)
	Authenticate(c *fiber.Ctx) (*domain.User, error)
	ForgotPassword(email string) error
	SetPassword(token, newPassword string) error
	VerifyEmail(token string) error

	// Profile
	GetProfile(userID uint) (*domain.User, error)
	UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error)

	// Admin: Role & Status
	SetStatus(userID uint, status string) error
	SetRoles(userID uint, roles []string) error
	IsAdmin(userID uint) (bool, error)
	CreateUniversity(adminID uint, input dto.UniversityCreateRequest) error

	// Pioneer Verification
	SubmitPioneerVerification(userID uint, input dto.PioneerInput) error
	ApprovePioneer(userID uint, adminID uint, note string) error
	RejectPioneer(userID uint, adminID uint, reason string) error
	ListPendingPioneerVerifications(limit, offset int) ([]dto.PioneerResponse, error)

	// Booster Verification (KYC)
	SubmitKYC(userID uint, input dto.KYCRequest) error
	GetKYCStatus(userID uint) (*dto.KYCStatusResponse, error)

	// Authorization for investment
	EnsureUserCanInvest(userID uint, projectOwnerID uint) error
}

type userService struct {
	// user
	repo repository.UserRepository

	// roles
	roleRepo     repository.RoleRepository
	userRoleRepo repository.UserRoleRepository

	// pioneer (student profile)
	studentRepo    repository.StudentProfileRepository
	universityRepo repository.UniversityRepository

	// kyc
	kycRepo repository.KYCRepository

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
) UserService {
	return &userService{
		repo:           repo,
		producer:       producer,
		kycRepo:        kycRepo,
		studentRepo:    studentRepo,
		universityRepo: universityRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
	}
}

/*
Register
- สร้าง user
- assign role
- ถ้า PIONEER -> สร้าง student_profile (pending) ทันที (เหมือน submit pioneer)
- สร้าง verify-email token แล้วส่ง event (ถ้ามี producer)
*/
func (u *userService) Register(input dto.RegisterRequest) error {
	// normalize
	email := strings.TrimSpace(strings.ToLower(input.Email))
	displayName := strings.TrimSpace(input.DisplayName)
	role := strings.TrimSpace(strings.ToUpper(input.Role))

	if email == "" || input.Password == "" || displayName == "" {
		return errors.New("invalid inputs")
	}
	if role != "BOOSTER" && role != "PIONEER" {
		return errors.New("invalid role")
	}

	// role-specific: pioneer ต้องมีข้อมูล
	if role == "PIONEER" {
		if input.Pioneer == nil {
			return errors.New("pioneer data is required")
		}
		// validate pioneer ขั้นต่ำ
		if strings.TrimSpace(input.Pioneer.StudentCode) == "" ||
			strings.TrimSpace(input.Pioneer.Faculty) == "" ||
			strings.TrimSpace(input.Pioneer.Major) == "" ||
			strings.TrimSpace(input.Pioneer.YearLevel) == "" {
			return errors.New("missing pioneer fields")
		}

		// strict: email domain ต้องอยู่ใน university table
		emailDomain, err := utils.ExtractEmailDomain(email)
		if err != nil {
			return err
		}
		if _, err := u.universityRepo.FindByDomain(emailDomain); err != nil {
			return errors.New("email domain is not associated with any university")
		}
	}

	// check duplicate email
	if existing, err := u.repo.FindUserByEmail(email); err == nil && existing != nil && existing.ID != 0 {
		return errors.New("email already exists")
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("hash password error: %v", err)
		return errors.New("failed to hash password")
	}

	// create user
	usr := &domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		DisplayName:  displayName,
		Phone:        input.Phone,
		Status:       "active",
	}
	if err := u.repo.CreateUser(usr); err != nil {
		log.Printf("create user error: %v", err)
		return errors.New("failed to create user")
	}

	// assign role -> user_roles
	roleObj, err := u.roleRepo.FindByCode(role)
	if err != nil {
		return err
	}
	if err := u.userRoleRepo.ReplaceUserRoles(usr.ID, []uint{roleObj.ID}); err != nil {
		return err
	}

	// ถ้า PIONEER -> submit pending ทันที (ใช้ helper เดียวกัน)
	if role == "PIONEER" {
		if err := u.upsertPioneerPending(usr.ID, usr.Email, *input.Pioneer); err != nil {
			return err
		}
	}

	// email verification token (เก็บ hash, ส่ง plain token)
	plainToken, err := utils.RandomToken(32)
	if err != nil {
		return errors.New("failed to generate verification token")
	}
	tokenHash := utils.Sha256Hex(plainToken)
	exp := time.Now().Add(24 * time.Hour)

	log.Println("VERIFY_EMAIL_TOKEN:", plainToken)

	usr.VerificationToken = tokenHash
	usr.VerificationTokenExpiresAt = &exp
	if err := u.repo.SaveUser(usr); err != nil {
		return err
	}

	// publish event (optional)
	if u.producer != nil {
		payload := fmt.Sprintf(`{"user_id":%d,"email":"%s","token":"%s","expires_at":"%s"}`,
			usr.ID, usr.Email, plainToken, exp.Format(time.RFC3339),
		)
		_ = u.producer.PublishMessage([]byte("user.verify_email"), []byte(payload))
	}

	return nil
}

/*
Helper: สร้าง/อัปเดต pioneer profile เป็น pending
- strict: ต้องหา university ได้จาก email domain
*/
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
		YearLevel:      strings.TrimSpace(p.YearLevel),
		StudentCardURL: p.StudentCardURL,
		VerifyStatus:   domain.StudentVerifyPending,
		UniversityID:   &university.ID,
	}

	return u.studentRepo.Upsert(profile)
}

func (u *userService) Login(email, password string) (*domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))

	user, err := u.repo.FindUserByEmail(email)
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	// 1. ต้อง verify email ก่อน
	if user.EmailVerifiedAt == nil {
		return nil, errors.New("please verify email")
	}

	// 2. check password
	if err := bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// 3. ถ้าเป็น PIONEER → ต้อง approve ก่อน
	isPioneer, _ := u.userRoleRepo.UserHasRole(user.ID, "PIONEER")
	if isPioneer {
		profile, err := u.studentRepo.FindByUserID(user.ID)
		if err != nil || profile == nil {
			return nil, errors.New("pioneer profile not found")
		}

		if profile.VerifyStatus != domain.StudentVerifyApproved {
			return nil, errors.New("pioneer account pending admin approval")
		}
	}

	return user, nil
}

/*
Authenticate
- อ่าน userID จาก middleware ใส่ใน ctx.Locals("userID")
*/
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

/*
VerifyEmail
- client ส่ง token plain มา
- service จะ sha256 แล้วหาใน DB
*/
func (u *userService) VerifyEmail(token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errors.New("token is required")
	}

	hash := utils.Sha256Hex(token)
	user, err := u.repo.FindUserByVerificationTokenHash(hash) // repo ต้องมี
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

/*
ForgotPassword
- สร้าง token plain
- เก็บ hash + expires ใน DB
- ส่ง plain token ไป email service (ถ้ามี)
*/
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

	user.ResetTokenHash = hash
	user.ResetTokenExpiresAt = &exp

	if err := u.repo.SaveUser(user); err != nil {
		return errors.New("fail to save user")
	}

	// publish event (optional)
	if u.producer != nil {
		payload := fmt.Sprintf(`{"user_id":%d,"email":"%s","token":"%s","expires_at":"%s"}`,
			user.ID, user.Email, plain, exp.Format(time.RFC3339),
		)
		_ = u.producer.PublishMessage([]byte("user.reset_password"), []byte(payload))
	}

	return nil
}

/*
SetPassword
- client ส่ง token plain มา
- service sha256 แล้วหา user
*/
func (u *userService) SetPassword(token, newPassword string) error {
	token = strings.TrimSpace(token)
	if token == "" || strings.TrimSpace(newPassword) == "" {
		return errors.New("invalid input")
	}

	hash := utils.Sha256Hex(token)
	user, err := u.repo.FindUserByResetToken(hash) // repo ต้องมี
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

	if err := u.repo.SaveUser(user); err != nil {
		return errors.New("fail to save user")
	}
	return nil
}

/*
Profile
*/
func (u *userService) GetProfile(userID uint) (*domain.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user id")
	}
	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (u *userService) UpdateProfile(userID uint, input dto.UpdateUserProfile) (*domain.User, error) {
	if userID == 0 {
		return nil, errors.New("invalid user_id")
	}

	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	if strings.TrimSpace(input.DisplayName) != "" {
		user.DisplayName = strings.TrimSpace(input.DisplayName)
	}

	// phone เป็น pointer: nil = ไม่แก้, &"" = เคลียร์
	if input.Phone != nil {
		p := strings.TrimSpace(*input.Phone)
		if p == "" {
			user.Phone = nil
		} else {
			user.Phone = &p
		}
	}

	if err := u.repo.SaveUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

/*
Admin: SetStatus
*/
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

/*
Admin: SetRoles
- รับ []string codes เช่น ["ADMIN","BOOSTER"]
*/
func (u *userService) SetRoles(userID uint, roles []string) error {
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

/*
IsAdmin
*/
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

/*
Pioneer: SubmitPioneerVerification
- ใช้กรณี user อยากแก้ข้อมูล/ส่งใหม่หลังสมัคร
*/
func (u *userService) SubmitPioneerVerification(userID uint, input dto.PioneerInput) error {
	if userID == 0 {
		return errors.New("invalid user id")
	}

	user, err := u.repo.FindUserById(userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// reuse helper เดียวกับ Register
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
			YearLevel:      p.YearLevel,
			StudentCardURL: p.StudentCardURL,
			VerifyStatus:   p.VerifyStatus,
		})
	}
	return out, nil
}

/*
KYC: Submit
- สร้าง submission pending + เพิ่ม documents
*/
func (u *userService) SubmitKYC(userID uint, input dto.KYCRequest) error {
	if userID == 0 {
		return errors.New("invalid user id")
	}
	if len(input.Documents) == 0 {
		return errors.New("documents are required")
	}

	sub := &domain.KYCSubmission{
		UserID: userID,
		Status: domain.KYCStatusPending,
	}
	if err := u.kycRepo.CreateSubmission(sub); err != nil {
		return err
	}

	docs := make([]domain.KYCDocument, 0, len(input.Documents))
	for _, d := range input.Documents {
		if strings.TrimSpace(d.DocType) == "" || strings.TrimSpace(d.FileURL) == "" {
			return errors.New("doc_type and file_url are required")
		}
		docs = append(docs, domain.KYCDocument{
			DocType: d.DocType,
			FileURL: d.FileURL,
		})
	}

	return u.kycRepo.AddDocuments(sub.ID, docs)
}

/*
KYC: GetStatus
*/
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
			DocType: d.DocType,
			FileURL: d.FileURL,
		})
	}

	submittedAt := sub.SubmittedAt.Format(time.RFC3339)

	var reviewedAt *string
	if sub.ReviewedAt != nil {
		s := sub.ReviewedAt.Format(time.RFC3339)
		reviewedAt = &s
	}

	var reviewNote *string
	if sub.Review != nil && strings.TrimSpace(sub.Review.Note) != "" {
		n := sub.Review.Note
		reviewNote = &n
	}

	return &dto.KYCStatusResponse{
		KYCID:       sub.ID,
		Status:      sub.Status,
		SubmittedAt: submittedAt,
		ReviewedAt:  reviewedAt,
		ReviewNote:  reviewNote,
		Documents:   docOut,
	}, nil
}

/*
EnsureUserCanInvest
- user ต้อง active
- KYC ต้อง approved
- role ต้องเป็น BOOSTER หรือ PIONEER
- pioneer ห้ามลงทุนโปรเจกต์ตัวเอง
*/
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
	if kyc.Status != domain.KYCStatusApproved {
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

func (u *userService) CreateUniversity(adminID uint, input dto.UniversityCreateRequest) error {
	if adminID == 0 {
		return errors.New("unauthorized")
	}

	// 2) validate input
	nameTh := strings.TrimSpace(input.NameTH)
	nameEn := strings.TrimSpace(input.NameEN)
	domainStr := strings.ToLower(strings.TrimSpace(input.Domain))
	if nameTh == "" || domainStr == "" || nameEn == "" {
		return errors.New("invalid input")
	}

	//optional) กัน user ใส่ @ มา
	domainStr = strings.TrimPrefix(domainStr, "@")

	// 3) map dto -> domain
	un := &domain.University{
		NameTH: nameTh,
		NameEN: nameEn,
		Domain: domainStr,
	}

	return u.universityRepo.AddUniversity(un)
}
