package helper

import (
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

func HasConsent(codes []string, need string) bool {
	need = strings.ToUpper(need)
	for _, c := range codes {
		if strings.ToUpper(strings.TrimSpace(c)) == need {
			return true
		}
	}
	return false
}

func IsDuplicateConsent(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.ConstraintName == "uidx_user_consents_code_ver"
	}
	return false
}
