package auth

import (
	"ecom/cmd/helpers"
	repo "ecom/internal/adapters/postgresql/sqlc"
	"ecom/internal/json"
	"errors"
	"strings"
	"time"

	"log"

	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type handler struct {
	service Service
}

func NewHandler(service Service) *handler {
	return &handler{
		service: service,
	}
}

func (h *handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupReq

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.Read(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := helpers.HashPassword(req.Password)

	if err != nil {
		http.Error(w, "password not hashed successfully", http.StatusInternalServerError)
		return
	}

	params := repo.CreateUserParams{
		Name:         req.Name,
		Email:        strings.ToLower(req.Email),
		PasswordHash: hashedPassword,
	}

	createdUser, err := h.service.CreateUser(r.Context(), params)
	log.Println(createdUser)
	if err != nil {
		log.Println(err)
	}
	accessToken, err := h.service.GenerateAccessToken(r.Context(), helpers.ToCreatedUser(createdUser))
	if err != nil {
		http.Error(w, "access token not generated", http.StatusExpectationFailed)
	}
	raw, hashed, err := helpers.GenerateRefreshToken()
	if err != nil {
		log.Println(err.Error())
		return
	}

	tokenId, _, ok := helpers.SplitToken(raw)
	if !ok {
		log.Println("unable to generate token id")
	}

	saved := repo.SaveRefreshTokenParams{
		UserID:      createdUser.ID,
		HashedToken: hashed,
		ExpiresAt: pgtype.Timestamptz{
			Time:  helpers.RefreshTokenExpiry,
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  time.Now(),
			Valid: true,
		},
		TokenID: tokenId,
	}
	h.service.SaveRefreshToken(r.Context(), saved)

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    raw,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	json.Write(w, http.StatusCreated, map[string]interface{}{
		"access_token": accessToken,
		"name":         createdUser.Name,
		"email":        createdUser.Email,
	})

}

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginReq
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.Read(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessToken, rawRefreshToken, err := h.service.Login(r.Context(), &req)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
		// os.Exit(1) never do this it kills the server

	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    rawRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	json.Write(w, http.StatusOK, map[string]string{
		"access_token": accessToken,
	})
}
func (h *handler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusBadRequest)
		return
	}
}

func (h *handler) RefreshToken(w http.ResponseWriter, r *http.Request) (string, error) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}

	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		return "", errors.New("missing refresh token")
	}
	rawToken := cookie.Value
	tokenID, secret, ok := helpers.SplitToken(rawToken)
	if !ok {
		return "", errors.New("invalid token format")
	}

	stored, err := h.service.GetRefreshTokenByID(r.Context(), tokenID)
	if err != nil {
		return "", errors.New("Unable to fetch token")
	}
	if time.Now().After(stored.ExpiresAt.Time) {
		return "", errors.New("token expired")
	}

	err = bcrypt.CompareHashAndPassword([]byte(stored.HashedToken), []byte(secret))
	if err != nil {
		return "", errors.New("Invalid token")
	}
	err = h.service.DeleteRefreshTokenByID(r.Context(), tokenID)

	newRaw, newHash, _ := helpers.GenerateRefreshToken()
	newTokenId, _, _ := helpers.SplitToken(newRaw)

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRaw,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	maxLifetime := 30 * 24 * time.Hour

	if time.Since(stored.CreatedAt.Time) > maxLifetime {
		return "", errors.New("refresh token reached max lifetime, login required")
	}

	newParams := repo.SaveRefreshTokenParams{
		UserID:      stored.UserID,
		HashedToken: newHash,
		ExpiresAt: pgtype.Timestamptz{
			Time:  time.Now().Add(7 * 24 * time.Hour),
			Valid: true,
		},
		CreatedAt: pgtype.Timestamptz{
			Time:  stored.CreatedAt.Time,
			Valid: true,
		},
		TokenID: newTokenId,
	}
	_, err = h.service.SaveRefreshToken(r.Context(), newParams)

	user, err := h.service.GetUserByID(r.Context(), stored.UserID)
	if err != nil {
		return "", errors.New("User does not exist")
	}

	newAccessToken, err := h.service.GenerateAccessToken(r.Context(), helpers.ToUserID(user))
	if err != nil {
		return "", err
	}

	json.Write(w, http.StatusOK, map[string]string{
		"access_token": newAccessToken})

	return newAccessToken, nil
}

func (h *handler) Logout(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		http.Error(w, "token does not exist", http.StatusMethodNotAllowed)
		return
	}
	tokenID, _, ok := helpers.SplitToken(cookie.Value)

	if !ok {
		http.Error(w, "token ID does not exist", http.StatusMethodNotAllowed)
		return
	}

	if err := h.service.DeleteRefreshTokenByID(r.Context(), tokenID); err != nil {
		http.Error(w, "failed to revoke token", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now(),
	})
	json.Write(w, http.StatusOK, "logged out")

}
