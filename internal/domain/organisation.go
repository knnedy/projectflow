package domain

import "time"

type Organisation struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	OwnerID   string     `json:"ownerId"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type OrganisationRepository interface {
	Create(org *Organisation) (*Organisation, error)
	GetByID(id string) (*Organisation, error)
	GetByOwner(ownerID string) ([]*Organisation, error)
	Update(org *Organisation) (*Organisation, error)
	Delete(id string) error
}
