package middleware

import (
	"net/http"
	"slices"

	"github.com/knnedy/projectflow/internal/domain"
	"github.com/knnedy/projectflow/internal/handler"
	"github.com/knnedy/projectflow/internal/repository"
)

// RequireRole returns a middleware that enforces the authenticated member
// has one of the allowed roles. Must run after Authenticate and ResolveOrg.
func RequireRole(allowedRoles ...repository.MemberRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// get member from context — ResolveOrg middleware must run first
			member, ok := GetMember(r.Context())
			if !ok {
				handler.WriteError(w, domain.ErrUnauthorized)
				return
			}

			// check if member's role is in allowed roles
			if !hasRole(member.Role, allowedRoles) {
				handler.WriteError(w, domain.ErrForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRole checks if the member's role is in the allowed roles
func hasRole(memberRole repository.MemberRole, allowedRoles []repository.MemberRole) bool {
	return slices.Contains(allowedRoles, memberRole)
}
