package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/knnedy/projectflow/internal/domain"
)

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

type errorResponse struct {
	Success bool        `json:"success"`
	Error   errorDetail `json:"error"`
}

// writeJSON writes a success response with the given status code and data
func writeJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(successResponse{
		Success: true,
		Data:    data,
	})
}

// writeError writes an error response mapping domain errors to HTTP status codes
func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")

	var status int
	var detail errorDetail

	switch {
	case errors.Is(err, domain.ErrValidation):
		status = http.StatusUnprocessableEntity
		detail.Code = "VALIDATION_ERROR"
		var valErr *domain.ValidationError
		if errors.As(err, &valErr) {
			detail.Message = valErr.Message
			detail.Field = valErr.Field
		}

	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		detail.Code = "CONFLICT"
		detail.Message = "resource already exists"

	case errors.Is(err, domain.ErrInvalidCredentials):
		status = http.StatusUnauthorized
		detail.Code = "INVALID_CREDENTIALS"
		detail.Message = "invalid email or password"

	case errors.Is(err, domain.ErrInvalidToken):
		status = http.StatusUnauthorized
		detail.Code = "INVALID_TOKEN"
		detail.Message = "invalid or expired token"

	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
		detail.Code = "UNAUTHORIZED"
		detail.Message = "unauthorized"

	case errors.Is(err, domain.ErrForbidden):
		status = http.StatusForbidden
		detail.Code = "FORBIDDEN"
		detail.Message = "insufficient permissions"

	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		detail.Code = "NOT_FOUND"
		detail.Message = "resource not found"

	case errors.Is(err, domain.ErrNotOrgMember):
		status = http.StatusForbidden
		detail.Code = "FORBIDDEN"
		detail.Message = "user is not a member of this organization"

	case errors.Is(err, domain.ErrInvalidStatusTransition):
		status = http.StatusUnprocessableEntity
		detail.Code = "INVALID_STATUS_TRANSITION"
		detail.Message = "invalid issue status transition"

	case errors.Is(err, domain.ErrCannotRemoveLastAdmin):
		status = http.StatusUnprocessableEntity
		detail.Code = "CANNOT_REMOVE_LAST_ADMIN"
		detail.Message = "cannot remove the last admin of an organization"

	default:
		status = http.StatusInternalServerError
		detail.Code = "INTERNAL_SERVER_ERROR"
		detail.Message = "an unexpected error occurred"
	}

	w.WriteHeader(status)
	json.NewEncoder(w).Encode(errorResponse{
		Success: false,
		Error:   detail,
	})
}
