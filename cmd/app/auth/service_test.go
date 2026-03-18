package auth

import (
	"context"
	repo "ecom/internal/adapters/postgresql/sqlc"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) ConsumeVerification(ctx context.Context, id pgtype.UUID) (repo.ConsumeVerificationRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repo.ConsumeVerificationRow), args.Error(1)
}
func (m *mockRepo) GetResetPasswordBySelector(ctx context.Context, selector string) (repo.GetResetPasswordBySelectorRow, error) {
	args := m.Called(ctx, selector)
	return args.Get(0).(repo.GetResetPasswordBySelectorRow), args.Error(1)
}
func (m *mockRepo) ConsumeRefreshTokenByID(ctx context.Context, tokenID string) (repo.ConsumeRefreshTokenByIDRow, error) {
	args := m.Called(ctx, tokenID)
	return args.Get(0).(repo.ConsumeRefreshTokenByIDRow), args.Error(1)
}
func (m *mockRepo) CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repo.User), args.Error(1)
}
func (m *mockRepo) GetUserByEmail(ctx context.Context, email string) (repo.GetUserByEmailRow, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(repo.GetUserByEmailRow), args.Error(1)
}
func (m *mockRepo) GetRefreshTokenByID(ctx context.Context, tokenID string) (repo.RefreshToken, error) {
	args := m.Called(ctx, tokenID)
	return args.Get(0).(repo.RefreshToken), args.Error(1)
}
func (m *mockRepo) GetUserByID(ctx context.Context, id int64) (repo.GetUserByIDRow, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(repo.GetUserByIDRow), args.Error(1)

}

func (m *mockRepo) GetVerificationByToken(ctx context.Context, token string) (repo.GetVerificationByTokenRow, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(repo.GetVerificationByTokenRow), args.Error(1)
}
func (m *mockRepo) SaveRefreshToken(ctx context.Context, arg repo.SaveRefreshTokenParams) (repo.RefreshToken, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repo.RefreshToken), args.Error(1)
}
func (m *mockRepo) SaveOneTimeToken(ctx context.Context, arg repo.SaveOneTimeTokenParams) (repo.SaveOneTimeTokenRow, error) {
	args := m.Called(arg)
	return args.Get(0).(repo.SaveOneTimeTokenRow), args.Error(1)
}
func (m *mockRepo) SaveResetPassword(ctx context.Context, arg repo.SaveResetPasswordParams) (repo.SaveResetPasswordRow, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(repo.SaveResetPasswordRow), args.Error(1)
}
func (m *mockRepo) UpdatePassword(ctx context.Context, arg repo.UpdatePasswordParams) (string, error) {
	args := m.Called(ctx, arg)
	return args.Get(0).(string), args.Error(1)
}
func (m *mockRepo) UpdateVerificationUsers(ctx context.Context, arg repo.UpdateVerificationUsersParams) error {
	args := m.Called(ctx, arg)
	return args.Error(1)
}
func (m *mockRepo) UpdateVerifiedState(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(1)
}
func (m *mockRepo) DeleteAllRefreshTokenByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(1)
}
func (m *mockRepo) UpdateResetPasswordStatus(ctx context.Context, selector string) error {
	args := m.Called(ctx, selector)
	return args.Error(1)
}
func (m *mockRepo) ConsumePasswordReset(ctx context.Context, selector string) (repo.ConsumePasswordResetRow, error) {
	args := m.Called(ctx, selector)

	return args.Get(0).(repo.ConsumePasswordResetRow), args.Error(1)
}

func (m *mockRepo) WithTx(tx pgx.Tx) *repo.Queries {
	args := m.Called(tx)
	return args.Get(0).(*repo.Queries)
}

func TestValidateResetPasswordTokens_SelectorNotFound(t *testing.T) {
	m := &mockRepo{}
	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{}, errors.New("not found"))

	svc := NewService(m)
	_, _, err := svc.ValidateResetPasswordTokens(context.Background(), "selector", "verifier")
	assert.ErrorIs(t, err, ErrTokenNotFound)
}
func TestValidateResetPasswordTokens_VerifierWrong(t *testing.T) {
	m := &mockRepo{}
	text := "dhchvhvbfhbvbfjsvbbvfhsbbvnsvs"

	hashed, _ := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
	}, nil)

	svc := NewService(m)
	_, _, err := svc.ValidateResetPasswordTokens(context.Background(), "selector", "verifier")

	assert.Error(t, err)
}
func TestValidateResetPasswordTokens_ResetTokenUsed(t *testing.T) {
	m := &mockRepo{}
	text := "verifier"

	hashed, _ := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
		IsUsed:       true,
	}, nil)

	svc := NewService(m)
	_, _, err := svc.ValidateResetPasswordTokens(context.Background(), "selector", "verifier")

	assert.ErrorIs(t, err, ErrTokenAlreadyUsed)
}
func TestValidateResetPasswordTokens_TokenExpired(t *testing.T) {
	m := &mockRepo{}
	text := "verifier"

	hashed, _ := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
		IsUsed:       false,
		Expiry: pgtype.Timestamptz{
			Time:  time.Now().Add(-time.Hour),
			Valid: true,
		},
	}, nil)

	svc := NewService(m)
	_, _, err := svc.ValidateResetPasswordTokens(context.Background(), "selector", "verifier")

	assert.ErrorIs(t, err, ErrLinkExpired)
}

func TestValidateResetPasswordTokens_Happy(t *testing.T) {
	m := &mockRepo{}
	svc := NewService(m)
	text := "verifier"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(text), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
		UserID:       int64(3),
		IsUsed:       false,
		Expiry: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	}, nil)
	userId, _, err := svc.ValidateResetPasswordTokens(context.Background(), "selector", text)
	assert.NoError(t, err)
	assert.Equal(t, userId, int64(3))
}

func TestResetPassword_ValidateResetPasswordTokens(t *testing.T) {
	m := &mockRepo{}
	var db *pgx.Conn
	testParams := ResetPassWordReq{
		Selector: "selector",
		Verifier: "verifier",
		Password: "password",
	}

	svc := NewService(m)
	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{}, errors.New("error"))

	err := svc.ResetPassword(context.Background(), db, testParams)
	assert.Error(t, err)
}
func TestResetPassword_UpdatePassword(t *testing.T) {
	m := &mockRepo{}
	svc := NewService(m)
	var db *pgx.Conn
	testParams := ResetPassWordReq{
		Selector: "selector",
		Verifier: "verifier",
		Password: "password",
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(testParams.Verifier), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
		UserID:       int64(3),
		IsUsed:       false,
		Expiry: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	}, nil)

	m.On("UpdatePassword", mock.Anything, mock.Anything).Return("", errors.New("error"))

	err := svc.ResetPassword(context.Background(), db, testParams)
	assert.Error(t, err)
}
func TestResetPassword_Happy(t *testing.T) {
	m := &mockRepo{}
	svc := NewService(m)
	var db *pgx.Conn
	testParams := ResetPassWordReq{
		Selector: "selector",
		Verifier: "verifier",
		Password: "password",
	}

	hashed, _ := bcrypt.GenerateFromPassword([]byte(testParams.Verifier), bcrypt.DefaultCost)

	m.On("GetResetPasswordBySelector", mock.Anything, "selector").Return(repo.GetResetPasswordBySelectorRow{
		VerifierHash: string(hashed),
		UserID:       int64(3),
		IsUsed:       false,
		Expiry: pgtype.Timestamptz{
			Time:  time.Now().Add(time.Hour),
			Valid: true,
		},
	}, nil)

	m.On("UpdatePassword", mock.Anything, mock.Anything).Return("success", nil)

	err := svc.ResetPassword(context.Background(), db, testParams)
	assert.NoError(t, err)
}
