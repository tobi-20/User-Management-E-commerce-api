package auth

import (
	"ecom/cmd/app/globals"
	"ecom/cmd/helpers"
	"time"

	"ecom/internal/err"
	"ecom/internal/json"
	"log"

	"net/http"
)

var e *err.Error

type handler struct {
	service AuthService
}

func NewHandler(service AuthService) *handler {
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

	if err := h.service.SendConfirmUserTokenToEmail(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusCreated, map[string]string{
		"message": "user account created",
	})

}

func (h *handler) VerifyUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	selector := r.URL.Query().Get("selector")
	verifier := r.URL.Query().Get("verifier")
	if selector == "" {
		http.Error(w, "token incomplete", http.StatusBadRequest)
		return
	}
	if verifier == "" {
		http.Error(w, "token missing", http.StatusBadRequest)
		return
	}

	user, err := h.service.ValidateVerification(r.Context(), verifier, selector)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, raw, err := h.service.IssueAuthTokens(r.Context(), helpers.ToUserID(user))
	if err != nil {
		http.Error(w, "token not generated", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    raw,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(globals.RefreshTokenExpiry),
	})

	json.Write(w, http.StatusCreated, map[string]interface{}{
		"name": user.Name,
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

	accessToken, rawRefreshToken, err := h.service.Login(r.Context(), req)
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

func (h *handler) RefreshToken(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cookie, err := r.Cookie("refresh_token")

	if err != nil {
		http.Error(w, "token missing", http.StatusExpectationFailed)
		return

	}
	rawToken := cookie.Value

	if rawToken == "" {
		http.Error(w, "token missing", http.StatusExpectationFailed)
		return
	}

	newAccessToken, newRaw, err := h.service.IssueRefreshToken(r.Context(), rawToken)
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRaw,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})

	json.Write(w, http.StatusOK, map[string]string{
		"access_token": newAccessToken})

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

	if err = h.service.ConsumeCookie(r.Context(), cookie); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

func (h *handler) SendResetTokenToEmail(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := json.Read(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.service.SendResetTokenToEmail(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.Write(w, http.StatusCreated, map[string]string{
		"message": "check your email for the verification link",
	})
}

func (h *handler) ValidateResetPassword(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	selector := r.URL.Query().Get("selector")
	verifier := r.URL.Query().Get("verifier")

	if _, err := h.service.ValidateResetPasswordTokens(r.Context(), selector, verifier); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.Write(w, http.StatusAccepted, map[string]string{
		"message": "Set new password",
	})
}

func (h *handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var newPassParams ResetPassWordReq
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := json.Read(r, &newPassParams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.service.ResetPassword(r.Context(), newPassParams)
	if err := json.Read(r, &newPassParams); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

}
