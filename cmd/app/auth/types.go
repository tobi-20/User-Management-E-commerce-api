package auth

import "github.com/jackc/pgx/v5/pgtype"

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type CreateUserResp struct {
	Name  string
	Email string
}

type SignupReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogoutReq struct {
	ID int64 `json:"id"`
}

type RefreshArgs struct {
	UserID      int64              `json:"user_id"`
	HashedToken string             `json:"hashed_token"`
	ExpiresAt   pgtype.Timestamptz `json:"expires_at"`
	TokenID     string             `json:"token_id"`
}
