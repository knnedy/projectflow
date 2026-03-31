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

type OrgService struct {
	db       *repository.DB
	queries  *repository.Queries
	activity *ActivityService
	validate *validator.Validate
	trans    ut.Translator
}

func NewOrgService(db *repository.DB, activity *ActivityService) *OrgService {
	validate, trans := newValidator()
	return &OrgService{
		db:       db,
		queries:  db.Queries(),
		activity: activity,
		validate: validate,
		trans:    trans,
	}
}

type CreateOrgInput struct {
	Name string `validate:"required,min=2,max=100"`
}

type UpdateOrgInput struct {
	Name string `validate:"required,min=2,max=100"`
}

// logActivity fires activity logging in a goroutine so it never blocks the response
func (s *OrgService) logActivity(input LogInput) {
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

func (s *OrgService) Create(ctx context.Context, userID string, input CreateOrgInput) (repository.Organisation, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Organisation{}, formatValidationError(err, s.trans)
	}

	// parse userID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return repository.Organisation{}, domain.ErrUnauthorized
	}

	orgID := uuid.New()
	var org repository.Organisation

	// fetch the actor name for the activity log
	actor, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: parsedUserID, Valid: true})
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// create organisation and add creator as owner in a single transaction
	err = s.db.WithTransaction(ctx, func(q *repository.Queries) error {
		var err error

		// create organisation
		org, err = q.CreateOrganisation(ctx, repository.CreateOrganisationParams{
			ID:      pgtype.UUID{Bytes: orgID, Valid: true},
			Name:    input.Name,
			OwnerID: pgtype.UUID{Bytes: parsedUserID, Valid: true},
		})
		if err != nil {
			return domain.ErrDatabase
		}

		// automatically add creator as owner member
		_, err = q.CreateMember(ctx, repository.CreateMemberParams{
			ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Role:           repository.MemberRoleOWNER,
			UserID:         pgtype.UUID{Bytes: parsedUserID, Valid: true},
			OrganisationID: pgtype.UUID{Bytes: orgID, Valid: true},
		})
		if err != nil {
			return domain.ErrDatabase
		}

		return nil
	})
	if err != nil {
		return repository.Organisation{}, err
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      org.ID.String(),
		EntityType: repository.ActivityEntityTypeORGANISATION,
		EntityID:   org.ID.String(),
		Action:     repository.ActivityActionCREATED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"organisation_name": org.Name,
			"created_by":        actor.Name,
		},
	})

	return org, nil
}

func (s *OrgService) GetByID(ctx context.Context, orgID string) (repository.Organisation, error) {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// get organisation
	org, err := s.queries.GetOrganisationById(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	return org, nil
}

func (s *OrgService) List(ctx context.Context, userID string) ([]repository.Organisation, error) {
	// parse userID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	// get all organisations the user belongs to
	orgs, err := s.queries.GetOrganisationsByUser(ctx, pgtype.UUID{Bytes: parsedUserID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return orgs, nil
}

func (s *OrgService) Update(ctx context.Context, orgID, actorID string, input UpdateOrgInput) (repository.Organisation, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Organisation{}, formatValidationError(err, s.trans)
	}

	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}
	actor, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// update organisation
	updatedOrg, err := s.queries.UpdateOrganisation(ctx, repository.UpdateOrganisationParams{
		ID:   pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		Name: input.Name,
	})
	if err != nil {
		return repository.Organisation{}, domain.ErrDatabase
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      updatedOrg.ID.String(),
		EntityType: repository.ActivityEntityTypeORGANISATION,
		EntityID:   updatedOrg.ID.String(),
		Action:     repository.ActivityActionUPDATED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"organisation_name": updatedOrg.Name,
			"updated_by":        actor.Name,
		},
	})

	return updatedOrg, nil
}

func (s *OrgService) Delete(ctx context.Context, orgID string) error {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return domain.ErrNotFound
	}

	// delete organisation — cascades to members, projects, issues via FK constraints
	if err := s.queries.DeleteOrganisation(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
