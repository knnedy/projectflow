package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/middleware"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/response"
	"github.com/knnedy/projectflow/internal/service"
)

type CommentHandler struct {
	comment *service.CommentService
}

func NewCommentHandler(comment *service.CommentService) *CommentHandler {
	return &CommentHandler{comment: comment}
}

type CommentResponse struct {
	ID        string  `json:"id"`
	Content   string  `json:"content"`
	AuthorID  string  `json:"author_id"`
	IssueID   string  `json:"issue_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

func toCommentResponse(comment repository.Comment) CommentResponse {
	var updatedAt *string
	if comment.UpdatedAt.Valid {
		t := comment.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return CommentResponse{
		ID:        comment.ID.String(),
		Content:   comment.Content,
		AuthorID:  comment.AuthorID.String(),
		IssueID:   comment.IssueID.String(),
		CreatedAt: comment.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: updatedAt,
	}
}

// Create godoc
// @Summary Create a comment on an issue
// @Tags comments
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Param body body service.CreateCommentInput true "Create comment input"
// @Success 201 {object} CommentResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments [post]
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get issueID from URL
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.CreateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// create comment
	comment, err := h.comment.Create(r.Context(), issueID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, toCommentResponse(comment))
}

// List godoc
// @Summary List all comments on an issue
// @Tags comments
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Success 200 {array} CommentResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments [get]
func (h *CommentHandler) List(w http.ResponseWriter, r *http.Request) {
	// get issueID from URL
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// list comments
	comments, err := h.comment.List(r.Context(), issueID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	data := make([]CommentResponse, len(comments))
	for i, comment := range comments {
		data[i] = toCommentResponse(comment)
	}

	response.WriteJSON(w, http.StatusOK, data)
}

// Update godoc
// @Summary Update a comment
// @Tags comments
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Param commentID path string true "Comment ID"
// @Param body body service.UpdateCommentInput true "Update comment input"
// @Success 200 {object} CommentResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments/{commentID} [patch]
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get issueID from URL
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get commentID from URL
	commentID := chi.URLParam(r, "commentID")
	if commentID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.UpdateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update comment
	comment, err := h.comment.Update(r.Context(), issueID, commentID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toCommentResponse(comment))
}

// Delete godoc
// @Summary Delete a comment
// @Tags comments
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Param commentID path string true "Comment ID"
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID}/comments/{commentID} [delete]
func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get issueID from URL
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get commentID from URL
	commentID := chi.URLParam(r, "commentID")
	if commentID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get member from context — needed for role based delete permission
	member, ok := middleware.GetMember(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// delete comment
	if err := h.comment.Delete(r.Context(), issueID, commentID, userID, member.Role); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}
