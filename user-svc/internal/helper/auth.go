package helper

import (
	"errors"
	"strings"
	"time"

	"github.com/SundayYogurt/user_service/internal/dto"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Secret string
}

func SetupAuth(s string) Auth {
	return Auth{
		Secret: s,
	}
}

func (a Auth) GenerateToken(userID int, email string) (string, error) {
	if userID == 0 || email == "" {
		return "", errors.New("required inputs are missing to generate token")
	}

	now := time.Now().Unix()
	exp := time.Now().Add(24 * time.Hour).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     now,
		"exp":     exp,
	})

	tokenStr, err := token.SignedString([]byte(a.Secret))
	if err != nil {
		return "", errors.New("unable to sign the token")
	}

	return tokenStr, nil
}

func (a Auth) VerifyToken(tokenString string) (dto.AuthResponse, error) {
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
		return []byte(a.Secret), nil
	})
	if err != nil {
		return dto.AuthResponse{}, errors.New("token parse error")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return dto.AuthResponse{}, errors.New("invalid token claims")
	}

	// safer expiry parse
	expAny, ok := claims["exp"]
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

func (a Auth) GetCurrentUser(ctx *fiber.Ctx) (dto.AuthResponse, error) {
	u := ctx.Locals("user")
	claims, ok := u.(dto.AuthResponse)
	if !ok {
		return dto.AuthResponse{}, errors.New("missing auth user in context")
	}
	return claims, nil
}

func (a Auth) VerifyPassword(plain, hashed string) error {
	if err := bcrypt.CompareHashAndPassword(
		[]byte(hashed),
		[]byte(plain),
	); err != nil {
		return errors.New("invalid email or password")
	}
	return nil
}
