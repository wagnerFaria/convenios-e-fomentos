package enduser

import (
	"context"
	"net/http"
	"strings"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// APITokenPrefix identifica tokens de API (distintos de JWTs) no cabeçalho Authorization.
const APITokenPrefix = "gat_"

// TokenAuthenticator autentica tokens de API (implementado na Fase 6).
type TokenAuthenticator interface {
	Authenticate(ctx context.Context, raw string) (authz.Subject, error)
}

type principalKey struct{}

// PrincipalFrom devolve o usuário final autenticado (quando o sujeito é um).
func PrincipalFrom(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey{}).(*Principal)
	return p, ok
}

// Identifier identifica quem faz a requisição na API pública: usuário final
// (JWT, audience api), token de API (prefixo gat_) ou anônimo (role Public).
type Identifier struct {
	Svc    *Service
	Engine *authz.Engine
	Tokens TokenAuthenticator // opcional
}

// Identify guarda o sujeito no contexto. Sem Authorization → anônimo (Public).
// Authorization presente porém inválido → 401 (não cai silenciosamente em anônimo).
func (i *Identifier) Identify(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		ctx := r.Context()
		if header == "" {
			publicID, err := i.Svc.PublicRoleID(ctx)
			if err != nil {
				httpapi.Error(w, r, err)
				return
			}
			subj := authz.Subject{Type: authz.SubjectAnonymous, Roles: []authz.Role{{ID: publicID, Code: RolePublic}}}
			next.ServeHTTP(w, r.WithContext(authz.WithSubject(ctx, subj)))
			return
		}
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			httpapi.Error(w, r, httpapi.ErrUnauthorized())
			return
		}
		if strings.HasPrefix(token, APITokenPrefix) {
			if i.Tokens == nil {
				httpapi.Error(w, r, httpapi.ErrUnauthorized())
				return
			}
			subj, err := i.Tokens.Authenticate(ctx, token)
			if err != nil {
				httpapi.Error(w, r, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(authz.WithSubject(ctx, subj)))
			return
		}
		p, err := i.Svc.Authenticate(ctx, token)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		ctx = context.WithValue(authz.WithSubject(ctx, p.Subject()), principalKey{}, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAction exige a permissão da ação para o sujeito identificado. Sem
// permissão: anônimo → 401 `unauthorized`; autenticado → 403 `forbidden`.
func (i *Identifier) RequireAction(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subj, ok := authz.SubjectFrom(r.Context())
			if !ok {
				httpapi.Error(w, r, httpapi.ErrUnauthorized())
				return
			}
			d, err := i.Engine.Can(r.Context(), subj, action)
			if err != nil {
				httpapi.Error(w, r, err)
				return
			}
			if !d.Allowed {
				if subj.Type == authz.SubjectAnonymous {
					httpapi.Error(w, r, httpapi.ErrUnauthorized())
				} else {
					httpapi.Error(w, r, httpapi.ErrForbidden())
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireUser exige um usuário final autenticado (JWT).
func RequireUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := PrincipalFrom(r.Context()); !ok {
			httpapi.Error(w, r, httpapi.ErrUnauthorized())
			return
		}
		next.ServeHTTP(w, r)
	})
}
