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

type Service interface {
	ConsumeRefreshTokenByID(ctx context.Context, tokenID string) (repo.ConsumeRefreshTokenByIDRow, error)
	ConsumeVerification(ctx context.Context, id pgtype.UUID) (repo.ConsumeVerificationRow, error)
	ConsumeCookie(ctx context.Context, cookie *http.Cookie) error
	CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error)
	GenerateAccessToken(ctx context.Context, user helpers.User) (string, error)
	GetUserByEmail(ctx context.Context, email string) (repo.GetUserByEmailRow, error)
	GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error)
	GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error)
	GetResetPasswordBySelector(ctx context.Context, selector string) (repo.GetResetPasswordBySelectorRow, error)
	GetVerificationByToken(ctx context.Context, token string) (repo.GetVerificationByTokenRow, error)
	IssueAuthTokens(ctx context.Context, user helpers.User) (string, string, error)
	IssueRefreshToken(ctx context.Context, rawToken string) (string, string, error)
	SaveRefreshToken(ctx context.Context, arg repo.SaveRefreshTokenParams) (repo.RefreshToken, error)
	Login(ctx context.Context, req LoginReq) (string, string, error)
	ResetPassword(ctx context.Context, newPasswordParam ResetPassWordReq) error
	SaveOneTimeToken(ctx context.Context, arg repo.SaveOneTimeTokenParams) (repo.SaveOneTimeTokenRow, error)
	SaveResetPassword(ctx context.Context, arg repo.SaveResetPasswordParams) (repo.SaveResetPasswordRow, error)
	SendResetTokenToEmail(ctx context.Context, req ForgotPasswordRequest) error
	SendConfirmUserTokenToEmail(ctx context.Context, req SignupReq) error
	UpdatePassword(ctx context.Context, arg repo.UpdatePasswordParams) (string, error)
	UpdateVerificationUsers(ctx context.Context, arg repo.UpdateVerificationUsersParams) error
	UpdateVerifiedState(ctx context.Context, id int64) error
	ValidateVerification(ctx context.Context, verifier string, selector string) (repo.GetUserByIDRow, error)
	ValidateResetPasswordTokens(ctx context.Context, selector string, verifier string) (int64, error)
}

type svc struct {
	repo      *repo.Queries
	jwtSecret string
}

func NewService(repo *repo.Queries) Service {
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

	createdUser, err := s.CreateUser(ctx, params)
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

	if _, err := s.SaveOneTimeToken(ctx, verifyParams); err != nil {
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
func (s *svc) CreateUser(ctx context.Context, args repo.CreateUserParams) (repo.User, error) {
	return s.repo.CreateUser(ctx, args)

}

func (s *svc) Login(ctx context.Context, req LoginReq) (string, string, error) {

	//Get the user details by means of the email
	user, err := s.GetUserByEmail(ctx, strings.ToLower(req.Email))
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

func (s *svc) GetUserByEmail(ctx context.Context, email string) (repo.GetUserByEmailRow, error) {
	return s.repo.GetUserByEmail(ctx, email)
}

// saving generated refresh token to database
func (s *svc) SaveRefreshToken(ctx context.Context, arg repo.SaveRefreshTokenParams) (repo.RefreshToken, error) {
	return s.repo.SaveRefreshToken(ctx, arg)
}

func (s *svc) ConsumeRefreshTokenByID(ctx context.Context, tokenID string) (repo.ConsumeRefreshTokenByIDRow, error) {
	return s.repo.ConsumeRefreshTokenByID(ctx, tokenID)
}

func (s *svc) GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error) {
	return s.repo.GetRefreshTokenByID(ctx, tokenID)
}

func (s *svc) GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *svc) ConsumeVerification(ctx context.Context, id pgtype.UUID) (repo.ConsumeVerificationRow, error) {
	return s.repo.ConsumeVerification(ctx, id)
}

func (s *svc) SaveOneTimeToken(ctx context.Context, arg repo.SaveOneTimeTokenParams) (repo.SaveOneTimeTokenRow, error) {
	return s.repo.SaveOneTimeToken(ctx, arg)
}

func (s *svc) GetVerificationByToken(ctx context.Context, token string) (repo.GetVerificationByTokenRow, error) {
	return s.repo.GetVerificationByToken(ctx, token)
}
func (s *svc) UpdateVerificationUsers(ctx context.Context, arg repo.UpdateVerificationUsersParams) error {
	return s.repo.UpdateVerificationUsers(ctx, arg)
}
func (s *svc) UpdateVerifiedState(ctx context.Context, id int64) error {
	return s.repo.UpdateVerifiedState(ctx, id)
}

func (s *svc) SaveResetPassword(ctx context.Context, arg repo.SaveResetPasswordParams) (repo.SaveResetPasswordRow, error) {
	return s.repo.SaveResetPassword(ctx, arg)
}

func (s *svc) GetResetPasswordBySelector(ctx context.Context, selector string) (repo.GetResetPasswordBySelectorRow, error) {
	return s.repo.GetResetPasswordBySelector(ctx, selector)
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
	if _, err = s.SaveRefreshToken(ctx, params); err != nil {
		log.Println(err)
		return "", "", err
	}

	return accessToken, rawToken, nil
}

func (s *svc) UpdatePassword(ctx context.Context, arg repo.UpdatePasswordParams) (string, error) {
	return s.repo.UpdatePassword(ctx, arg)
}

func (s *svc) ValidateVerification(ctx context.Context, verifier string, selector string) (repo.GetUserByIDRow, error) {

	verifyRow, err := s.GetVerificationByToken(ctx, selector)
	if err != nil {
		return repo.GetUserByIDRow{}, ErrInvalidToken
	}
	user, err := s.GetUserByID(ctx, verifyRow.UserID)
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

	if _, err = s.ConsumeVerification(ctx, verifyRow.ID); err != nil {

		return repo.GetUserByIDRow{}, ErrVerificationFailed
	}
	if err = s.UpdateVerifiedState(ctx, verifyRow.UserID); err != nil {
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

	refreshToken, err := s.GetRefreshTokenByID(ctx, tokenID)
	if err != nil {
		return "", "", ErrInvalidToken
	}
	if time.Now().After(refreshToken.ExpiresAt.Time) {
		return "", "", ErrRefreshTokenExpired
	}

	if time.Since(refreshToken.CreatedAt.Time) > globals.MaxLifetime {
		return "", "", ErrRefreshTokenMaxLifetime
	}

	stored, err := s.ConsumeRefreshTokenByID(ctx, tokenID)
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
	if _, err = s.SaveRefreshToken(ctx, newParams); err != nil {
		return "", "", err
	}

	user, err := s.GetUserByID(ctx, stored.UserID)
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
	if _, err := s.ConsumeRefreshTokenByID(ctx, tokenID); err != nil {
		return err
	}
	return nil
}

func (s *svc) SendResetTokenToEmail(ctx context.Context, req ForgotPasswordRequest) error {
	user, err := s.GetUserByEmail(ctx, strings.ToLower(req.Email))
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

	if _, err = s.SaveResetPassword(ctx, params); err != nil {
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
	resetPasswordRow, err := s.GetResetPasswordBySelector(ctx, selector)
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

	if _, err = s.UpdatePassword(ctx, params); err != nil {
		return err
	}
	return nil
}
