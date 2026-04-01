package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/middleware"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/response"
	"github.com/knnedy/projectflow/internal/service"
)

type ActivityHandler struct {
	activity *service.ActivityService
}

func NewActivityHandler(activity *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{activity: activity}
}

type ActivityLogResponse struct {
	ID             string          `json:"id"`
	OrganisationID string          `json:"organisation_id"`
	ProjectID      *string         `json:"project_id,omitempty"`
	EntityType     string          `json:"entity_type"`
	EntityID       string          `json:"entity_id"`
	Action         string          `json:"action"`
	ActorID        string          `json:"actor_id"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	CreatedAt      string          `json:"created_at"`
}

type ActivityPageResponse struct {
	Logs       []ActivityLogResponse `json:"logs"`
	NextCursor *string               `json:"next_cursor,omitempty"`
	HasMore    bool                  `json:"has_more"`
}

func toActivityLogResponse(log repository.ActivityLog) ActivityLogResponse {
	r := ActivityLogResponse{
		ID:             log.ID.String(),
		OrganisationID: log.OrganisationID.String(),
		EntityType:     string(log.EntityType),
		EntityID:       log.EntityID.String(),
		Action:         string(log.Action),
		ActorID:        log.ActorID.String(),
		CreatedAt:      log.CreatedAt.Time.Format(time.RFC3339),
	}

	if log.ProjectID.Valid {
		id := log.ProjectID.String()
		r.ProjectID = &id
	}

	if log.Metadata != nil {
		r.Metadata = json.RawMessage(log.Metadata)
	}

	return r
}

func toActivityPageResponse(page service.ActivityPage) ActivityPageResponse {
	logs := make([]ActivityLogResponse, len(page.Logs))
	for i, log := range page.Logs {
		logs[i] = toActivityLogResponse(log)
	}

	var nextCursor *string
	if page.NextCursor != nil {
		c := page.NextCursor.Format(time.RFC3339Nano)
		nextCursor = &c
	}

	return ActivityPageResponse{
		Logs:       logs,
		NextCursor: nextCursor,
		HasMore:    page.HasMore,
	}
}

// parsePaginationParams extracts cursor and limit from query params
func parsePaginationParams(r *http.Request) (*time.Time, int32) {
	var cursor *time.Time
	if c := r.URL.Query().Get("cursor"); c != "" {
		if t, err := time.Parse(time.RFC3339Nano, c); err == nil {
			cursor = &t
		}
	}

	limit := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.ParseInt(l, 10, 32); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	return cursor, limit
}

// ListByOrg godoc
// @Summary List activity logs for an organisation
// @Tags activity
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of logs to return"
// @Success 200 {object} ActivityPageResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/activity [get]
func (h *ActivityHandler) ListByOrg(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// parse pagination params
	cursor, limit := parsePaginationParams(r)

	// list activity logs
	page, err := h.activity.ListByOrg(r.Context(), orgID, cursor, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toActivityPageResponse(page))
}

// ListByProject godoc
// @Summary List activity logs for a project
// @Tags activity
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param projectID path string true "Project ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of logs to return"
// @Success 200 {object} ActivityPageResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/projects/{projectID}/activity [get]
func (h *ActivityHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	// get project ID from URL
	projectID := chi.URLParam(r, "projectID")
	if projectID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// parse pagination params
	cursor, limit := parsePaginationParams(r)

	// list activity logs
	page, err := h.activity.ListByProject(r.Context(), projectID, cursor, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toActivityPageResponse(page))
}

// ListByEntity godoc
// @Summary List activity logs for a specific entity
// @Tags activity
// @Produce json
// @Param entityID path string true "Entity ID"
// @Param cursor query string false "Cursor for pagination"
// @Param limit query int false "Number of logs to return"
// @Success 200 {object} ActivityPageResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /activity/{entityID} [get]
func (h *ActivityHandler) ListByEntity(w http.ResponseWriter, r *http.Request) {
	// get entity ID from URL
	entityID := chi.URLParam(r, "entityID")
	if entityID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// parse pagination params
	cursor, limit := parsePaginationParams(r)

	// list activity logs
	page, err := h.activity.ListByEntity(r.Context(), entityID, cursor, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toActivityPageResponse(page))
}
