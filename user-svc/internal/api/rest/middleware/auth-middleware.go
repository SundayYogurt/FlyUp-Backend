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
		// grab authorization header
		authHeader := ctx.Get("Authorization")

		//	verify token
		user, err := VerifyToken(authHeader)
		if err != nil {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		//	if token is valid, process to the next handler

		//	assign the decoded user to the context
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
	tokenArr := strings.Split(tokenString, " ")

	if len(tokenArr) != 2 || tokenArr[0] != "Bearer" {
		return dto.AuthResponse{}, errors.New("invalid token format")
	}
	tokenStr := tokenArr[1]

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
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

	if float64(time.Now().Unix()) > claims["expiry"].(float64) {
		return dto.AuthResponse{}, errors.New("token expired")
	}

	return dto.AuthResponse{
		UserID: int(claims["user_id"].(float64)),
		Email:  claims["email"].(string),
		Expiry: claims["expiry"].(float64),
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
