package domain

import (
	"time"

	"github.com/knnedy/projectflow/internal/repository"
)

type Issue struct {
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description *string                  `json:"description,omitempty"`
	Status      repository.IssueStatus   `json:"status"`
	Priority    repository.IssuePriority `json:"priority"`
	ProjectID   string                   `json:"projectId"`
	ReporterID  string                   `json:"reporterId"`
	AssigneeID  *string                  `json:"assigneeId,omitempty"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   *time.Time               `json:"updatedAt,omitempty"`
}

type IssueRepository interface {
	Create(issue *Issue) (*Issue, error)
	GetByID(id string) (*Issue, error)
	ListByProject(projectID string) ([]*Issue, error)
	Update(issue *Issue) (*Issue, error)
	Delete(id string) error
}
