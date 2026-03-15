package auth

import (
	"Lanixpress/cmd/app/globals"
	"Lanixpress/cmd/helpers"
	repo "Lanixpress/internal/adapters/postgresql/sqlc"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidToken            = errors.New("invalid token")
	ErrUserNotFound            = errors.New("user not found")
	ErrTokenAlreadyUsed        = errors.New("token already used")
	ErrLinkExpired             = errors.New("one time link expired")
	ErrVerificationFailed      = errors.New("email verification failed")
	ErrUpdateVerifiedFailed    = errors.New("updating verified state failed")
	ErrRefreshTokenExpired     = errors.New("refresh token expired")
	ErrRefreshTokenMaxLifetime = errors.New("refresh token reached max lifetime")
	ErrTokenNotFound           = errors.New("refresh token not found")
	ErrTokenRevokeFailed       = errors.New("failed to revoke refresh token")
)

type AuthService interface {
	ConsumeCookie(ctx context.Context, cookie *http.Cookie) error
	GenerateAccessToken(ctx context.Context, user helpers.User) (string, error)
	IssueAuthTokens(ctx context.Context, user helpers.User) (string, string, error)
	IssueRefreshToken(ctx context.Context, rawToken string) (string, string, error)
	Login(ctx context.Context, req LoginReq) (string, string, error)
	ResetPassword(ctx context.Context, newPasswordParam ResetPassWordReq) error
	ValidateResetPasswordTokens(ctx context.Context, selector string, verifier string) (int64, error)
	SendConfirmUserTokenToEmail(ctx context.Context, req SignupReq) error
	SendResetTokenToEmail(ctx context.Context, req ForgotPasswordRequest) error
	ValidateVerification(ctx context.Context, verifier string, selector string) (repo.GetUserByIDRow, error)
}

type svc struct {
	repo      AuthRepository
	jwtSecret string
}

func NewService(repo AuthRepository) AuthService {
	return &svc{
		repo:      repo,
		jwtSecret: os.Getenv("JWT_ACCESS_SECRET"),
	}
}

// methods implementation
func (s *svc) SendConfirmUserTokenToEmail(ctx context.Context, req SignupReq) error {
	email := strings.ToLower(req.Email)
	selector := helpers.GenerateRandString(32)
	verifier := helpers.GenerateRandString(32)
	hashedPassword, err := helpers.HashPassword(req.Password)
	if err != nil {
		return err
	}

	params := repo.CreateUserParams{
		Name:         req.Name,
		Email:        email,
		PasswordHash: hashedPassword,
	}

	createdUser, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
				return ErrEmailAlreadyExists
			}
		}
		return err
	}

	hashedVerifier, err := helpers.HashPassword(verifier)
	if err != nil {
		return ErrInvalidToken
	}

	verifyParams := repo.SaveOneTimeTokenParams{
		UserID:   createdUser.ID,
		Selector: selector,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(globals.MaxVerifyTime),
			Valid: true,
		},
		VerifierHash: hashedVerifier,
	}

	if _, err := s.repo.SaveOneTimeToken(ctx, verifyParams); err != nil {
		return err
	}

	path := "/verify-user"
	link := fmt.Sprintf(
		"%s%s?selector=%s&verifier=%s",
		globals.SitePath, path, selector, verifier)

	go func() {
		verifyErr := helpers.SendVerificationLinkToEmail(link, createdUser.Email)
		if verifyErr != nil {
			log.Println("failed to send verification email:", err)
		}
	}()
	return nil
}

func (s *svc) Login(ctx context.Context, req LoginReq) (string, string, error) {

	//Get the user details by means of the email
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return "", "", ErrUserNotFound
	}
	accessToken, refreshToken, err := s.IssueAuthTokens(ctx, helpers.ToUserEmail(user))

	//compare to validate that password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", "", ErrUserNotFound
	}
	log.Println("successful login!")

	//Generate the access token with a helper function
	return accessToken, refreshToken, nil
}

func (s *svc) GenerateAccessToken(ctx context.Context, user helpers.User) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"exp":   time.Now().Add(15 * time.Minute).Unix(), // expires in 15 minutes
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func (s *svc) IssueAuthTokens(ctx context.Context, user helpers.User) (string, string, error) {

	accessToken, err := s.GenerateAccessToken(ctx, user)
	if err != nil {
		return "", "", err
	}

	//Generate raw token and the hashed token with another helper function
	rawToken, hashedToken, err := helpers.GenerateRefreshToken()

	if err != nil {
		return "", "", err
	}
	tokenID, _, err := helpers.SplitToken(rawToken)
	if err != nil {
		return "", "", err
	}

	params := repo.SaveRefreshTokenParams{
		UserID:      user.ID,
		HashedToken: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(globals.RefreshTokenExpiry),
			Valid: true,
		},

		TokenID: tokenID,
	}
	if _, err = s.repo.SaveRefreshToken(ctx, params); err != nil {
		log.Println(err)
		return "", "", err
	}

	return accessToken, rawToken, nil
}

func (s *svc) ValidateVerification(ctx context.Context, verifier string, selector string) (repo.GetUserByIDRow, error) {

	verifyRow, err := s.repo.GetVerificationByToken(ctx, selector)
	if err != nil {
		return repo.GetUserByIDRow{}, ErrInvalidToken
	}
	user, err := s.repo.GetUserByID(ctx, verifyRow.UserID)
	if err != nil {
		return repo.GetUserByIDRow{}, ErrUserNotFound
	}
	if user.IsVerified.Bool {
		return repo.GetUserByIDRow{}, ErrTokenAlreadyUsed
	}

	err = bcrypt.CompareHashAndPassword([]byte(verifyRow.VerifierHash), []byte(verifier))
	if err != nil {

		return repo.GetUserByIDRow{}, err
	}
	if time.Now().After(verifyRow.ExpiresAt.Time) {

		return repo.GetUserByIDRow{}, ErrLinkExpired
	}

	if _, err = s.repo.ConsumeVerification(ctx, verifyRow.ID); err != nil {

		return repo.GetUserByIDRow{}, ErrVerificationFailed
	}
	if err = s.repo.UpdateVerifiedState(ctx, verifyRow.UserID); err != nil {
		log.Println(err.Error())
		return repo.GetUserByIDRow{}, ErrUpdateVerifiedFailed
	}

	return user, nil
}

func (s *svc) IssueRefreshToken(ctx context.Context, rawToken string) (string, string, error) {
	tokenID, secret, err := helpers.SplitToken(rawToken)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.repo.GetRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return "", "", ErrInvalidToken
	}
	if time.Now().After(refreshToken.ExpiresAt.Time) {
		return "", "", ErrRefreshTokenExpired
	}

	if time.Since(refreshToken.CreatedAt.Time) > globals.MaxLifetime {
		return "", "", ErrRefreshTokenMaxLifetime
	}

	stored, err := s.repo.ConsumeRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return "", "", errors.New("token missing")
	}

	if err = bcrypt.CompareHashAndPassword([]byte(stored.HashedToken), []byte(secret)); err != nil {
		return "", "", errors.New("token mismatch")
	}

	newRaw, newHash, err := helpers.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	newTokenId, _, err := helpers.SplitToken(newRaw)
	if err != nil {
		return "", "", err
	}

	newParams := repo.SaveRefreshTokenParams{
		UserID:      stored.UserID,
		HashedToken: newHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(globals.RefreshTokenExpiry),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  stored.CreatedAt.Time,
			Valid: true,
		},
		TokenID: newTokenId,
	}
	if _, err = s.repo.SaveRefreshToken(ctx, newParams); err != nil {
		return "", "", err
	}

	user, err := s.repo.GetUserByID(ctx, stored.UserID)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	newAccessToken, err := s.GenerateAccessToken(ctx, helpers.ToUserID(user))
	if err != nil {
		return "", "", err
	}
	return newAccessToken, rawToken, nil
}

func (s *svc) ConsumeCookie(ctx context.Context, cookie *http.Cookie) error {
	tokenID, _, err := helpers.SplitToken(cookie.Value)

	if err != nil {
		return err
	}
	if _, err := s.repo.ConsumeRefreshTokenByID(ctx, tokenID); err != nil {
		return err
	}
	return nil
}

func (s *svc) SendResetTokenToEmail(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(req.Email))
	log.Println(req.Email)
	if err != nil {
		return ErrUserNotFound
	}
	selector := helpers.GenerateRandString(32)
	verifier := helpers.GenerateRandString(32)
	hashed, err := helpers.HashPassword(verifier)
	if err != nil {
		return err
	}
	params := repo.SaveResetPasswordParams{
		UserID:       user.ID,
		VerifierHash: hashed,
		Selector:     selector,

		Expiry: pgtype.Timestamptz{
			Time:  time.Now().Add(globals.MaxVerifyTime),
			Valid: true,
		},
	}

	if _, err = s.repo.SaveResetPassword(ctx, params); err != nil {
		return err
	}

	route := "/reset-password"
	link := fmt.Sprintf("%s%s?selector=%s&verifier=%s", globals.SitePath, route, selector, verifier)

	go func() {
		err := helpers.SendResetPasswordLinkToEmail(link, req.Email) // this blocks and sometimes email might be dramatic
		if err != nil {
			log.Println("email error:", err)
		}
	}()
	return nil
}

func (s *svc) ValidateResetPasswordTokens(ctx context.Context, selector string, verifier string) (int64, error) {
	resetPasswordRow, err := s.repo.GetResetPasswordBySelector(ctx, selector)
	if err != nil {
		return 0, ErrTokenNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(resetPasswordRow.VerifierHash), []byte(verifier))
	if err != nil {
		return 0, err
	}

	if resetPasswordRow.IsUsed {
		return 0, ErrTokenAlreadyUsed
	}
	if time.Now().After(resetPasswordRow.Expiry.Time) {
		return 0, ErrLinkExpired
	}
	return resetPasswordRow.UserID, nil
}
func (s *svc) ResetPassword(ctx context.Context, newPassParams ResetPassWordReq) error {
	userId, err := s.ValidateResetPasswordTokens(ctx, newPassParams.Selector, newPassParams.Verifier)
	hashedNew, err := helpers.HashPassword(newPassParams.Password)
	if err != nil {
		return err
	}
	params := repo.UpdatePasswordParams{
		PasswordHash: hashedNew,
		ID:           userId,
	}

	if _, err = s.repo.UpdatePassword(ctx, params); err != nil {
		return err
	}
	return nil
}
