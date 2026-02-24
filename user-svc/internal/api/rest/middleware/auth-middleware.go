package middleware

import (
	"strings"

	"github.com/SundayYogurt/user_service/internal/helper"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(auth helper.Auth) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 1) try cookie first
		tokenStr := strings.TrimSpace(ctx.Cookies("access_token"))

		// 2) fallback to Authorization header
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(ctx.Get("Authorization"))
		}

		user, err := auth.VerifyToken(tokenStr) // จะปรับ VerifyToken ให้รับได้ 2 แบบ
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		ctx.Locals("userID", uint(user.UserID))
		ctx.Locals("user", user)
		return ctx.Next()
	}
}

func AdminOnly(userSvc services.UserService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(uint)
		if !ok || userID == 0 {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		isAdmin, err := userSvc.IsAdmin(userID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if !isAdmin {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "admin only",
			})
		}

		return ctx.Next()
	}
}

func PioneerOnly(userSvc services.UserService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(uint)
		if !ok || userID == 0 {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		isPioneer, err := userSvc.IsPioneer(userID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if !isPioneer {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "pioneer only",
			})
		}

		return ctx.Next()
	}
}

func BoosterOnly(userSvc services.UserService) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userID, ok := ctx.Locals("userID").(uint)
		if !ok || userID == 0 {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		isBooster, err := userSvc.IsBooster(userID)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if !isBooster {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "booster only",
			})
		}

		return ctx.Next()
	}
}
