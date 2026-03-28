package service

import (
	"context"
	"log/slog"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
)

const invitationDuration = 7 * 24 * time.Hour

type MemberService struct {
	db       *repository.DB
	queries  *repository.Queries
	activity *ActivityService
	validate *validator.Validate
	trans    ut.Translator
}

func NewMemberService(db *repository.DB, activity *ActivityService) *MemberService {
	validate, trans := newValidator()
	return &MemberService{
		db:       db,
		queries:  db.Queries(),
		activity: activity,
		validate: validate,
		trans:    trans,
	}
}

// logActivity fires activity logging in a goroutine so it never blocks the response
func (s *MemberService) logActivity(input LogInput) {
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

type InviteMemberInput struct {
	Email string                `validate:"required,email"`
	Role  repository.MemberRole `validate:"required,oneof=ADMIN MEMBER"`
}

type UpdateMemberRoleInput struct {
	Role repository.MemberRole `validate:"required,oneof=OWNER ADMIN MEMBER"`
}

type InviteResult struct {
	Invitation repository.Invitation
	InviteLink string
}

func (s *MemberService) ListMembers(ctx context.Context, orgID string) ([]repository.Member, error) {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// get all members of the organisation
	members, err := s.queries.GetMembersByOrg(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return members, nil
}

func (s *MemberService) UpdateMemberRole(ctx context.Context, orgID, memberID, actorID string, input UpdateMemberRoleInput) (repository.Member, error) {
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
	member, err := s.queries.GetMemberById(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true})
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

	// fetch user to get their name for activity log
	user, err := s.queries.GetUserById(ctx, member.UserID)
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}
	actor, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// update member role
	updatedMember, err := s.queries.UpdateMember(ctx, repository.UpdateMemberParams{
		ID:   pgtype.UUID{Bytes: parsedMemberID, Valid: true},
		Role: input.Role,
	})
	if err != nil {
		return repository.Member{}, domain.ErrDatabase
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      orgID,
		EntityType: repository.ActivityEntityTypeMEMBER,
		EntityID:   member.ID.String(),
		Action:     repository.ActivityActionROLECHANGED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"member_name":  user.Name,
			"member_email": user.Email,
			"from":         string(member.Role),
			"to":           string(updatedMember.Role),
			"updated_by":   actor.Name,
		},
	})

	return updatedMember, nil
}

func (s *MemberService) DeleteMember(ctx context.Context, orgID, memberID, actorID string) error {
	// parse memberID
	parsedMemberID, err := uuid.Parse(memberID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get member to verify they belong to this org
	member, err := s.queries.GetMemberById(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true})
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
	count, err := s.queries.GetOwnerAndAdminCountByOrg(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return domain.ErrDatabase
	}

	if count <= 1 && (member.Role == repository.MemberRoleOWNER || member.Role == repository.MemberRoleADMIN) {
		return domain.ErrCannotRemoveLastAdmin
	}

	// fetch user to get their name for activity log
	user, err := s.queries.GetUserById(ctx, member.UserID)
	if err != nil {
		return domain.ErrNotFound
	}

	// fetch the actor name for the activity log
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return domain.ErrNotFound
	}
	actor, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: parsedActorID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// delete member
	if err := s.queries.DeleteMember(ctx, pgtype.UUID{Bytes: parsedMemberID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	// log activity
	s.logActivity(LogInput{
		OrgID:      orgID,
		EntityType: repository.ActivityEntityTypeMEMBER,
		EntityID:   member.ID.String(),
		Action:     repository.ActivityActionMEMBERREMOVED,
		ActorID:    actor.ID.String(),
		Metadata: map[string]string{
			"member_name":  user.Name,
			"member_email": user.Email,
			"role":         string(member.Role),
			"deleted_by":   actor.Name,
		},
	})

	return nil
}

func (s *MemberService) LeaveOrg(ctx context.Context, orgID, userID string) error {
	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return domain.ErrNotFound
	}

	// parse userID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	// get member record for this user in this org
	member, err := s.queries.GetMemberByUserAndOrg(ctx, repository.GetMemberByUserAndOrgParams{
		UserID:         pgtype.UUID{Bytes: parsedUserID, Valid: true},
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
	})
	if err != nil {
		return domain.ErrNotOrgMember
	}

	// owner cannot leave — must delete org or transfer ownership first
	if member.Role == repository.MemberRoleOWNER {
		return domain.ErrForbidden
	}

	// enforce at least one owner or admin must remain
	count, err := s.queries.GetOwnerAndAdminCountByOrg(ctx, pgtype.UUID{Bytes: parsedOrgID, Valid: true})
	if err != nil {
		return domain.ErrDatabase
	}

	if count <= 1 && member.Role == repository.MemberRoleADMIN {
		return domain.ErrCannotRemoveLastAdmin
	}

	// fetch user to get their name for activity log
	user, err := s.queries.GetUserById(ctx, member.UserID)
	if err != nil {
		return domain.ErrNotFound
	}

	// remove member from org
	if err := s.queries.DeleteMember(ctx, member.ID); err != nil {
		return domain.ErrDatabase
	}

	// log activity — actor is the user leaving
	s.logActivity(LogInput{
		OrgID:      orgID,
		EntityType: repository.ActivityEntityTypeMEMBER,
		EntityID:   member.ID.String(),
		Action:     repository.ActivityActionMEMBERLEFT,
		ActorID:    userID,
		Metadata: map[string]string{
			"member_name":  user.Name,
			"member_email": user.Email,
			"role":         string(member.Role),
		},
	})

	return nil
}

func (s *MemberService) InviteMember(ctx context.Context, orgID, invitedByUserID string, input InviteMemberInput) (InviteResult, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return InviteResult{}, formatValidationError(err, s.trans)
	}

	// parse orgID
	parsedOrgID, err := uuid.Parse(orgID)
	if err != nil {
		return InviteResult{}, domain.ErrNotFound
	}

	// parse inviter ID
	parsedInviterID, err := uuid.Parse(invitedByUserID)
	if err != nil {
		return InviteResult{}, domain.ErrUnauthorized
	}

	// check if user already exists and is already a member
	invitedUser, err := s.queries.GetUserByEmail(ctx, input.Email)
	if err == nil {
		_, err = s.queries.GetMemberByUserAndOrg(ctx, repository.GetMemberByUserAndOrgParams{
			UserID:         invitedUser.ID,
			OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		})
		if err == nil {
			return InviteResult{}, domain.ErrAlreadyExists
		}
	}

	// check if a pending invitation already exists for this email and org
	_, err = s.queries.GetInvitationByEmailAndOrg(ctx, repository.GetInvitationByEmailAndOrgParams{
		Email:          input.Email,
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
	})
	if err == nil {
		return InviteResult{}, domain.ErrAlreadyExists
	}

	// generate invite token
	token := uuid.New().String()

	// create invitation
	invitation, err := s.queries.CreateInvitation(ctx, repository.CreateInvitationParams{
		ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Email:          input.Email,
		OrganisationID: pgtype.UUID{Bytes: parsedOrgID, Valid: true},
		Role:           input.Role,
		Token:          token,
		InvitedBy:      pgtype.UUID{Bytes: parsedInviterID, Valid: true},
		ExpiresAt:      pgtype.Timestamp{Time: time.Now().Add(invitationDuration), Valid: true},
	})
	if err != nil {
		return InviteResult{}, domain.ErrDatabase
	}

	// return invite link — email service plugged in later
	return InviteResult{
		Invitation: invitation,
		InviteLink: "/api/v1/invitations/accept?token=" + token,
	}, nil
}

func (s *MemberService) AcceptInvitation(ctx context.Context, userID, token string) (repository.Member, error) {
	// validate token is present
	if token == "" {
		return repository.Member{}, domain.ErrInvalidToken
	}

	// parse userID
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return repository.Member{}, domain.ErrUnauthorized
	}

	// get invitation by token — query already checks not accepted and not expired
	invitation, err := s.queries.GetInvitationByToken(ctx, token)
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// get user to verify their email matches the invitation
	user, err := s.queries.GetUserById(ctx, pgtype.UUID{Bytes: parsedUserID, Valid: true})
	if err != nil {
		return repository.Member{}, domain.ErrNotFound
	}

	// verify the accepting user's email matches the invitation email
	if user.Email != invitation.Email {
		return repository.Member{}, domain.ErrForbidden
	}

	var member repository.Member

	// accept invitation and create member in a single transaction
	err = s.db.WithTransaction(ctx, func(q *repository.Queries) error {
		var err error

		// mark invitation as accepted
		_, err = q.AcceptInvitation(ctx, token)
		if err != nil {
			return domain.ErrDatabase
		}

		// add user as member of the organisation
		member, err = q.CreateMember(ctx, repository.CreateMemberParams{
			ID:             pgtype.UUID{Bytes: uuid.New(), Valid: true},
			Role:           invitation.Role,
			UserID:         pgtype.UUID{Bytes: parsedUserID, Valid: true},
			OrganisationID: invitation.OrganisationID,
		})
		if err != nil {
			return domain.ErrDatabase
		}

		return nil
	})
	if err != nil {
		return repository.Member{}, err
	}

	// log activity — actor is the user who accepted the invitation
	s.logActivity(LogInput{
		OrgID:      invitation.OrganisationID.String(),
		EntityType: repository.ActivityEntityTypeMEMBER,
		EntityID:   member.ID.String(),
		Action:     repository.ActivityActionMEMBERJOINED,
		ActorID:    userID,
		Metadata: map[string]string{
			"member_name":  user.Name,
			"member_email": user.Email,
			"role":         string(invitation.Role),
		},
	})

	return member, nil
}
