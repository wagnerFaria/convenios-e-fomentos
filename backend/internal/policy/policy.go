// Pacote policy define policies: regras em código executadas antes do handler,
// depois da verificação de permissão, que recebem a requisição e o sujeito
// autenticado e decidem permitir ou negar. São compostas na definição da rota:
//
//	r.With(identify, enduser.RequireAction(engine, "api::document.update"),
//	    policy.Require(policy.IsOwner(lookup))).Put("/api/documents/{id}", h.update)
package policy

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Context é o que uma policy recebe.
type Context struct {
	Request *http.Request
	Subject authz.Subject
}

// Policy decide se a requisição segue para o handler. Um erro retornado é
// respondido como está (use *httpapi.APIError para 404 etc.).
type Policy interface {
	Name() string
	Allow(c Context) (bool, error)
}

type fn struct {
	name string
	f    func(Context) (bool, error)
}

func (p fn) Name() string                  { return p.name }
func (p fn) Allow(c Context) (bool, error) { return p.f(c) }

// New cria uma policy a partir de uma função.
func New(name string, f func(Context) (bool, error)) Policy { return fn{name: name, f: f} }

// All exige que todas as policies permitam (E).
func All(ps ...Policy) Policy {
	return New("all", func(c Context) (bool, error) {
		for _, p := range ps {
			ok, err := p.Allow(c)
			if err != nil || !ok {
				return false, err
			}
		}
		return true, nil
	})
}

// Any exige que ao menos uma policy permita (OU).
func Any(ps ...Policy) Policy {
	return New("any", func(c Context) (bool, error) {
		for _, p := range ps {
			ok, err := p.Allow(c)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	})
}

// Require executa as policies, em ordem, antes do handler. Falha em qualquer
// uma → 403 `forbidden` (com o nome da policy em details). A requisição precisa
// ter passado por um middleware que guarda o sujeito (authz.WithSubject); sem
// sujeito responde 401.
func Require(ps ...Policy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			subj, ok := authz.SubjectFrom(r.Context())
			if !ok {
				httpapi.Error(w, r, httpapi.ErrUnauthorized())
				return
			}
			c := Context{Request: r, Subject: subj}
			for _, p := range ps {
				allowed, err := p.Allow(c)
				if err != nil {
					httpapi.Error(w, r, err)
					return
				}
				if !allowed {
					e := httpapi.ErrForbidden()
					e.Details = map[string]string{"policy": p.Name()}
					httpapi.Error(w, r, e)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// OwnerLookup devolve o ID do dono do registro referenciado pela requisição
// (normalmente lendo o parâmetro de rota). Registro inexistente → *httpapi.APIError 404.
type OwnerLookup func(ctx context.Context, r *http.Request) (uuid.UUID, error)

// IsOwner permite apenas ao usuário final dono do registro.
func IsOwner(lookup OwnerLookup) Policy {
	return New("is-owner", func(c Context) (bool, error) {
		if c.Subject.Type != authz.SubjectEndUser {
			return false, nil
		}
		owner, err := lookup(c.Request.Context(), c.Request)
		if err != nil {
			return false, err
		}
		return owner == c.Subject.ID, nil
	})
}
