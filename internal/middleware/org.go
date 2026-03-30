package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/response"
)

const (
	contextKeyOrgID    = "organisation_id"
	contextKeyMemberID = "member_id"
)

type OrgMiddleware struct {
	db *repository.Queries
}

func NewOrgMiddleware(db *repository.Queries) *OrgMiddleware {
	return &OrgMiddleware{db: db}
}

// ResolveOrg extracts the orgID from the URL, verifies it exists,
// verifies the authenticated user is a member of the org and attaches
// the org and member to the context
func (om *OrgMiddleware) ResolveOrg(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract orgID from the URL param
		orgIDStr := chi.URLParam(r, "orgID")
		if orgIDStr == "" {
			response.WriteError(w, domain.ErrNotFound)
			return
		}

		// Parse orgID into uuid
		parsedOrgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			response.WriteError(w, domain.ErrNotFound)
			return
		}

		orgID := pgtype.UUID{Bytes: parsedOrgID, Valid: true}

		// Verify org exists
		org, err := om.db.GetOrganisationById(r.Context(), orgID)
		if err != nil {
			response.WriteError(w, domain.ErrNotFound)
			return
		}

		// Get authenticated user ID from context - auth middleware should have already verified must run first
		userIDStr, ok := GetUserID(r.Context())
		if !ok {
			response.WriteError(w, domain.ErrUnauthorized)
			return
		}

		// Parse userID into uuid
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			response.WriteError(w, domain.ErrUnauthorized)
			return
		}

		// Verify user is a member of the org
		member, err := om.db.GetMemberByUserAndOrg(r.Context(), repository.GetMemberByUserAndOrgParams{
			UserID:         pgtype.UUID{Bytes: parsedUserID, Valid: true},
			OrganisationID: orgID,
		})
		if err != nil {
			response.WriteError(w, domain.ErrNotOrgMember)
			return
		}

		// Attach org and member to context
		ctx := context.WithValue(r.Context(), contextKeyOrgID, org.ID.String())
		ctx = context.WithValue(ctx, contextKeyMemberID, member)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetOrgID retrieves the orgID from the context
func GetOrgID(ctx context.Context) (string, bool) {
	orgID, ok := ctx.Value(contextKeyOrgID).(string)
	return orgID, ok
}

// GetMember retrieves the member from the context
func GetMember(ctx context.Context) (repository.Member, bool) {
	member, ok := ctx.Value(contextKeyMemberID).(repository.Member)
	return member, ok
}
