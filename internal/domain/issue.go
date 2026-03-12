package domain

import "time"

type IssueStatus string

const (
	IssueStatusOpen       IssueStatus = "open"
	IssueStatusInProgress IssueStatus = "in_progress"
	IssueStatusClosed     IssueStatus = "closed"
)

type IssuePriority string

const (
	IssuePriorityLow    IssuePriority = "low"
	IssuePriorityMedium IssuePriority = "medium"
	IssuePriorityHigh   IssuePriority = "high"
)

type Issue struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description *string       `json:"description,omitempty"`
	Status      IssueStatus   `json:"status"`
	Priority    IssuePriority `json:"priority"`
	ProjectID   string        `json:"projectId"`
	ReporterID  string        `json:"reporterId"`
	AssigneeID  *string       `json:"assigneeId,omitempty"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   *time.Time    `json:"updatedAt,omitempty"`
}

type IssueRepository interface {
	Create(issue *Issue) error
	GetByID(id string) (*Issue, error)
	ListByProject(projectID string) ([]*Issue, error)
	Update(issue *Issue) error
	Delete(id string) error
}
