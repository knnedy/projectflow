package domain

import "time"

type Member struct {
	ID             string     `json:"id"`
	Role           string     `json:"role"`
	UserID         string     `json:"userId"`
	OrganisationID string     `json:"organisationId"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

type MemberRepository interface {
	Create(member *Member) error
	GetByID(id string) (*Member, error)
	GetByUserAndOrg(userID, orgID string) (*Member, error)
	Update(member *Member) error
	Delete(id string) error
}
