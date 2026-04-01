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

type MemberHandler struct {
	member *service.MemberService
}

func NewMemberHandler(member *service.MemberService) *MemberHandler {
	return &MemberHandler{member: member}
}

type MemberResponse struct {
	ID             string  `json:"id"`
	Role           string  `json:"role"`
	UserID         string  `json:"user_id"`
	OrganisationID string  `json:"organisation_id"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
}

func toMemberResponse(member repository.Member) MemberResponse {
	var updatedAt *string
	if member.UpdatedAt.Valid {
		t := member.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return MemberResponse{
		ID:             member.ID.String(),
		Role:           string(member.Role),
		UserID:         member.UserID.String(),
		OrganisationID: member.OrganisationID.String(),
		CreatedAt:      member.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt:      updatedAt,
	}
}

// List godoc
// @Summary List all members of an organisation
// @Tags members
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200 {array} MemberResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Router /organisations/{orgID}/members [get]
func (h *MemberHandler) List(w http.ResponseWriter, r *http.Request) {
	// get organisationID from context
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get members
	members, err := h.member.ListMembers(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	data := make([]MemberResponse, len(members))
	for i, member := range members {
		data[i] = toMemberResponse(member)
	}

	response.WriteJSON(w, http.StatusOK, data)
}

// UpdateMemberRole godoc
// @Summary Update a member's role
// @Tags members
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param memberID path string true "Member ID"
// @Param body body service.UpdateMemberRoleInput true "Update member role input"
// @Success 200 {object} MemberResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/members/{memberID} [patch]
func (h *MemberHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	actorID, ok := middleware.GetUserID(r.Context())
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

	// get target memberID from URL
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.UpdateMemberRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// update member role
	updatedMember, err := h.member.UpdateMemberRole(r.Context(), orgID, memberID, actorID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toMemberResponse(updatedMember))
}

// Delete godoc
// @Summary Remove a member from an organisation
// @Tags members
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param memberID path string true "Member ID"
// @Success 200
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/members/{memberID} [delete]
func (h *MemberHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// get authenticated userID from context
	actorID, ok := middleware.GetUserID(r.Context())
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

	// get target memberID from URL
	memberID := chi.URLParam(r, "memberID")
	if memberID == "" {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// delete member
	if err := h.member.DeleteMember(r.Context(), orgID, memberID, actorID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}

// LeaveOrg godoc
// @Summary Leave an organisation
// @Tags members
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Router /organisations/{orgID}/members/me [delete]
func (h *MemberHandler) LeaveOrg(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// leave organisation
	if err := h.member.LeaveOrg(r.Context(), orgID, userID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}
