package authorization

import (
	"net/http"

	"github.com/observatorium/api/authentication"
	"github.com/observatorium/api/rbac"
)

// WithAuthorizers returns a middleware that authorizes subjects taken from a request context
// for the given permission on the given resource for a tenant taken from a request context.
func WithAuthorizers(authorizers map[string]rbac.Authorizer, permission rbac.Permission, resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, ok := authentication.GetTenant(r.Context())
			if !ok {
				http.Error(w, "error finding tenant", http.StatusInternalServerError)

				return
			}
			subject, ok := authentication.GetSubject(r.Context())
			if !ok {
				http.Error(w, "unknown subject", http.StatusUnauthorized)

				return
			}
			groups, ok := authentication.GetGroups(r.Context())
			if !ok {
				groups = []string{}
			}
			a, ok := authorizers[tenant]
			if !ok {
				http.Error(w, "error finding tenant", http.StatusUnauthorized)

				return
			}

			token, ok := authentication.GetAccessToken(r.Context())
			if !ok {
				http.Error(w, "error finding access token", http.StatusUnauthorized)

				return
			}

			tenantID, _ := authentication.GetTenantID(r.Context())

			if statusCode, ok := a.Authorize(subject, groups, permission, resource, tenant, tenantID, token); !ok {
				w.WriteHeader(statusCode)

				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
