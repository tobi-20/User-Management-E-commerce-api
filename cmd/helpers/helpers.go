package helpers

import (
	repo "Lanixpress/internal/adapters/postgresql/sqlc"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func TextToString(t pgtype.Text) string {
	if t.Valid {
		return t.String
	}
	return ""
}

func GenerateRefreshToken() (string, string, error) {
	tokenID := uuid.New().String()
	secret := randomString(32)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(secret), 11)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s.%s", tokenID, secret), string(hashedToken), nil
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func SplitToken(token string) (tokenID, secret string, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
func ToUserEmail(row repo.GetUserByEmailRow) User {
	return User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		TokenVersion: row.TokenVersion,
	}
}
func ToUserID(row repo.GetUserByIDRow) User {
	return User{
		ID:           row.ID,
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		TokenVersion: row.TokenVersion,
	}
}
