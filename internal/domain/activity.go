package domain

import "time"

type ActivityLog struct {
	ID        string    `json:"id"`
	Action    string    `json:"action"`
	UserID    string    `json:"user_id"`
	ProjectID string    `json:"project_id"`
	TargetID  string    `json:"target_id"`
	Timestamp time.Time `json:"timestamp"`
}

type ActivityLogRepository interface {
	Create(activity *ActivityLog) error
	ListByProjectID(projectID string) ([]*ActivityLog, error)
}
