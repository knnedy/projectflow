package domain

import "errors"

var (
	// Resource not found
	ErrNotFound = errors.New("resource not found")

	// Input validation
	ErrValidation = errors.New("validation error")

	// Auth
	ErrUnauthorized       = errors.New("unauthorized")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Permissions
	ErrForbidden        = errors.New("forbidden")
	ErrNotOrgMember     = errors.New("user is not a member of the organisation")
	ErrNotProjectMember = errors.New("user is not a member of the project")

	// Unprocessable entity
	ErrAlreadyExists = errors.New("resource already exists")

	// Token errors
	ErrMissingToken   = errors.New("missing authorization token")
	ErrMalformedToken = errors.New("malformed authorization token")
	ErrInvalidToken   = errors.New("invalid or expired token")

	// Issue status transitions
	ErrInvalidStatusTransition = errors.New("invalid issue status transition")

	// Organization admin constraints
	ErrCannotRemoveLastAdmin = errors.New("cannot remove the last admin of an organization")
)

// ValidationError carries field-level detail on top of ErrValidation
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func (e *ValidationError) Unwrap() error {
	return ErrValidation
}

// NotFoundError carries the resource type for richer error messages
type NotFoundError struct {
	Resource string
}

func (e *NotFoundError) Error() string {
	return e.Resource + " not found"
}

func (e *NotFoundError) Unwrap() error {
	return ErrNotFound
}
