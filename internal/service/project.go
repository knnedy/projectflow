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

type ProjectService struct {
	db       *repository.Queries
	validate *validator.Validate
	trans    ut.Translator
}

func NewProjectService(db *repository.Queries) *ProjectService {
	validate, trans := newValidator()
	return &ProjectService{
		db:       db,
		validate: validate,
		trans:    trans,
	}
}

type CreateProjectInput struct {
	Name        string `validate:"required,min=2,max=100"`
	Description string `validate:"max=500"`
}

type UpdateProjectInput struct {
	Name        string `validate:"required,min=2,max=100"`
	Description string `validate:"max=500"`
}

func (s *ProjectService) Create(ctx context.Context, orgID string, input CreateProjectInput) (repository.Project, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Project{}, formatValidationError(err, s.trans)
	}

	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// create project
	project, err := s.db.CreateProject(ctx, repository.CreateProjectParams{
		ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Name:           input.Name,
		Description:    pgtype.Text{String: input.Description, Valid: input.Description != ""},
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
	})
	if err != nil {
		return repository.Project{}, domain.ErrDatabase
	}

	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, orgID, projectID string) (repository.Project, error) {
	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// get project
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// verify project belongs to this org
	if project.OrganisationID.Bytes != parsedOrgID {
		return repository.Project{}, domain.ErrNotFound
	}

	return project, nil
}

func (s *ProjectService) List(ctx context.Context, orgID string) ([]repository.Project, error) {
	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// get all projects for this org
	projects, err := s.db.GetProjectsByOrganisation(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return projects, nil
}

func (s *ProjectService) Update(ctx context.Context, orgID, projectID string, input UpdateProjectInput) (repository.Project, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Project{}, formatValidationError(err, s.trans)
	}

	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// get project to verify it belongs to this org
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Project{}, domain.ErrNotFound
	}

	// verify project belongs to this org
	if project.OrganisationID.Bytes != parsedOrgID {
		return repository.Project{}, domain.ErrNotFound
	}

	// update project
	updated, err := s.db.UpdateProject(ctx, repository.UpdateProjectParams{
		ID:          pgtype.UUID{Bytes: parsedProjectID, Valid: true},
		Name:        input.Name,
		Description: pgtype.Text{String: input.Description, Valid: input.Description != ""},
	})
	if err != nil {
		return repository.Project{}, domain.ErrDatabase
	}

	return updated, nil
}

func (s *ProjectService) Delete(ctx context.Context, orgID, projectID string) error {
	// parse project ID
	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get project to verify it belongs to this org
	project, err := s.db.GetProjectById(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// parse org ID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return domain.ErrNotFound
	}

	// verify project belongs to this org
	if project.OrganisationID.Bytes != parsedOrgID {
		return domain.ErrNotFound
	}

	// delete project
	if err := s.db.DeleteProject(ctx, pgtype.UUID{Bytes: parsedProjectID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
