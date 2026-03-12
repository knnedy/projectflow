package domain

import "time"

type Organisation struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
	OwnerID   string     `json:"ownerId"`
}

type OrganisationRepository interface {
	Create(org *Organisation) error
	GetByID(id string) (*Organisation, error)
	Update(org *Organisation) error
	Delete(id string) error
}
