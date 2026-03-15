package token

import (
	"time"

	"github.com/google/uuid"
)

const refreshTokenDuration = 7 * 24 * time.Hour

type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
}

func (tm *TokenManager) GenerateRefreshToken(userID uuid.UUID) (RefreshToken, error) {
	now := time.Now()
	return RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: now.Add(refreshTokenDuration),
		CreatedAt: now,
	}, nil
}
