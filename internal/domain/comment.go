package domain

import "time"

type Comment struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	AuthorID  string     `json:"author_id"`
	IssueID   string     `json:"issue_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type CommentRepository interface {
	Create(comment *Comment) error
	GetByID(id string) (*Comment, error)
	Update(comment *Comment) error
	Delete(id string) error
	ListByIssueID(issueID string) ([]*Comment, error)
}
