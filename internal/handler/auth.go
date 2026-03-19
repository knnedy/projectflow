package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// UserResponse is the public representation of a user — never expose repository.User directly
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// authDataResponse is the data envelope returned on register, login and refresh
type authDataResponse struct {
	User        UserResponse `json:"user"`
	AccessToken string       `json:"access_token"`
}

func toUserResponse(user repository.User) UserResponse {
	return UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
	}
}

func toAuthDataResponse(result service.AuthResult) authDataResponse {
	return authDataResponse{
		User:        toUserResponse(result.User),
		AccessToken: result.AccessToken,
	}
}

func setRefreshTokenCookie(w http.ResponseWriter, refreshToken string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 60 * 60, // 7 days in seconds — matches refresh token duration
	})
}

func clearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
	})
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body service.RegisterInput true "Register input"
// @Success 201 {object} authDataResponse
// @Failure 409 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, err)
		return
	}

	result, err := h.auth.Register(r.Context(), input)
	if err != nil {
		WriteError(w, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken)
	WriteJSON(w, http.StatusCreated, toAuthDataResponse(result))
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body service.LoginInput true "Login input"
// @Success 200 {object} authDataResponse
// @Failure 401 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteError(w, err)
		return
	}

	result, err := h.auth.Login(r.Context(), input)
	if err != nil {
		WriteError(w, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken)
	WriteJSON(w, http.StatusOK, toAuthDataResponse(result))
}

// RefreshAccessToken godoc
// @Summary Refresh access token
// @Tags auth
// @Produce json
// @Success 200 {object} authDataResponse
// @Failure 401 {object} errorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	// read refresh token from httponly cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		WriteError(w, err)
		return
	}

	result, err := h.auth.RefreshAccessToken(r.Context(), cookie.Value)
	if err != nil {
		WriteError(w, err)
		return
	}

	setRefreshTokenCookie(w, result.RefreshToken)
	WriteJSON(w, http.StatusOK, toAuthDataResponse(result))
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Produce json
// @Success 200
// @Failure 401 {object} errorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// read refresh token from httponly cookie
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		WriteError(w, err)
		return
	}

	if err := h.auth.Logout(r.Context(), cookie.Value); err != nil {
		WriteError(w, err)
		return
	}

	clearRefreshTokenCookie(w)
	WriteJSON(w, http.StatusOK, nil)
}
