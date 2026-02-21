package handlers

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SundayYogurt/user_service/internal/api/rest/middleware"
	"github.com/SundayYogurt/user_service/internal/clients/iapp"
	"github.com/SundayYogurt/user_service/internal/domain"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/SundayYogurt/user_service/pkg/utils"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	svc  services.UserService
	cld  *cloudinary.Cloudinary
	iapp *iapp.Client
}

func NewUserHandler(svc services.UserService, cld *cloudinary.Cloudinary, iappClient *iapp.Client) *UserHandler {
	return &UserHandler{svc: svc, cld: cld, iapp: iappClient}
}

func (h *UserHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api/v1/users")

	// public
	api.Post("/register", h.Register)
	api.Post("/login", h.Login)
	api.Post("/forgot-password", h.ForgotPassword)
	api.Post("/reset-password", h.SetPassword)
	api.Post("/verify-email", h.VerifyEmail)
	// upload student card (pioneer)
	api.Post("/uploads/student-card", h.UploadStudentCard)
	// protected (ต้องมี token)
	api.Use(middleware.AuthMiddleware())

	api.Get("/me", h.Me)
	api.Put("/profile/:userID", h.UpdateProfile)
	api.Post("/:userID/pioneer/verify", h.SubmitPioneerVerification)
	api.Post("/uploads/kyc", h.UploadKYCFile)
	api.Post("/:userID/kyc", h.SubmitKYC)
	api.Get("/:userID/kyc/status", h.GetKYCStatus)

	// admin-only (ต้องมี token + ต้องเป็น admin)
	api.Use(middleware.AdminOnly(h.svc))

	api.Patch("/:userID/status", h.SetStatus)
	api.Put("/:userID/roles", h.SetRoles)
	api.Post("/:userID/pioneer/approve", h.ApprovePioneer)
	api.Post("/:userID/pioneer/reject", h.RejectPioneer)
	api.Get("/pioneer/pending", h.ListPendingPioneerVerifications)
	api.Post("/universities/create", h.CreateUniversity)
}

func (h *UserHandler) Register(ctx *fiber.Ctx) error {
	var requestBody dto.RegisterRequest
	if err := ctx.BodyParser(&requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	if requestBody.Email == "" || requestBody.Password == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	if err := h.svc.Register(requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	if requestBody.Role == "PIONEER" {
		return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": "Registration successful. Your pioneer account is pending verification.",
			"status":  "PENDING_VERIFICATION",
		})
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "User registered successfully")
}

func (h *UserHandler) Login(ctx *fiber.Ctx) error {
	var requestBody dto.UserLogin
	if err := ctx.BodyParser(&requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "email and password are required")
	}

	user, err := h.svc.Login(requestBody.Email, requestBody.Password)
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, err.Error())
	}

	token, err := middleware.GenerateToken(int(user.ID), user.Email)

	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, "could not generate token")
	}
	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"token": token,
	})
}

func (h *UserHandler) ForgotPassword(ctx *fiber.Ctx) error {
	var requestBody struct {
		Email string `json:"email"`
	}

	if err := ctx.BodyParser(&requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid email id")
	}

	if err := h.svc.ForgotPassword(requestBody.Email); err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "Password reset link sent")
}

func (h *UserHandler) SetPassword(ctx *fiber.Ctx) error {
	var requestBody struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}

	if err := ctx.BodyParser(&requestBody); err != nil || requestBody.Token == "" || requestBody.Password == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid input")
	}

	if err := h.svc.SetPassword(requestBody.Token, requestBody.Password); err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "Password reset successfully")
}

func (h *UserHandler) Authorize(ctx *fiber.Ctx) error {
	user, err := h.svc.Authenticate(ctx)
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}
	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"authenticated": true,
		"user":          user,
	})
}

func (h *UserHandler) Me(ctx *fiber.Ctx) error {
	userID := ctx.Locals("userID").(uint)

	user, err := h.svc.GetProfile(userID)

	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"user": user,
	})
}

func (h *UserHandler) GetProfile(ctx *fiber.Ctx) error {
	paramID, err := ctx.ParamsInt("userID")

	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	profile, err := h.svc.GetProfile(uint(paramID))

	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, profile)
}

func (h *UserHandler) UpdateProfile(ctx *fiber.Ctx) error {
	// auth user from token
	authUserID, ok := ctx.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	// target user from param
	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	// permission: allow only self (เพิ่ม role admin ได้ทีหลัง)
	if targetUserID != authUserID {
		return utils.ResponseError(ctx, fiber.StatusForbidden, "forbidden")
	}

	// parse body
	var req dto.UpdateUserProfile
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	updated, err := h.svc.UpdateProfile(targetUserID, req)
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"user": updated,
	})
}

func (h *UserHandler) SetStatus(ctx *fiber.Ctx) error {
	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	var req struct {
		Status string `json:"status"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}
	if strings.TrimSpace(req.Status) == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "status is required")
	}

	if err := h.svc.SetStatus(targetUserID, req.Status); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "status updated")
}

func (h *UserHandler) SetRoles(ctx *fiber.Ctx) error {
	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	var req struct {
		Roles []string `json:"roles"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}
	if len(req.Roles) == 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "roles are required")
	}

	// normalize roles
	clean := make([]string, 0, len(req.Roles))
	for _, r := range req.Roles {
		r = strings.TrimSpace(strings.ToUpper(r))
		if r != "" {
			clean = append(clean, r)
		}
	}
	if len(clean) == 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "roles are required")
	}

	if err := h.svc.SetRoles(targetUserID, clean); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "roles updated")
}

func (h *UserHandler) SubmitPioneerVerification(ctx *fiber.Ctx) error {
	authUserID, ok := ctx.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	// กันคนอื่นส่งแทน
	if targetUserID != authUserID {
		return utils.ResponseError(ctx, fiber.StatusForbidden, "forbidden")
	}

	var req dto.PioneerInput
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	// validate ขั้นต่ำ (ปรับตาม field dto ของคุณ)
	if strings.TrimSpace(req.StudentCode) == "" ||
		strings.TrimSpace(req.Faculty) == "" ||
		strings.TrimSpace(req.Major) == "" ||
		strings.TrimSpace(req.YearLevel) == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "missing required fields")
	}

	if err := h.svc.SubmitPioneerVerification(targetUserID, req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "pioneer verification submitted")
}

func (h *UserHandler) ApprovePioneer(ctx *fiber.Ctx) error {
	adminID, ok := ctx.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	var req struct {
		Note string `json:"note"`
	}
	_ = ctx.BodyParser(&req) // note optional

	if err := h.svc.ApprovePioneer(targetUserID, adminID, strings.TrimSpace(req.Note)); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "pioneer approved")
}

func (h *UserHandler) RejectPioneer(ctx *fiber.Ctx) error {
	adminID, ok := ctx.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	var req struct {
		Reason string `json:"reason"`
	}
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "reason is required")
	}

	if err := h.svc.RejectPioneer(targetUserID, adminID, strings.TrimSpace(req.Reason)); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "pioneer rejected")
}

func (h *UserHandler) ListPendingPioneerVerifications(ctx *fiber.Ctx) error {
	limit := 20
	offset := 0

	if v := ctx.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := ctx.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	items, err := h.svc.ListPendingPioneerVerifications(limit, offset)
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *UserHandler) SubmitKYC(ctx *fiber.Ctx) error {
	authUserID, ok := ctx.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	if targetUserID != authUserID {
		return utils.ResponseError(ctx, fiber.StatusForbidden, "forbidden")
	}

	var req dto.KYCRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}
	if len(req.Documents) == 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "documents are required")
	}

	// basic validate
	for _, d := range req.Documents {
		if strings.TrimSpace(d.DocType) == "" || strings.TrimSpace(d.FileURL) == "" {
			return utils.ResponseError(ctx, fiber.StatusBadRequest, "doc_type and file_url are required")
		}
	}

	if err := h.svc.SubmitKYC(targetUserID, req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "kyc submitted")
}

func (h *UserHandler) GetKYCStatus(ctx *fiber.Ctx) error {
	authUserID, ok := ctx.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	if targetUserID != authUserID {
		return utils.ResponseError(ctx, fiber.StatusForbidden, "forbidden")
	}

	status, err := h.svc.GetKYCStatus(targetUserID)
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, status)
}

func (h *UserHandler) EnsureUserCanInvest(ctx *fiber.Ctx) error {
	authUserID, ok := ctx.Locals("userID").(uint)
	if !ok || authUserID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	paramID, err := ctx.ParamsInt("userID")
	if err != nil || paramID <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid user id")
	}
	targetUserID := uint(paramID)

	if targetUserID != authUserID {
		return utils.ResponseError(ctx, fiber.StatusForbidden, "forbidden")
	}

	projectOwnerStr := ctx.Query("projectOwnerID")
	if strings.TrimSpace(projectOwnerStr) == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "projectOwnerID is required")
	}
	projectOwnerInt, err := strconv.Atoi(projectOwnerStr)
	if err != nil || projectOwnerInt <= 0 {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid projectOwnerID")
	}
	projectOwnerID := uint(projectOwnerInt)

	if err := h.svc.EnsureUserCanInvest(targetUserID, projectOwnerID); err != nil {
		return utils.ResponseError(ctx, fiber.StatusForbidden, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"can_invest": true,
	})
}

func (h *UserHandler) UploadKYCFile(ctx *fiber.Ctx) error {
	userID, ok := ctx.Locals("userID").(uint)
	if !ok || userID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "file is required")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "only jpg/jpeg/png/webp are allowed")
	}

	// doc_type optional
	docType := strings.TrimSpace(strings.ToLower(ctx.FormValue("doc_type")))
	switch docType {
	case domain.KYCDocTypeIDCard, domain.KYCDocTypeStudentCard, domain.KYCDocTypeSelfie, domain.KYCDocTypeOther:
		// ok
	case "":
		docType = domain.KYCDocTypeOther
	default:
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "invalid doc_type")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "cannot open file")
	}
	defer src.Close()

	uploadRes, err := h.cld.Upload.Upload(
		context.Background(),
		src,
		uploader.UploadParams{
			Folder: "kyc/user_" + strconv.Itoa(int(userID)) + "/" + docType,
		},
	)
	if err != nil || uploadRes.SecureURL == "" {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, "upload failed")
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"url":       uploadRes.SecureURL,
		"public_id": uploadRes.PublicID,
	})
}

func (h *UserHandler) VerifyEmail(ctx *fiber.Ctx) error {
	var req dto.VerifyEmailRequest

	if err := ctx.BodyParser(&req); err != nil || strings.TrimSpace(req.Token) == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "token is required")
	}

	if err := h.svc.VerifyEmail(req.Token); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "email verified")
}

func (h *UserHandler) CreateUniversity(ctx *fiber.Ctx) error {
	adminID, ok := ctx.Locals("userID").(uint)
	if !ok || adminID == 0 {
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "unauthorized")
	}

	var req dto.UniversityCreateRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	if err := h.svc.CreateUniversity(adminID, req); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, err.Error())
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, "university created")
}

func (h *UserHandler) UploadStudentCard(ctx *fiber.Ctx) error {
	file, err := ctx.FormFile("file")
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "file is required")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "only jpg/jpeg/png/webp allowed")
	}

	const maxSize = 5 * 1024 * 1024
	if file.Size > maxSize {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "file too large (max 5MB)")
	}

	src, err := file.Open()
	if err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, "cannot open file")
	}
	defer src.Close()

	uploadRes, err := h.cld.Upload.Upload(
		context.Background(),
		src,
		uploader.UploadParams{
			Folder: "flyup/student_cards",
		},
	)
	if err != nil || uploadRes.SecureURL == "" {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, "upload failed")
	}

	return utils.ResponseSuccess(ctx, fiber.StatusOK, fiber.Map{
		"url":       uploadRes.SecureURL,
		"public_id": uploadRes.PublicID,
	})
}
