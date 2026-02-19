package handlers

import (
	"github.com/SundayYogurt/user_service/internal/api/rest/middleware"
	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/SundayYogurt/user_service/pkg/utils"
	"github.com/gofiber/fiber/v2"
)

type UserHandler struct {
	svc services.UserService
}

func NewUserHandler(svc services.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	// =========================
	// USER
	// =========================
	user := api.Group("/user")

	// Auth
	user.Post("/register", h.Register)
	user.Post("/login", h.Login)
	user.Post("/forgot-password", h.ForgotPassword)
	user.Post("/reset-password", h.SetPassword)

	app.Use(middleware.AuthMiddleware())
	// Profile
	user.Get("/me", h.Me) // ใช้ token ใน handler เอง (ถ้ามี)
	user.Get("/profile/:userID", h.GetProfile)
	user.Put("/profile/:userID", h.UpdateProfile)

	// Role & Status
	user.Patch("/:userID/status", h.SetStatus)
	user.Put("/:userID/roles", h.SetRoles)

	// Pioneer Verification
	user.Post("/:userID/pioneer/verify", h.SubmitPioneerVerification)
	user.Post("/:userID/pioneer/approve", h.ApprovePioneer)
	user.Post("/:userID/pioneer/reject", h.RejectPioneer)
	user.Get("/pioneer/pending", h.ListPendingPioneerVerifications)

	// Booster KYC
	user.Post("/:userID/kyc", h.SubmitKYC)
	user.Get("/:userID/kyc/status", h.GetKYCStatus)

	// Investment guard
	user.Get("/:userID/can-invest", h.EnsureUserCanInvest)
}

func (h *UserHandler) Register(ctx *fiber.Ctx) error {
	var requestBody dto.UserSignup

	if err := ctx.BodyParser(&requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	if requestBody.Email == "" || requestBody.Password == "" {
		return utils.ResponseError(ctx, fiber.StatusBadRequest, "Please provide valid inputs")
	}

	if err := h.svc.Register(requestBody); err != nil {
		return utils.ResponseError(ctx, fiber.StatusInternalServerError, err.Error())
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
		return utils.ResponseError(ctx, fiber.StatusUnauthorized, "Invalid email or password")
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

func (h *UserHandler) Me(ctx *fiber.Ctx) error {

}

func (h *UserHandler) GetProfile(ctx *fiber.Ctx) error {

}

func (h *UserHandler) UpdateProfile(ctx *fiber.Ctx) error {

}

func (h *UserHandler) SetStatus(ctx *fiber.Ctx) error {

}

func (h *UserHandler) SetRoles(ctx *fiber.Ctx) error {

}

func (h *UserHandler) SubmitPioneerVerification(ctx *fiber.Ctx) error {

}

func (h *UserHandler) ApprovePioneer(ctx *fiber.Ctx) error {

}

func (h *UserHandler) RejectPioneer(ctx *fiber.Ctx) error {

}

func (h *UserHandler) ListPendingPioneerVerifications(ctx *fiber.Ctx) error {

}

func (h *UserHandler) SubmitKYC(ctx *fiber.Ctx) error {

}

func (h *UserHandler) GetKYCStatus(ctx *fiber.Ctx) error {

}

func (h *UserHandler) EnsureUserCanInvest(ctx *fiber.Ctx) error {

}
