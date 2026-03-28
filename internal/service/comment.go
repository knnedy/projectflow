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

type CommentService struct {
	db       *repository.Queries
	validate *validator.Validate
	trans    ut.Translator
}

func NewCommentService(db *repository.DB) *CommentService {
	validate, trans := newValidator()
	return &CommentService{
		db:       db.Queries(),
		validate: validate,
		trans:    trans,
	}
}

type CreateCommentInput struct {
	Content string `validate:"required,min=1,max=2000"`
}

type UpdateCommentInput struct {
	Content string `validate:"required,min=1,max=2000"`
}

func (s *CommentService) Create(ctx context.Context, issueID, authorID string, input CreateCommentInput) (repository.Comment, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Comment{}, formatValidationError(err, s.trans)
	}

	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return repository.Comment{}, domain.ErrNotFound
	}

	// verify issue exists
	_, err = s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return repository.Comment{}, domain.ErrNotFound
	}

	// parse author ID
	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		return repository.Comment{}, domain.ErrUnauthorized
	}

	// create comment
	comment, err := s.db.CreateComment(ctx, repository.CreateCommentParams{
		ID:       pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Content:  input.Content,
		AuthorID: pgtype.UUID{Bytes: parsedAuthorID, Valid: true},
		IssueID:  pgtype.UUID{Bytes: parsedIssueID, Valid: true},
	})
	if err != nil {
		return repository.Comment{}, domain.ErrDatabase
	}

	return comment, nil
}

func (s *CommentService) List(ctx context.Context, issueID string) ([]repository.Comment, error) {
	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// verify issue exists
	_, err = s.db.GetIssueById(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return nil, domain.ErrNotFound
	}

	// list comments by issue
	comments, err := s.db.ListCommentsByIssueId(ctx, pgtype.UUID{Bytes: parsedIssueID, Valid: true})
	if err != nil {
		return nil, domain.ErrDatabase
	}

	return comments, nil
}

func (s *CommentService) Update(ctx context.Context, issueID, commentID, authorID string, input UpdateCommentInput) (repository.Comment, error) {
	// validate input
	if err := s.validate.Struct(input); err != nil {
		return repository.Comment{}, formatValidationError(err, s.trans)
	}

	// parse comment ID
	parsedCommentID, err := uuid.Parse(commentID)
	if err != nil {
		return repository.Comment{}, domain.ErrNotFound
	}

	// get comment
	comment, err := s.db.GetCommentById(ctx, pgtype.UUID{Bytes: parsedCommentID, Valid: true})
	if err != nil {
		return repository.Comment{}, domain.ErrNotFound
	}

	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return repository.Comment{}, domain.ErrNotFound
	}

	// verify comment belongs to this issue
	if comment.IssueID.Bytes != parsedIssueID {
		return repository.Comment{}, domain.ErrNotFound
	}

	// parse author ID
	parsedAuthorID, err := uuid.Parse(authorID)
	if err != nil {
		return repository.Comment{}, domain.ErrUnauthorized
	}

	// verify only the author can update their own comment
	if comment.AuthorID.Bytes != parsedAuthorID {
		return repository.Comment{}, domain.ErrForbidden
	}

	// update comment
	updated, err := s.db.UpdateComment(ctx, repository.UpdateCommentParams{
		ID:      pgtype.UUID{Bytes: parsedCommentID, Valid: true},
		Content: input.Content,
	})
	if err != nil {
		return repository.Comment{}, domain.ErrDatabase
	}

	return updated, nil
}

func (s *CommentService) Delete(ctx context.Context, issueID, commentID, actorID string, memberRole repository.MemberRole) error {
	// parse comment ID
	parsedCommentID, err := uuid.Parse(commentID)
	if err != nil {
		return domain.ErrNotFound
	}

	// get comment
	comment, err := s.db.GetCommentById(ctx, pgtype.UUID{Bytes: parsedCommentID, Valid: true})
	if err != nil {
		return domain.ErrNotFound
	}

	// parse issue ID
	parsedIssueID, err := uuid.Parse(issueID)
	if err != nil {
		return domain.ErrNotFound
	}

	// verify comment belongs to this issue
	if comment.IssueID.Bytes != parsedIssueID {
		return domain.ErrNotFound
	}

	// parse actor ID
	parsedActorID, err := uuid.Parse(actorID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	// only the author or an admin/owner can delete a comment
	isAdmin := memberRole == repository.MemberRoleOWNER || memberRole == repository.MemberRoleADMIN
	isAuthor := comment.AuthorID.Bytes == parsedActorID
	if !isAuthor && !isAdmin {
		return domain.ErrForbidden
	}

	// delete comment
	if err := s.db.DeleteComment(ctx, pgtype.UUID{Bytes: parsedCommentID, Valid: true}); err != nil {
		return domain.ErrDatabase
	}

	return nil
}
