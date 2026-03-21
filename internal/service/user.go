package service

import (
	"context"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	db       *repository.Queries
	validate *validator.Validate
	trans    ut.Translator
}

func NewUserService(db *repository.Queries) *UserService {
	validate, trans := newValidator()
	return &UserService{
		db:       db,
		validate: validate,
		trans:    trans,
	}
}

type UpdateProfileInput struct {
	Name  string `validate:"required,min=2,max=100"`
	Email string `validate:"required,email"`
}

type UpdatePasswordInput struct {
	CurrentPassword string `validate:"required,min=8"`
	NewPassword     string `validate:"required,min=8,max=100,nefield=CurrentPassword"`
}

func (s *UserService) GetMe(ctx context.Context, userID string) (repository.User, error) {
	// parse user ID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return repository.User{}, domain.ErrNotFound
	}

	// get user from DB
	user, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		return repository.User{}, domain.ErrNotFound
	}

	return user, nil
}

func (s *UserService) UpdateMe(ctx context.Context, userID string, input UpdateProfileInput) (repository.User, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.User{}, formatValidationError(err, s.trans)
	}

	// parse user ID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return repository.User{}, domain.ErrNotFound
	}

	// check if email is already taken by another user
	existing, err := s.db.GetUserByEmail(ctx, input.Email)
	if err == nil && existing.ID.Bytes != parsedID {
		return repository.User{}, domain.ErrAlreadyExists
	}

	// update profile
	user, err := s.db.UpdateUserProfile(ctx, repository.UpdateUserProfileParams{
		ID:    pgtype.UUID{Bytes: parsedID, Valid: true},
		Name:  input.Name,
		Email: input.Email,
	})
	if err != nil {
		return repository.User{}, domain.ErrDatabase
	}

	return user, nil
}

func (s *UserService) UpdatePassword(ctx context.Context, userID string, input UpdatePasswordInput) error {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return formatValidationError(err, s.trans)
	}

	// parse user ID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get user from DB
	user, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.CurrentPassword))
	if err != nil {
		return domain.ErrInvalidCredentials
	}

	// hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return domain.ErrInternal
	}

	// update password in DB
	_, err = s.db.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
		ID:       pgtype.UUID{Bytes: parsedID, Valid: true},
		Password: string(hashedPassword),
	})
	if err != nil {
		return domain.ErrDatabase
	}

	return nil
}

func (s *UserService) DeleteMe(ctx context.Context, userID string) error {
	// parse user ID
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrNotFound
	}

	// delete user from DB
	if err := s.db.DeleteUser(ctx, pgtype.UUID{Bytes: parsedID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
