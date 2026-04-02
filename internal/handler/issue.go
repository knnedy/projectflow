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

type IssueHandler struct {
	issue *service.IssueService
}

func NewIssueHandler(issue *service.IssueService) *IssueHandler {
	return &IssueHandler{issue: issue}
}

type IssueResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Priority    string  `json:"priority"`
	ProjectID   string  `json:"project_id"`
	ReporterID  string  `json:"reporter_id"`
	AssigneeID  string  `json:"assignee_id"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
}

func toIssueResponse(issue repository.Issue) IssueResponse {
	var updatedAt *string
	if issue.UpdatedAt.Valid {
		t := issue.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return IssueResponse{
		ID:          issue.ID.String(),
		Title:       issue.Title,
		Description: issue.Description.String,
		Status:      string(issue.Status),
		Priority:    string(issue.Priority),
		ProjectID:   issue.ProjectID.String(),
		ReporterID:  issue.ReporterID.String(),
		AssigneeID:  issue.AssigneeID.String(),
		CreatedAt:   issue.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:   updatedAt,
	}
}

// Create godoc
// @Summary Create a new issue
// @Tags issues
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param body body service.CreateIssueInput true "Create issue input"
// @Success 201 {object} IssueResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues [post]
func (h *IssueHandler) Create(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get projectID from URL
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.CreateIssueInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// create issue
	issue, err := h.issue.Create(r.Context(), projectID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, toIssueResponse(issue))
}

// GetByID godoc
// @Summary Get an issue by ID
// @Tags issues
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Success 200 {object} IssueResponse
// @Failure 401 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID} [get]
func (h *IssueHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// get projectID from the url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issueID from the url
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issue
	issue, err := h.issue.GetByID(r.Context(), projectID, issueID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toIssueResponse(issue))
}

// List godoc
// @Summary List all issues in a project
// @Tags issues
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Success 200 {array} IssueResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues [get]
func (h *IssueHandler) List(w http.ResponseWriter, r *http.Request) {
	// get projectID from the url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issues
	issues, err := h.issue.List(r.Context(), projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	data := make([]IssueResponse, len(issues))
	for i, issue := range issues {
		data[i] = toIssueResponse(issue)
	}

	response.WriteJSON(w, http.StatusOK, data)
}

// UpdateDetails godoc
// @Summary Update issue details
// @Tags issues
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Param body body service.UpdateIssueDetailsInput true "Update issue details input"
// @Success 200 {object} IssueResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID} [patch]
func (h *IssueHandler) UpdateDetails(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get projectID from url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issueID from url
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.UpdateIssueDetailsInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update issue details
	issue, err := h.issue.UpdateDetails(r.Context(), projectID, issueID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toIssueResponse(issue))
}

// UpdateStatus godoc
// @Summary Update issue status
// @Tags issues
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Param body body service.UpdateIssueStatusInput true "Update issue status input"
// @Success 200 {object} IssueResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 422 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID}/status [patch]
func (h *IssueHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get projectID from url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issueID from url
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get member from context — needed for role based status transition rules
	member, ok := middleware.GetMember(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// decode request body
	var input service.UpdateIssueStatusInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update issue status
	issue, err := h.issue.UpdateStatus(r.Context(), projectID, issueID, userID, input, member.Role)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toIssueResponse(issue))
}

// Delete godoc
// @Summary Delete an issue
// @Tags issues
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param issueID path string true "Issue ID"
// @Success 200
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /organisations/{orgID}/projects/{projectID}/issues/{issueID} [delete]
func (h *IssueHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get projectID from url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get issueID from url
	issueID := chi.URLParam(r, "issueID")
	if issueID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// delete issue
	if err := h.issue.Delete(r.Context(), projectID, issueID, userID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}
