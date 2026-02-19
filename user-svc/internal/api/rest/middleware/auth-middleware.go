package middleware

import (
	"errors"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		// grab authorization header

		//	verify token

		//	if token is valid, process to the next handler

		//	assign the decoded user to the context
		return ctx.Next()
	}
}

func GenerateToken(userID int, email string) (string, error) {
	if userID == 0 || email == "" {
		return "", errors.New("userID and email are required")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     jwt.NewNumericDate(time.Now()),
		"exp":     jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("ACCESS_SECRET")))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}
