package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

type principalKey struct{}

// AdminFrom devolve o administrador autenticado no contexto.
func AdminFrom(ctx context.Context) (*AdminPrincipal, bool) {
	p, ok := ctx.Value(principalKey{}).(*AdminPrincipal)
	return p, ok
}

// RequireAdmin exige um access token do MT Login válido e um administrador ativo.
func (s *AdminAuth) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, found := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !found || token == "" {
			httpapi.Error(w, r, httpapi.ErrUnauthorized())
			return
		}
		p, err := s.Authenticate(r.Context(), token)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	})
}
