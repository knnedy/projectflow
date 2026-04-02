package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/middleware"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/response"
	"github.com/knnedy/projectflow/internal/service"
)

type UserHandler struct {
	user *service.UserService
}

func NewUserHandler(user *service.UserService) *UserHandler {
	return &UserHandler{user: user}
}

// UserResponse is the public representation of a user
type UserResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Email     string  `json:"email"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

func toUserResponse(user repository.User) UserResponse {
	var updatedAt *string
	if user.UpdatedAt.Valid {
		t := user.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: updatedAt,
	}
}

// GetMe godoc
// @Summary Get authenticated user
// @Tags users
// @Produce json
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get user
	user, err := h.user.GetMe(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdateMe godoc
// @Summary Update authenticated user profile
// @Tags users
// @Accept json
// @Produce json
// @Param body body service.UpdateProfileInput true "Update profile input"
// @Success 200 {object} UserResponse
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /users/me [patch]
func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// decode request body
	var input service.UpdateProfileInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update user profile
	user, err := h.user.UpdateMe(r.Context(), userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toUserResponse(user))
}

// UpdatePassword godoc
// @Summary Update authenticated user password
// @Tags users
// @Accept json
// @Produce json
// @Param body body service.UpdatePasswordInput true "Update password input"
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /users/me/password [patch]
func (h *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// decode request body
	var input service.UpdatePasswordInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update password
	if err := h.user.UpdatePassword(r.Context(), userID, input); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}

// DeleteMe godoc
// @Summary Delete authenticated user
// @Tags users
// @Produce json
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Router /users/me [delete]
func (h *UserHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// delete user
	if err := h.user.DeleteMe(r.Context(), userID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}
