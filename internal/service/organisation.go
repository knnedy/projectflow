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

type OrgService struct {
	db       *repository.Queries
	validate *validator.Validate
	trans    ut.Translator
}

func NewOrgService(db *repository.Queries) *OrgService {
	validate, trans := newValidator()
	return &OrgService{
		db:       db,
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

type UpdateMemberRoleInput struct {
	Role repository.MemberRole `validate:"required,oneof=OWNER ADMIN MEMBER"`
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

	// create organisation
	org, err := s.db.CreateOrganisation(ctx, repository.CreateOrganisationParams{
		ID:      pgtype.UUID{Bytes: orgID, Valid: true},
		Name:    input.Name,
		OwnerID: pgtype.UUID{Bytes: parsedUserID, Valid: true},
	})
	if err != nil {
		return repository.Organisation{}, domain.ErrDatabase
	}

	// automatically add creator as owner member
	_, err = s.db.CreateMember(ctx, repository.CreateMemberParams{
		ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Role:           repository.MemberRoleOWNER,
		UserID:         pgtype.UUID{Bytes: parsedUserID, Valid: true},
		OrganisationID: pgtype.UUID{Bytes: orgID, Valid: true},
	})
	if err != nil {
		return repository.Organisation{}, domain.ErrDatabase
	}

	return org, nil
}

func (s *OrgService) GetByID(ctx context.Context, orgID string) (repository.Organisation, error) {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// get organisation
	org, err := s.db.GetOrganisationById(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
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
	orgs, err := s.db.GetOrganisationsByUser(ctx, pgtype.UUID{Bytes: parsedUserID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return orgs, nil
}

func (s *OrgService) Update(ctx context.Context, orgID string, input UpdateOrgInput) (repository.Organisation, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Organisation{}, formatValidationError(err, s.trans)
	}

	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Organisation{}, domain.ErrNotFound
	}

	// update organisation
	updatedOrg, err := s.db.UpdateOrganisation(ctx, repository.UpdateOrganisationParams{
		ID:   pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		Name: input.Name,
	})
	if err != nil {
		return repository.Organisation{}, domain.ErrDatabase
	}

	return updatedOrg, nil
}

func (s *OrgService) Delete(ctx context.Context, orgID string) error {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return domain.ErrNotFound
	}

	// delete organisation
	if err := s.db.DeleteOrganisation(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}

func (s *OrgService) ListMembers(ctx context.Context, orgID string) ([]repository.Member, error) {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// get all members of the organisation
	members, err := s.db.GetMembersByOrg(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return members, nil
}

func (s *OrgService) UpdateMember(ctx context.Context, orgID string, memberID string, input UpdateMemberRoleInput) (repository.Member, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Member{}, formatValidationError(err, s.trans)
	}

	// parse memberID
	parsedMemberID, err := uuid.Parse(memberID)
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// get member to verify they belong to this org
	member, err := s.db.GetMemberById(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true})
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// verify member belongs to this org
	if member.OrganisationID.Bytes != parsedOrgID {
		return repository.Member{}, domain.ErrNotFound
	}

	// update member role
	updatedMember, err := s.db.UpdateMember(ctx, repository.UpdateMemberParams{
		ID:   pgtype.UUID{Bytes: parsedMemberID, Valid: true},
		Role: input.Role,
	})
	if err != nil {
		return repository.Member{}, domain.ErrDatabase
	}

	return updatedMember, nil
}

func (s *OrgService) DeleteMember(ctx context.Context, orgID string, memberID string) error {
	// parse memberID
	parsedMemberID, err := uuid.Parse(memberID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get member to verify they belong to this org
	member, err := s.db.GetMemberById(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return domain.ErrNotFound
	}

	// verify member belongs to this org
	if member.OrganisationID.Bytes != parsedOrgID {
		return domain.ErrNotFound
	}

	// enforce at least one owner or admin must remain
	count, err := s.db.GetOwnerAndAdminCountByOrg(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return domain.ErrDatabase
	}

	if count <= 1 && (member.Role == repository.MemberRoleOWNER || member.Role == repository.MemberRoleADMIN) {
		return domain.ErrCannotRemoveLastAdmin
	}

	// delete member
	if err := s.db.DeleteMember(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
