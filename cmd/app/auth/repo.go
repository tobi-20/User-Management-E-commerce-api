package auth

import (
	repo "ecom/internal/adapters/postgresql/sqlc"
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type AuthRepository interface {
	ConsumeRefreshTokenByID(ctx context.Context, tokenID string) (repo.ConsumeRefreshTokenByIDRow, error)
	ConsumeVerification(ctx context.Context, id pgtype.UUID) (repo.ConsumeVerificationRow, error)
	CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error)
	GetUserByEmail(ctx context.Context, email string) (repo.GetUserByEmailRow, error)
	GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error)
	GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error)
	GetResetPasswordBySelector(ctx context.Context, selector string) (repo.GetResetPasswordBySelectorRow, error)
	GetVerificationByToken(ctx context.Context, token string) (repo.GetVerificationByTokenRow, error)
	SaveRefreshToken(ctx context.Context, arg repo.SaveRefreshTokenParams) (repo.RefreshToken, error)
	SaveOneTimeToken(ctx context.Context, arg repo.SaveOneTimeTokenParams) (repo.SaveOneTimeTokenRow, error)
	SaveResetPassword(ctx context.Context, arg repo.SaveResetPasswordParams) (repo.SaveResetPasswordRow, error)
	UpdatePassword(ctx context.Context, arg repo.UpdatePasswordParams) (string, error)
	UpdateVerificationUsers(ctx context.Context, arg repo.UpdateVerificationUsersParams) error
	UpdateVerifiedState(ctx context.Context, id int64) error
	// UpdateResetPasswordStatus(ctx context.Context, id int64) error
}
