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

type ProjectHandler struct {
	project *service.ProjectService
}

func NewProjectHandler(project *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{project: project}
}

type ProjectResponse struct {
	ID             string  `json:"id"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	OrganisationID string  `json:"organisation_id"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
}

func toProjectResponse(project repository.Project) ProjectResponse {
	var updatedAt *string
	if project.UpdatedAt.Valid {
		t := project.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return ProjectResponse{
		ID:             project.ID.String(),
		Name:           project.Name,
		Description:    project.Description.String,
		OrganisationID: project.OrganisationID.String(),
		CreatedAt:      project.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      updatedAt,
	}
}

// Create godoc
// @Summary Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param body body service.CreateProjectInput true "Create project input"
// @Success 201 {object} ProjectResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /organisations/{orgID}/projects [post]
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.CreateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// create project
	project, err := h.project.Create(r.Context(), orgID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, toProjectResponse(project))
}

// GetByID godoc
// @Summary Get a project by ID
// @Tags projects
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Success 200 {object} ProjectResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/projects/{projectID} [get]
func (h *ProjectHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get projectID from the url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get project
	project, err := h.project.GetByID(r.Context(), orgID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toProjectResponse(project))
}

// List godoc
// @Summary List all projects in an organisation
// @Tags projects
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200 {array} ProjectResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Router /organisations/{orgID}/projects [get]
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get projects
	projects, err := h.project.List(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	data := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		data[i] = toProjectResponse(project)
	}

	response.WriteJSON(w, http.StatusOK, data)
}

// Update godoc
// @Summary Update a project
// @Tags projects
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param body body service.UpdateProjectInput true "Update project input"
// @Success 200 {object} ProjectResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /organisations/{orgID}/projects/{projectID} [patch]
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get projectID from url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.UpdateProjectInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update project
	project, err := h.project.Update(r.Context(), orgID, projectID, userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toProjectResponse(project))
}

// Delete godoc
// @Summary Delete a project
// @Tags projects
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Success 200
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/projects/{projectID} [delete]
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get projectID from url
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// delete project
	if err := h.project.Delete(r.Context(), orgID, projectID, userID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}
