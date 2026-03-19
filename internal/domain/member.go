package domain

import "time"

type MemberRole string

const (
	MemberRoleOwner  MemberRole = "OWNER"
	MemberRoleAdmin  MemberRole = "ADMIN"
	MemberRoleMember MemberRole = "MEMBER"
)

type Member struct {
	ID             string     `json:"id"`
	Role           MemberRole `json:"role"`
	UserID         string     `json:"userId"`
	OrganisationID string     `json:"organisationId"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

type MemberRepository interface {
	Create(member *Member) (*Member, error)
	GetByID(id string) (*Member, error)
	GetByUserAndOrg(userID, orgID string) (*Member, error)
	Update(member *Member) (*Member, error)
	Delete(id string) error
}
