package service

import (
	"context"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
)

type IssueService struct {
	db       *repository.Queries
	validate *validator.Validate
	trans    ut.Translator
}

func NewIssueService(db *repository.Queries) *IssueService {
	validate, trans := newValidator()
	return &IssueService{
		db:       db,
		validate: validate,
		trans:    trans,
	}
}

type CreateIssueInput struct {
	Title       string                   `validate:"required,min=2,max=200"`
	Description string                   `validate:"max=2000"`
	Priority    repository.IssuePriority `validate:"required,oneof=NO_PRIORITY LOW MEDIUM HIGH"`
	AssigneeID  string                   `validate:"omitempty,uuid"`
}

type UpdateIssueDetailsInput struct {
	Title       string                   `validate:"required,min=2,max=200"`
	Description string                   `validate:"max=2000"`
	Priority    repository.IssuePriority `validate:"required,oneof=NO_PRIORITY LOW MEDIUM HIGH"`
	AssigneeID  string                   `validate:"omitempty,uuid"`
}

type UpdateIssueStatusInput struct {
	Status repository.IssueStatus `validate:"required,oneof=BACKLOG TODO IN_PROGRESS IN_REVIEW DONE CANCELLED"`
}

// validTransitions defines allowed status transitions
var validTransitions = map[repository.IssueStatus][]repository.IssueStatus{
	repository.IssueStatusBACKLOG:    {repository.IssueStatusTODO, repository.IssueStatusCANCELLED},
	repository.IssueStatusTODO:       {repository.IssueStatusINPROGRESS, repository.IssueStatusCANCELLED},
	repository.IssueStatusINPROGRESS: {repository.IssueStatusINREVIEW, repository.IssueStatusCANCELLED},
	repository.IssueStatusINREVIEW:   {repository.IssueStatusDONE, repository.IssueStatusINPROGRESS, repository.IssueStatusCANCELLED},
	repository.IssueStatusDONE:       {}, // terminal
	repository.IssueStatusCANCELLED:  {}, // terminal
}

// adminOnlyTransitions defines transitions only owners and admins can make
var adminOnlyTransitions = map[repository.IssueStatus]bool{
	repository.IssueStatusTODO:      true, // BACKLOG → TODO: admin assigns
	repository.IssueStatusDONE:      true, // IN_REVIEW → DONE: admin approves
	repository.IssueStatusCANCELLED: true, // any → CANCELLED: admin only
}

func isValidTransition(from, to repository.IssueStatus) bool {
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

func isAdminOnlyTransition(to repository.IssueStatus) bool {
	return adminOnlyTransitions[to]
}

func (s *IssueService) Create(ctx context.Context, projectID, reporterID string, input CreateIssueInput) (repository.Issue, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Issue{}, formatValidationError(err, s.trans)
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// verify project exists
	_, err = s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// parse reporter ID
	parsedReporterID, err := uuid.Parse(reporterID)
	if err != nil {
		return repository.Issue{}, domain.ErrUnauthorized
	}

	// handle optional assignee
	var assigneeID pgtype.UUID
	if input.AssigneeID != "" {
		parsedAssigneeID, err := uuid.Parse(input.AssigneeID)
		if err != nil {
			return repository.Issue{}, &domain.ValidationError{Field: "assignee_id", Message: "must be a valid uuid"}
		}
		assigneeID = pgtype.UUID{Bytes: parsedAssigneeID, Valid: true}
	}

	// create issue — new issues always start at BACKLOG
	issue, err := s.db.CreateIssue(ctx, repository.CreateIssueParams{
		ID:          pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Title:       input.Title,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Status:      repository.IssueStatusBACKLOG,
		Priority:    input.Priority,
		ProjectID:   pgtype.UUID{Bytes: parsedProjectID, Valid: true},
		ReporterID:  pgtype.UUID{Bytes: parsedReporterID, Valid: true},
		AssigneeID:  assigneeID,
	})
	if err != nil {
		return repository.Issue{}, domain.ErrDatabase
	}

	return issue, nil
}

func (s *IssueService) GetByID(ctx context.Context, projectID, issueID string) (repository.Issue, error) {
	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// get issue
	issue, err := s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// verify issue belongs to this project
	if issue.ProjectID.Bytes != parsedProjectID {
		return repository.Issue{}, domain.ErrNotFound
	}

	return issue, nil
}

func (s *IssueService) List(ctx context.Context, projectID string) ([]repository.Issue, error) {
	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// list issues by project
	issues, err := s.db.ListIssuesByProject(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return issues, nil
}

func (s *IssueService) UpdateDetails(ctx context.Context, projectID, issueID string, input UpdateIssueDetailsInput) (repository.Issue, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Issue{}, formatValidationError(err, s.trans)
	}

	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// get issue to verify it belongs to this project
	issue, err := s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// verify issue belongs to this project
	if issue.ProjectID.Bytes != parsedProjectID {
		return repository.Issue{}, domain.ErrNotFound
	}

	// handle optional assignee
	var assigneeID pgtype.UUID
	if input.AssigneeID != "" {
		parsedAssigneeID, err := uuid.Parse(input.AssigneeID)
		if err != nil {
			return repository.Issue{}, &domain.ValidationError{Field: "assignee_id", Message: "must be a valid uuid"}
		}
		assigneeID = pgtype.UUID{Bytes: parsedAssigneeID, Valid: true}
	}

	// update issue details
	updated, err := s.db.UpdateIssueDetails(ctx, repository.UpdateIssueDetailsParams{
		ID:          pgtype.UUID{Bytes: parsedIssueID, Valid: true},
		Title:       input.Title,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
		Priority:    input.Priority,
		AssigneeID:  assigneeID,
	})
	if err != nil {
		return repository.Issue{}, domain.ErrDatabase
	}

	return updated, nil
}

func (s *IssueService) UpdateStatus(ctx context.Context, projectID, issueID string, input UpdateIssueStatusInput, memberRole repository.MemberRole) (repository.Issue, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Issue{}, formatValidationError(err, s.trans)
	}

	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// get issue to verify it belongs to this project and get current status
	issue, err := s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// verify issue belongs to this project
	if issue.ProjectID.Bytes != parsedProjectID {
		return repository.Issue{}, domain.ErrNotFound
	}

	// enforce strict status transition rules
	if !isValidTransition(issue.Status, input.Status) {
		return repository.Issue{}, domain.ErrInvalidStatusTransition
	}

	// enforce role based transition rules
	isAdmin := memberRole == repository.MemberRoleOWNER || memberRole == repository.MemberRoleADMIN
	if isAdminOnlyTransition(input.Status) && !isAdmin {
		return repository.Issue{}, domain.ErrForbidden
	}

	// update issue status
	updated, err := s.db.UpdateIssueStatus(ctx, repository.UpdateIssueStatusParams{
		ID:     pgtype.UUID{Bytes: parsedIssueID, Valid: true},
		Status: input.Status,
	})
	if err != nil {
		return repository.Issue{}, domain.ErrDatabase
	}

	return updated, nil
}

func (s *IssueService) Delete(ctx context.Context, projectID, issueID string) error {
	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get issue to verify it belongs to this project
	issue, err := s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return domain.ErrNotFound
	}

	// verify issue belongs to this project
	if issue.ProjectID.Bytes != parsedProjectID {
		return domain.ErrNotFound
	}

	// delete issue
	if err := s.db.DeleteIssue(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
