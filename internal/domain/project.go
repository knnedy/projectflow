package domain

import "time"

type Project struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	OrganisationID string     `json:"organisationId"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      *time.Time `json:"updatedAt"`
}

type ProjectRepository interface {
	Create(project *Project) error
	GetByID(id string) (*Project, error)
	Update(project *Project) error
	Delete(id string) error
}
