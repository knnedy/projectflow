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

type OrgHandler struct {
	org    *service.OrgService
	member *service.MemberService
}

func NewOrgHandler(org *service.OrgService, member *service.MemberService) *OrgHandler {
	return &OrgHandler{org: org, member: member}
}

type OrgResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	OwnerID   string  `json:"owner_id"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt *string `json:"updated_at,omitempty"`
}

type InviteResponse struct {
	InviteLink string `json:"invite_link"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	ExpiresAt  string `json:"expires_at"`
}

func toOrgResponse(org repository.Organisation) OrgResponse {
	var updatedAt *string
	if org.UpdatedAt.Valid {
		t := org.UpdatedAt.Time.Format(time.RFC3339)
		updatedAt = &t
	}

	return OrgResponse{
		ID:        org.ID.String(),
		Name:      org.Name,
		OwnerID:   org.OwnerID.String(),
		CreatedAt: org.CreatedAt.Time.Format(time.RFC3339),
		UpdatedAt: updatedAt,
	}
}

// Create godoc
// @Summary Create a new organisation
// @Tags organisations
// @Accept json
// @Produce json
// @Param body body service.CreateOrgInput true "Create organisation input"
// @Success 201 {object} OrgResponse
// @Failure 401 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /organisations [post]
func (h *OrgHandler) Create(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// decode request body
	var input service.CreateOrgInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// create organisation
	org, err := h.org.Create(r.Context(), userID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, toOrgResponse(org))
}

// List godoc
// @Summary List all organisations for the authenticated user
// @Tags organisations
// @Produce json
// @Success 200 {array} OrgResponse
// @Failure 401 {object} errorResponse
// @Router /organisations [get]
func (h *OrgHandler) List(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// list organisations
	orgs, err := h.org.List(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	data := make([]OrgResponse, len(orgs))
	for i, org := range orgs {
		data[i] = toOrgResponse(org)
	}

	response.WriteJSON(w, http.StatusOK, data)
}

// GetByID godoc
// @Summary Get an organisation by ID
// @Tags organisations
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200 {object} OrgResponse
// @Failure 401 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID} [get]
func (h *OrgHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get organisation
	org, err := h.org.GetByID(r.Context(), orgID)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toOrgResponse(org))
}

// Update godoc
// @Summary Update an organisation
// @Tags organisations
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param body body service.UpdateOrgInput true "Update organisation input"
// @Success 200 {object} OrgResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 422 {object} errorResponse
// @Router /organisations/{orgID} [patch]
func (h *OrgHandler) Update(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// decode request body
	var input service.UpdateOrgInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// get actorID from context
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// update organisation
	org, err := h.org.Update(r.Context(), orgID, actorID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toOrgResponse(org))
}

// Delete godoc
// @Summary Delete an organisation
// @Tags organisations
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID} [delete]
func (h *OrgHandler) Delete(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// delete organisation
	if err := h.org.Delete(r.Context(), orgID); err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, nil)
}

// ListMembers godoc
// @Summary List all members of an organisation
// @Tags organisations
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Success 200 {array} MemberResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Router /organisations/{orgID}/members [get]
func (h *OrgHandler) ListMembers(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// list members
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

// UpdateMember godoc
// @Summary Update a member's role
// @Tags organisations
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
func (h *OrgHandler) UpdateMember(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get actor ID from context
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get member ID from URL
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
	member, err := h.member.UpdateMemberRole(r.Context(), orgID, memberID, actorID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toMemberResponse(member))
}

// DeleteMember godoc
// @Summary Remove a member from an organisation
// @Tags organisations
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param memberID path string true "Member ID"
// @Success 200
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /organisations/{orgID}/members/{memberID} [delete]
func (h *OrgHandler) DeleteMember(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get actor ID from context
	actorID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get member ID from URL
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

// InviteMember godoc
// @Summary Invite a member to an organisation
// @Tags organisations
// @Accept json
// @Produce json
// @Param orgID path string true "Organisation ID"
// @Param body body service.InviteMemberInput true "Invite member input"
// @Success 201 {object} InviteResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Router /organisations/{orgID}/invitations [post]
func (h *OrgHandler) InviteMember(w http.ResponseWriter, r *http.Request) {
	// get org ID from context — resolved by org middleware
	orgID, ok := middleware.GetOrgID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrNotFound)
		return
	}

	// get inviter ID from context
	inviterID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// decode request body
	var input service.InviteMemberInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.WriteError(w, domain.ErrValidation)
		return
	}

	// invite member
	result, err := h.member.InviteMember(r.Context(), orgID, inviterID, input)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, InviteResponse{
		InviteLink: result.InviteLink,
		Email:      result.Invitation.Email,
		Role:       string(result.Invitation.Role),
		ExpiresAt:  result.Invitation.ExpiresAt.Time.Format(time.RFC3339),
	})
}

// AcceptInvitation godoc
// @Summary Accept an organisation invitation
// @Tags organisations
// @Produce json
// @Param token query string true "Invitation token"
// @Success 200 {object} MemberResponse
// @Failure 401 {object} errorResponse
// @Failure 403 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Router /invitations/accept [post]
func (h *OrgHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	// get authenticated user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.WriteError(w, domain.ErrUnauthorized)
		return
	}

	// get token from query param
	token := r.URL.Query().Get("token")
	if token == "" {
		response.WriteError(w, domain.ErrInvalidToken)
		return
	}

	// accept invitation
	member, err := h.member.AcceptInvitation(r.Context(), userID, token)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, toMemberResponse(member))
}
