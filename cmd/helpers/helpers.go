package helpers

import (
	repo "Lanixpress/internal/adapters/postgresql/sqlc"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
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
	secret := GenerateRandString(32)

	hashedToken, err := bcrypt.GenerateFromPassword([]byte(secret), 11)
	if err != nil {
		return "", "", err
	}
	return fmt.Sprintf("%s.%s", tokenID, secret), string(hashedToken), nil
}

func GenerateRandString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func SplitToken(token string) (tokenID, secret string, err error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", err
	}
	return parts[0], parts[1], nil
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
func ToCreatedUser(u repo.User) User {
	return User{
		ID:           u.ID,
		Name:         u.Name,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		TokenVersion: u.TokenVersion,
	}
}

func SendVerificationLinkToEmail(link, email string) error {

	from := "olutobiseun18@gmail.com"
	password := "bfwhtpyrmycnsoqx"

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	to := []string{email}
	msg := []byte(fmt.Sprintf("MIME-Version: 1.0\r\n"+"Content-Type: text/html; charset=UTF-8\r\n"+"To: %s\r\n"+"Subject: Verify your email\r\n"+"\r\n"+`<p>Click the link to verify your account:</p><a href="%s">Verify Email</a>`, email, link))

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
	if err != nil {
		log.Println("failed to send email:", err)
		return err
	}
	return nil
}
func SendResetPasswordLinkToEmail(link, email string) error {

	from := "olutobiseun18@gmail.com"
	password := "bfwhtpyrmycnsoqx"

	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	to := []string{email}
	msg := []byte(fmt.Sprintf("MIME-Version: 1.0\r\n"+"Content-Type: text/html; charset=UTF-8\r\n"+"To: %s\r\n"+"Subject: Reset your Password\r\n"+"\r\n"+`<p>Click the link to rest your password:</p><a href="%s">Reset password</a>`, email, link))

	err := smtp.SendMail("smtp.gmail.com:587", auth, from, to, msg)
	if err != nil {
		log.Println("failed to send email:", err)
		return err
	}
	return nil
}
