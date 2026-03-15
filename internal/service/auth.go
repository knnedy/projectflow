package service

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/token"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db       *repository.Queries
	tokens   *token.TokenManager
	validate *validator.Validate
}

func NewAuthService(db *repository.Queries, tokens *token.TokenManager) *AuthService {
	return &AuthService{
		db:       db,
		tokens:   tokens,
		validate: validator.New(),
	}
}

type RegisterInput struct {
	Name     string `validate:"required,min=2,max=100"`
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=8,max=100"`
}

type LoginInput struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required"`
}

type AuthResult struct {
	User         repository.User
	AccessToken  string
	RefreshToken string
}

// generateAuthTokens is a shared helper that generates both tokens and saves the refresh token
func (s *AuthService) generateAuthTokens(ctx context.Context, user repository.User) (AuthResult, error) {
	// generate access token
	accessToken, err := s.tokens.GenerateAccessToken(user.ID.String())
	if err != nil {
		return AuthResult{}, err
	}

	// generate refresh token
	refreshToken, err := s.tokens.GenerateRefreshToken(
		uuid.UUID(user.ID.Bytes),
	)
	if err != nil {
		return AuthResult{}, err
	}

	// save refresh token to DB
	_, err = s.db.CreateRefreshToken(ctx, repository.CreateRefreshTokenParams{
		ID:        pgtype.UUID{Bytes: refreshToken.ID, Valid: true},
		UserID:    pgtype.UUID{Bytes: refreshToken.UserID, Valid: true},
		Token:     refreshToken.Token,
		ExpiresAt: pgtype.Timestamp{Time: refreshToken.ExpiresAt, Valid: true},
	})
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		firstErr := validationErrors[0]
		return AuthResult{}, &domain.ValidationError{
			Field:   firstErr.Field(),
			Message: firstErr.Tag(),
		}
	}

	// check if email already exists
	_, err := s.db.GetUserByEmail(ctx, input.Email)
	if err == nil {
		return AuthResult{}, domain.ErrEmailAlreadyExists
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	// create user
	user, err := s.db.CreateUser(ctx, repository.CreateUserParams{
		ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hashedPassword),
	})
	if err != nil {
		return AuthResult{}, err
	}

	return s.generateAuthTokens(ctx, user)
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		firstErr := validationErrors[0]
		return AuthResult{}, &domain.ValidationError{
			Field:   firstErr.Field(),
			Message: firstErr.Tag(),
		}
	}

	// get user by email
	user, err := s.db.GetUserByEmail(ctx, input.Email)
	if err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}

	return s.generateAuthTokens(ctx, user)
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (AuthResult, error) {
	// validate refresh token exists and is not revoked or expired
	dbToken, err := s.db.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		return AuthResult{}, domain.ErrInvalidToken
	}

	// revoke current refresh token - rotation
	err = s.db.RevokeRefreshToken(ctx, refreshToken)
	if err != nil {
		return AuthResult{}, err
	}

	// get user
	user, err := s.db.GetUserById(ctx, dbToken.UserID)
	if err != nil {
		return AuthResult{}, domain.ErrNotFound
	}

	return s.generateAuthTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	// revoke refresh token
	return s.db.RevokeRefreshToken(ctx, refreshToken)
}
