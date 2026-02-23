package middleware

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/SundayYogurt/user_service/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// 1) try cookie first
		tokenStr := strings.TrimSpace(ctx.Cookies("access_token"))

		// 2) fallback to Authorization header
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(ctx.Get("Authorization"))
		}

		user, err := VerifyToken(tokenStr) // จะปรับ VerifyToken ให้รับได้ 2 แบบ
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

func GenerateToken(userID int, email string) (string, error) {
	if userID == 0 || email == "" {
		return "", errors.New("userID and email are required")
	}

	now := time.Now().Unix()
	exp := time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     now, //int64
		"expiry":  exp, //int64
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func VerifyToken(tokenString string) (dto.AuthResponse, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return dto.AuthResponse{}, errors.New("missing token")
	}

	// support both:
	// - "Bearer <token>"
	// - "<token>"
	if strings.HasPrefix(strings.ToLower(tokenString), "bearer ") {
		parts := strings.SplitN(tokenString, " ", 2)
		if len(parts) != 2 || strings.TrimSpace(parts[1]) == "" {
			return dto.AuthResponse{}, errors.New("invalid token format")
		}
		tokenString = strings.TrimSpace(parts[1])
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(os.Getenv("ACCESS_SECRET")), nil
	})
	if err != nil {
		return dto.AuthResponse{}, errors.New("token parse error")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return dto.AuthResponse{}, errors.New("invalid token claims")
	}

	// safer expiry parse
	expAny, ok := claims["expiry"]
	if !ok {
		return dto.AuthResponse{}, errors.New("missing expiry")
	}
	expFloat, ok := expAny.(float64)
	if !ok {
		return dto.AuthResponse{}, errors.New("invalid expiry type")
	}
	if float64(time.Now().Unix()) > expFloat {
		return dto.AuthResponse{}, errors.New("token expired")
	}

	return dto.AuthResponse{
		UserID: int(claims["user_id"].(float64)),
		Email:  claims["email"].(string),
		Expiry: expFloat,
		Iat:    claims["iat"].(float64),
	}, nil
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
