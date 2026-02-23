package handlers

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/api/rest/middleware"
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/SundayYogurt/user_service/pkg/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	svc services.UserService
	cld *cloudinary.Cloudinary
}

func NewUserHandler(svc services.UserService, cld *cloudinary.Cloudinary) *UserHandler {
	return &UserHandler{svc: svc, cld: cld}
}

// Routes
func (h *UserHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1/users")

	// public
	api.Post("/register", h.Register)
	api.Post("/login", h.Login)
	api.Post("/forgot-password", h.ForgotPassword)
	api.Post("/reset-password", h.SetPassword)
	api.Post("/verify-email", h.VerifyEmail)

	// protected
	auth := api.Group("/", middleware.AuthMiddleware())
	auth.Get("/me", h.Me)
	auth.Patch("/profile/:userID", h.UpdateProfile)

	// booster kyc
	booster := auth.Group("/booster", middleware.BoosterOnly(h.svc))
	booster.Post("/kyc/submit", h.SubmitKYCMultipart)
	booster.Get("/kyc/status", h.GetMyKYCStatus)

	// pioneer
	pioneer := auth.Group("/pioneer", middleware.PioneerOnly(h.svc))

	pioneer.Post("/verify", h.SubmitPioneerVerification)
	pioneer.Post("/uploads/student-card", h.UploadStudentCard)

	// admin only
	admin := auth.Group("/admin", middleware.AdminOnly(h.svc))
	admin.Patch("/:userID/status", h.SetStatus)
	admin.Put("/:userID/roles", h.SetRoles)
	admin.Post("/:userID/pioneer/approve", h.ApprovePioneer)
	admin.Post("/:userID/pioneer/reject", h.RejectPioneer)
	admin.Get("/pioneer/pending", h.ListPendingPioneerVerifications)
	admin.Post("/universities/create", h.CreateUniversity)
}

// ========================
// Auth
// ========================

// Register godoc
// @Summary Register a new user
// @Description Register user (BOOSTER or PIONEER). Pioneer will be pending verification.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.RegisterRequest true "Register payload"
// @Success 201 {object} dto.APISuccessAny
// @Failure 400 {object} dto.APIError
// @Router /api/v1/users/register [post]
func (h *UserHandler) Register(c *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}
	if err := h.svc.Register(req); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}

	// ข้อความให้ FE แยกกรณี pioneer ได้
	if strings.ToUpper(strings.TrimSpace(req.Role)) == "PIONEER" {
		return utils.ResponseSuccess(c, 201, fiber.Map{
			"message": "registered. please verify email and wait for admin approval",
			"status":  "PENDING_VERIFICATION",
		})
	}

	return utils.ResponseSuccess(c, 201, "registered. please verify email")
}

// Login godoc
// @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.UserLogin true "Login payload"
// @Success 200 {object} dto.APISuccessLogin
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/login [post]
func (h *UserHandler) Login(c *fiber.Ctx) error {
	var req dto.UserLogin
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}

	user, err := h.svc.Login(req)
	if err != nil {
		return utils.ResponseError(c, 401, err.Error())
	}

	token, err := middleware.GenerateToken(int(user.ID), user.Email)
	if err != nil {
		return utils.ResponseError(c, 500, "could not generate token")
	}

	//set cookie
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token,
		HTTPOnly: true,
		SameSite: fiber.CookieSameSiteNoneMode,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		Secure:   true,
		// Secure: true, // เปิดเมื่อเป็น https
	})

	return utils.ResponseSuccess(c, 200, dto.LoginResponse{
		Token: token,
		User: dto.UserProfileResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Status:    user.Status,
			Phone:     user.Phone,
			CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// ForgotPassword godoc
// @Summary Request password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.ForgotPasswordRequest true "Email payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Router /api/v1/users/forgot-password [post]
func (h *UserHandler) ForgotPassword(c *fiber.Ctx) error {
	var req dto.ForgotPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid email")
	}

	if err := h.svc.ForgotPassword(req.Email); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "reset link sent")
}

// SetPassword godoc
// @Summary Reset password using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.SetPasswordRequest true "Set password payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Router /api/v1/users/reset-password [post]
func (h *UserHandler) SetPassword(c *fiber.Ctx) error {
	var req dto.SetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid json body (check Content-Type: application/json)")
	}

	if strings.TrimSpace(req.Token) == "" || strings.TrimSpace(req.NewPassword) == "" {
		return utils.ResponseError(c, 400, "token and new_password are required")
	}

	if err := h.svc.SetPassword(req); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}

	return utils.ResponseSuccess(c, 200, "password updated")
}

// VerifyEmail godoc
// @Summary Verify email using token
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body dto.VerifyEmailRequest true "Verify email payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Router /api/v1/users/verify-email [post]
func (h *UserHandler) VerifyEmail(c *fiber.Ctx) error {
	var req dto.VerifyEmailRequest
	if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		return utils.ResponseError(c, 400, "token required")
	}

	if err := h.svc.VerifyEmail(req.Token); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "email verified")
}

// ========================
// Profile
// ========================

// Me godoc
// @Summary Get my profile
// @Tags Profile
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APISuccessAny
// @Failure 401 {object} dto.APIError
// @Failure 404 {object} dto.APIError
// @Router /api/v1/users/me [get]
func (h *UserHandler) Me(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	user, err := h.svc.GetProfile(userID)
	if err != nil {
		return utils.ResponseError(c, 404, err.Error())
	}
	return utils.ResponseSuccess(c, 200, fiber.Map{"user": user})
}

// UpdateProfile godoc
// @Summary Update my profile
// @Tags Profile
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.UpdateUserProfile true "Update profile payload"
// @Success 200 {object} dto.APISuccessAny
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Failure 403 {object} dto.APIError
// @Router /api/v1/users/profile/{userID} [patch]
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	authID, ok := c.Locals("userID").(uint)
	if !ok || authID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	if authID != targetID {
		return utils.ResponseError(c, 403, "forbidden")
	}

	var req dto.UpdateUserProfile
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}

	user, err := h.svc.UpdateProfile(targetID, req)
	if err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, fiber.Map{"user": user})
}

// ========================
// Pioneer
// ========================

// SubmitPioneerVerification godoc
// @Summary Submit pioneer verification
// @Tags Pioneer
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.PioneerInput true "Pioneer verification payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Failure 403 {object} dto.APIError
// @Router /api/v1/users/pioneer/{userID}/verify [post]
func (h *UserHandler) SubmitPioneerVerification(c *fiber.Ctx) error {
	authID, ok := c.Locals("userID").(uint)
	if !ok || authID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	if authID != targetID {
		return utils.ResponseError(c, 403, "forbidden")
	}

	var req dto.PioneerInput
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}

	if strings.TrimSpace(req.StudentCode) == "" ||
		strings.TrimSpace(req.Faculty) == "" ||
		strings.TrimSpace(req.Major) == "" ||
		req.StudentCardURL == nil || strings.TrimSpace(*req.StudentCardURL) == "" {
		return utils.ResponseError(c, 400, "missing required fields")
	}

	if err := h.svc.SubmitPioneerVerification(targetID, req); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}

	return utils.ResponseSuccess(c, 200, "pioneer verification submitted")
}

// UploadStudentCard godoc
// @Summary Upload student card image
// @Tags Pioneer
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Student card image (jpg/jpeg/png/webp, max 5MB)"
// @Success 200 {object} dto.APISuccessAny
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Failure 500 {object} dto.APIError
// @Router /api/v1/users/pioneer/uploads/student-card [post]
func (h *UserHandler) UploadStudentCard(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return utils.ResponseError(c, 400, "file required")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return utils.ResponseError(c, 400, "only jpg/jpeg/png/webp allowed")
	}

	const maxSize = 5 * 1024 * 1024
	if file.Size > maxSize {
		return utils.ResponseError(c, 400, "file too large (max 5MB)")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ResponseError(c, 400, "cannot open file")
	}
	defer src.Close()

	res, err := h.cld.Upload.Upload(
		context.Background(),
		src,
		uploader.UploadParams{Folder: "flyup/student_cards"},
	)
	if err != nil || res.SecureURL == "" {
		return utils.ResponseError(c, 500, "upload failed")
	}

	return utils.ResponseSuccess(c, 200, fiber.Map{
		"url":       res.SecureURL,
		"public_id": res.PublicID,
	})
}

// ========================
// KYC
// ========================

// SubmitKYCMultipart godoc
// @Summary Submit KYC (multipart)
// @Description Upload id_front + selfie, optional others[]
// @Tags KYC
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param id_front formData file true "ID card front image (jpg/jpeg/png/webp)"
// @Param selfie formData file true "Selfie image (jpg/jpeg/png/webp)"
// @Param others formData file false "Other documents (can send multiple)"
// @Success 200 {object} dto.APISuccessAny
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/booster/kyc/submit [post]
func (h *UserHandler) SubmitKYCMultipart(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	// id_front (required)
	idFile, err := c.FormFile("id_front")
	if err != nil || idFile == nil {
		return utils.ResponseError(c, 400, "id_front required")
	}
	if !isAllowedImageExt(idFile.Filename) {
		return utils.ResponseError(c, 400, "id_front must be jpg/jpeg/png/webp")
	}

	idSrc, err := idFile.Open()
	if err != nil {
		return utils.ResponseError(c, 400, "cannot open id_front")
	}
	defer idSrc.Close()

	idBytes, err := utils.ReadAllLimit(idSrc, 12*1024*1024)
	if err != nil {
		return utils.ResponseError(c, 400, "id_front too large or invalid")
	}

	// selfie (REQUIRED)
	sf, err := c.FormFile("selfie")
	if err != nil || sf == nil {
		return utils.ResponseError(c, 400, "selfie required")
	}
	if !isAllowedImageExt(sf.Filename) {
		return utils.ResponseError(c, 400, "selfie must be jpg/jpeg/png/webp")
	}

	sSrc, err := sf.Open()
	if err != nil {
		return utils.ResponseError(c, 400, "cannot open selfie")
	}
	defer sSrc.Close()

	sb, err := utils.ReadAllLimit(sSrc, 12*1024*1024)
	if err != nil {
		return utils.ResponseError(c, 400, "selfie too large or invalid")
	}

	selfie := &dto.FileBytes{Filename: sf.Filename, Bytes: sb}

	// others (optional multiple)
	form, _ := c.MultipartForm()
	others := make([]dto.TypedFile, 0)

	if form != nil && form.File != nil {
		if files, ok := form.File["others"]; ok && len(files) > 0 {
			for _, f := range files {
				if f == nil {
					continue
				}
				if !isAllowedImageExt(f.Filename) {
					return utils.ResponseError(c, 400, "others must be jpg/jpeg/png/webp")
				}

				src, err := f.Open()
				if err != nil {
					return utils.ResponseError(c, 400, "cannot open others file")
				}

				b, rerr := utils.ReadAllLimit(src, 12*1024*1024)
				_ = src.Close()

				if rerr != nil {
					return utils.ResponseError(c, 400, "others file too large or invalid")
				}

				others = append(others, dto.TypedFile{
					DocType:  string(domain.KYCDocTypeOther), // หรือ "other"
					Filename: f.Filename,
					Bytes:    b,
				})
			}
		}
	}

	input := dto.KYCSubmitFiles{
		IDFront: dto.FileBytes{Filename: idFile.Filename, Bytes: idBytes},
		Selfie:  selfie,
		Others:  others,
	}

	resp, err := h.svc.SubmitKYCMultipart(c.Context(), userID, input)
	if err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}

	return utils.ResponseSuccess(c, 200, resp)
}

// GetMyKYCStatus godoc
// @Summary Get my KYC status
// @Tags KYC
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.APISuccessAny
// @Failure 401 {object} dto.APIError
// @Failure 404 {object} dto.APIError
// @Router /api/v1/users/booster/kyc/status [get]
func (h *UserHandler) GetMyKYCStatus(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(uint)
	if !ok || userID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	status, err := h.svc.GetKYCStatus(userID)
	if err != nil {
		return utils.ResponseError(c, 404, err.Error())
	}

	return utils.ResponseSuccess(c, 200, status)
}

// ========================
// Admin
// ========================

// SetStatus godoc
// @Summary Set user status (admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.SetStatusRequest true "Status payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/admin/{userID}/status [patch]
func (h *UserHandler) SetStatus(c *fiber.Ctx) error {
	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	var req dto.SetStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}
	if strings.TrimSpace(req.Status) == "" {
		return utils.ResponseError(c, 400, "status required")
	}

	if err := h.svc.SetStatus(targetID, req.Status); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "status updated")
}

// SetRoles godoc
// @Summary Set user roles (admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.SetRolesRequest true "Roles payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/admin/{userID}/roles [put]
func (h *UserHandler) SetRoles(c *fiber.Ctx) error {
	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	var req dto.SetRolesRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}
	if len(req.Roles) == 0 {
		return utils.ResponseError(c, 400, "roles required")
	}

	seen := map[string]bool{}
	clean := make([]string, 0, len(req.Roles))
	for _, r := range req.Roles {
		r = strings.TrimSpace(strings.ToUpper(r))
		if r == "" || seen[r] {
			continue
		}
		seen[r] = true
		clean = append(clean, r)
	}
	if len(clean) == 0 {
		return utils.ResponseError(c, 400, "roles required")
	}
	req.Roles = clean

	if err := h.svc.SetRoles(targetID, req); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "roles updated")
}

// ApprovePioneer godoc
// @Summary Approve pioneer (admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.ApprovePioneerRequest false "Optional note"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/admin/{userID}/pioneer/approve [post]
func (h *UserHandler) ApprovePioneer(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	var req dto.ApprovePioneerRequest
	_ = c.BodyParser(&req) // optional

	if err := h.svc.ApprovePioneer(targetID, adminID, strings.TrimSpace(req.Note)); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "pioneer approved")
}

// RejectPioneer godoc
// @Summary Reject pioneer (admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param userID path int true "User ID"
// @Param payload body dto.RejectPioneerRequest true "Reject reason"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/admin/{userID}/pioneer/reject [post]
func (h *UserHandler) RejectPioneer(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	paramID, err := c.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(c, 400, "invalid user id")
	}
	targetID := uint(paramID)

	var req dto.RejectPioneerRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return utils.ResponseError(c, 400, "reason required")
	}

	if err := h.svc.RejectPioneer(targetID, adminID, strings.TrimSpace(req.Reason)); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "pioneer rejected")
}

// ListPendingPioneerVerifications godoc
// @Summary List pending pioneer verifications (admin)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit (max 200)" default(20)
// @Param offset query int false "Offset" default(0)
// @Success 200 {object} dto.APISuccessAny
// @Failure 401 {object} dto.APIError
// @Failure 500 {object} dto.APIError
// @Router /api/v1/users/admin/pioneer/pending [get]
func (h *UserHandler) ListPendingPioneerVerifications(c *fiber.Ctx) error {
	limit := 20
	offset := 0

	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.svc.ListPendingPioneerVerifications(limit, offset)
	if err != nil {
		return utils.ResponseError(c, 500, err.Error())
	}

	return utils.ResponseSuccess(c, 200, fiber.Map{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateUniversity godoc
// @Summary Create university (admin)
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body dto.UniversityCreateRequest true "University payload"
// @Success 200 {object} dto.APISuccessString
// @Failure 400 {object} dto.APIError
// @Failure 401 {object} dto.APIError
// @Router /api/v1/users/admin/universities/create [post]
func (h *UserHandler) CreateUniversity(c *fiber.Ctx) error {
	adminID, ok := c.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(c, 401, "unauthorized")
	}

	var req dto.UniversityCreateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ResponseError(c, 400, "invalid input")
	}

	if err := h.svc.CreateUniversity(adminID, req); err != nil {
		return utils.ResponseError(c, 400, err.Error())
	}
	return utils.ResponseSuccess(c, 200, "university created")
}

// ========================
// Helpers
// ========================

func isAllowedImageExt(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		return true
	default:
		return false
	}
}
