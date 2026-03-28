package service

import (
	"context"
	"log/slog"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
)

type IssueService struct {
	db       *repository.Queries
	activity *ActivityService
	validate *validator.Validate
	trans    ut.Translator
}

func NewIssueService(db *repository.DB, activity *ActivityService) *IssueService {
	validate, trans := newValidator()
	return &IssueService{
		db:       db.Queries(),
		activity: activity,
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

// logActivity fires activity logging in a goroutine so it never blocks the response
func (s *IssueService) logActivity(input LogInput) {
	go func() {
		if err := s.activity.Log(context.Background(), input); err != nil {
			slog.Error("failed to log activity",
				"error", err,
				"action", string(input.Action),
				"entity_type", string(input.EntityType),
				"entity_id", input.EntityID,
				"actor_id", input.ActorID,
			)
		}
	}()
}

func (s *IssueService) Create(ctx context.Context, projectID, reporterID, actorID string, input CreateIssueInput) (repository.Issue, error) {
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
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
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

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}
	actor, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
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

	// log activity
	s.logActivity(LogInput{
		OrgID:      project.OrganisationID.String(),
		EntityType: repository.ActivityEntityTypeISSUE,
		EntityID:   issue.ID.String(),
		Action:     repository.ActivityActionCREATED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"issue_title": issue.Title,
			"created_by":  actor.Name,
		},
	})

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

func (s *IssueService) UpdateDetails(ctx context.Context, projectID, issueID, actorID string, input UpdateIssueDetailsInput) (repository.Issue, error) {
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
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
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

	// verify issue belongs to this project
	if issue.ProjectID.String() != project.ID.String() {
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

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}
	actor, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
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

	// log activity
	s.logActivity(LogInput{
		OrgID:      project.OrganisationID.String(),
		EntityType: repository.ActivityEntityTypeISSUE,
		EntityID:   issue.ID.String(),
		Action:     repository.ActivityActionUPDATED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"issue_title": issue.Title,
			"updated_by":  actor.Name,
		},
	})

	return updated, nil
}

func (s *IssueService) UpdateStatus(ctx context.Context, projectID, issueID, actorID string, input UpdateIssueStatusInput, memberRole repository.MemberRole) (repository.Issue, error) {
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
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
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

	// verify issue belongs to this project
	if issue.ProjectID.String() != project.ID.String() {
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

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}
	actor, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return repository.Issue{}, domain.ErrNotFound
	}

	// update issue status
	updated, err := s.db.UpdateIssueStatus(ctx, repository.UpdateIssueStatusParams{
		ID:     pgtype.UUID{Bytes: parsedIssueID, Valid: true},
		Status: input.Status,
	})
	if err != nil {
		return repository.Issue{}, domain.ErrDatabase
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      project.OrganisationID.String(),
		EntityType: repository.ActivityEntityTypeISSUE,
		EntityID:   issue.ID.String(),
		Action:     repository.ActivityActionSTATUSCHANGED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"issue_title": issue.Title,
			"from":        string(issue.Status),
			"to":          string(input.Status),
			"updated_by":  actor.Name,
		},
	})

	return updated, nil
}

func (s *IssueService) Delete(ctx context.Context, projectID, issueID, actorID string) error {
	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return domain.ErrNotFound
	}

	// verify project exists
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

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

	// verify issue belongs to this project
	if issue.ProjectID.String() != project.ID.String() {
		return domain.ErrNotFound
	}

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return domain.ErrNotFound
	}
	actor, err := s.db.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// delete issue
	if err := s.db.DeleteIssue(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      project.OrganisationID.String(),
		EntityType: repository.ActivityEntityTypeISSUE,
		EntityID:   issue.ID.String(),
		Action:     repository.ActivityActionDELETED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"issue_title": issue.Title,
			"deleted_by":  actor.Name,
		},
	})

	return nil
}
