package admin

import (
	"context"
	"net/http"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

type decisionKey struct{}

// SubjectOf converte o administrador autenticado no sujeito do motor de permissões.
func SubjectOf(p *auth.AdminPrincipal) authz.Subject {
	return authz.Subject{Type: authz.SubjectAdmin, ID: p.ID, Roles: p.Roles, DepartmentID: p.DepartmentID}
}

// DecisionFrom devolve a decisão de permissão calculada pelo RequirePermission.
// O handler a usa para aplicar o filtro de registros e o de campos.
func DecisionFrom(ctx context.Context) (authz.Decision, bool) {
	d, ok := ctx.Value(decisionKey{}).(authz.Decision)
	return d, ok
}

// RequirePermission exige que o administrador (já autenticado por RequireAdmin)
// tenha a ação. Sem autenticação responde 401; sem permissão, 403 `forbidden`.
// Com permissão condicionada, a decisão segue no contexto (DecisionFrom).
func RequirePermission(engine *authz.Engine, action string, resource ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := auth.AdminFrom(r.Context())
			if !ok {
				httpapi.Error(w, r, httpapi.ErrUnauthorized())
				return
			}
			d, err := engine.Can(r.Context(), SubjectOf(p), action, resource...)
			if err != nil {
				httpapi.Error(w, r, err)
				return
			}
			if !d.Allowed {
				httpapi.Error(w, r, httpapi.ErrForbidden())
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), decisionKey{}, d)))
		})
	}
}

// RequireAnyPermission exige que o administrador (já autenticado por RequireAdmin) tenha ao menos
// uma das ações. Serve a leituras compartilhadas, como a lista de departamentos que alimenta
// o seletor da tela de administradores.
func RequireAnyPermission(engine *authz.Engine, actions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			p, ok := auth.AdminFrom(r.Context())
			if !ok {
				httpapi.Error(w, r, httpapi.ErrUnauthorized())
				return
			}
			for _, action := range actions {
				d, err := engine.Can(r.Context(), SubjectOf(p), action)
				if err != nil {
					httpapi.Error(w, r, err)
					return
				}
				if d.Allowed {
					next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), decisionKey{}, d)))
					return
				}
			}
			httpapi.Error(w, r, httpapi.ErrForbidden())
		})
	}
}
