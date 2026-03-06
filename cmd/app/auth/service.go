package auth

import (
	"Lanixpress/cmd/helpers"
	repo "Lanixpress/internal/adapters/postgresql/sqlc"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"time"

	"github.com/golang-jwt/jwt"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error)
	GetUserByEmail(ctx context.Context, email string) (repo.GetUserByEmailRow, error)
	GenerateAccessToken(ctx context.Context, user helpers.User) (string, error)
	SaveRefreshToken(ctx context.Context, arg repo.SaveRefreshTokenParams) (repo.RefreshToken, error)
	DeleteRefreshTokenByID(ctx context.Context, tokenID string) error
	Login(ctx context.Context, req *LoginReq) (string, string, error)
	GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error)
	GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error)
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

//methods implementation

func (s *svc) CreateUser(ctx context.Context, args repo.CreateUserParams) (repo.User, error) {
	return s.repo.CreateUser(ctx, args)

}

func (s *svc) Login(ctx context.Context, req *LoginReq) (string, string, error) {

	//Get the user details by means of the email
	user, err := s.GetUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return "", "", err
	}

	//compare to validate that password is correct
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return "", "", errors.New("invalid credentials")
	}
	log.Println("successful login!")

	//Generate the access token with a helper function
	accessToken, err := s.GenerateAccessToken(ctx, helpers.ToUserEmail(user))
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	//Generate raw token and the hashed token with another helper function
	rawToken, hashedToken, err := helpers.GenerateRefreshToken()
	tokenID, _, _ := helpers.SplitToken(rawToken)

	params := repo.SaveRefreshTokenParams{
		UserID:      user.ID,
		HashedToken: hashedToken,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(7 * 24 * time.Hour),
			Valid: true,
		},

		TokenID: tokenID,
	}
	_, err = s.SaveRefreshToken(ctx, params)
	log.Println(err)

	return accessToken, rawToken, nil
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

func (s *svc) DeleteRefreshTokenByID(ctx context.Context, tokenID string) error {
	return s.repo.DeleteRefreshTokenByID(ctx, tokenID)
}

func (s *svc) GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error) {
	return s.repo.GetRefreshTokenByID(ctx, tokenID)
}

func (s *svc) GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error) {
	return s.repo.GetUserByID(ctx, id)
}
