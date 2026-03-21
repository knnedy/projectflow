package domain

import (
	"time"

	"github.com/knnedy/projectflow/internal/repository"
)

type Member struct {
	ID             string                `json:"id"`
	Role           repository.MemberRole `json:"role"`
	UserID         string                `json:"userId"`
	OrganisationID string                `json:"organisationId"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      *time.Time            `json:"updatedAt,omitempty"`
}

type MemberRepository interface {
	Create(member *Member) (*Member, error)
	GetByID(id string) (*Member, error)
	GetByUserAndOrg(userID, orgID string) (*Member, error)
	Update(member *Member) (*Member, error)
	Delete(id string) error
}
