package utils

import (
	"errors"
	"strings"
)

func ExtractEmailDomain(email string) (string, error) {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", errors.New("invalid email format")
	}
	return parts[1], nil
}
