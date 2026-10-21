package policy_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/policy"
)

func run(t *testing.T, subj *authz.Subject, ps ...policy.Policy) (int, bool) {
	t.Helper()
	called := false
	h := policy.Require(ps...)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { called = true; w.WriteHeader(204) }))
	req := httptest.NewRequestWithContext(context.Background(), "GET", "/", nil)
	if subj != nil {
		req = req.WithContext(authz.WithSubject(req.Context(), *subj))
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code, called
}

func yes() policy.Policy {
	return policy.New("yes", func(policy.Context) (bool, error) { return true, nil })
}
func no() policy.Policy {
	return policy.New("no", func(policy.Context) (bool, error) { return false, nil })
}

func TestRequireECombinadores(t *testing.T) {
	s := &authz.Subject{Type: authz.SubjectEndUser, ID: uuid.New()}
	tests := []struct {
		nome   string
		ps     []policy.Policy
		status int
		called bool
	}{
		{"sem policies", nil, 204, true},
		{"permite", []policy.Policy{yes()}, 204, true},
		{"nega", []policy.Policy{no()}, 403, false},
		{"composição em ordem: todas devem permitir", []policy.Policy{yes(), no()}, 403, false},
		{"all", []policy.Policy{policy.All(yes(), yes())}, 204, true},
		{"all nega", []policy.Policy{policy.All(yes(), no())}, 403, false},
		{"any", []policy.Policy{policy.Any(no(), yes())}, 204, true},
		{"any nega", []policy.Policy{policy.Any(no(), no())}, 403, false},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			code, called := run(t, s, tt.ps...)
			assert.Equal(t, tt.status, code)
			assert.Equal(t, tt.called, called)
		})
	}
	code, _ := run(t, nil, yes())
	assert.Equal(t, 401, code, "sem sujeito")
}

func TestErroDaPolicyEPropagado(t *testing.T) {
	s := &authz.Subject{Type: authz.SubjectEndUser}
	boom := policy.New("boom", func(policy.Context) (bool, error) { return false, httpapi.ErrNotFound() })
	code, called := run(t, s, boom)
	assert.Equal(t, 404, code)
	assert.False(t, called)
	code, _ = run(t, s, policy.All(boom))
	assert.Equal(t, 404, code)
	code, _ = run(t, s, policy.Any(boom, yes()))
	assert.Equal(t, 404, code)
	failing := policy.New("f", func(policy.Context) (bool, error) { return false, errors.New("interno") })
	code, _ = run(t, s, failing)
	assert.Equal(t, 500, code)
}

func TestIsOwner(t *testing.T) {
	owner := uuid.New()
	lookup := func(context.Context, *http.Request) (uuid.UUID, error) { return owner, nil }
	p := policy.IsOwner(lookup)
	assert.Equal(t, "is-owner", p.Name())

	code, called := run(t, &authz.Subject{Type: authz.SubjectEndUser, ID: owner}, p)
	assert.Equal(t, 204, code)
	assert.True(t, called)
	code, _ = run(t, &authz.Subject{Type: authz.SubjectEndUser, ID: uuid.New()}, p)
	assert.Equal(t, 403, code)
	code, _ = run(t, &authz.Subject{Type: authz.SubjectAnonymous}, p)
	assert.Equal(t, 403, code, "anônimo nunca é dono")
	code, _ = run(t, &authz.Subject{Type: authz.SubjectAdmin, ID: owner}, p)
	assert.Equal(t, 403, code, "a policy é de usuário final")

	missing := policy.IsOwner(func(context.Context, *http.Request) (uuid.UUID, error) { return uuid.Nil, httpapi.ErrNotFound() })
	code, _ = run(t, &authz.Subject{Type: authz.SubjectEndUser, ID: owner}, missing)
	assert.Equal(t, 404, code)
	require.Equal(t, "no", no().Name())
}
