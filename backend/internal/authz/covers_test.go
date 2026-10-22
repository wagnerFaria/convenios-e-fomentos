package authz

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWidened(t *testing.T) {
	f := newFixture(t, nil)
	doc := func(fields, conds []string) Permission {
		return Permission{Action: read, Subject: "document", Fields: fields, Conditions: conds}
	}
	cases := []struct {
		nome          string
		before, after []Permission
		widened       int
	}{
		{"permissão nova", nil, []Permission{doc(nil, nil)}, 1},
		{"igual", []Permission{doc([]string{"a"}, []string{"c"})}, []Permission{doc([]string{"a"}, []string{"c"})}, 0},
		{"menos campos", []Permission{doc([]string{"a", "b"}, nil)}, []Permission{doc([]string{"a"}, nil)}, 0},
		{"mais campos", []Permission{doc([]string{"a"}, nil)}, []Permission{doc([]string{"a", "b"}, nil)}, 1},
		{"todos os campos amplia lista", []Permission{doc([]string{"a"}, nil)}, []Permission{doc(nil, nil)}, 1},
		{"lista vazia (nenhum campo) cabe em tudo", []Permission{doc([]string{"a"}, nil)}, []Permission{doc([]string{}, nil)}, 0},
		{"nova condição restringe", []Permission{doc(nil, nil)}, []Permission{doc(nil, []string{"admin::is-creator"})}, 0},
		{"remover condição amplia", []Permission{doc(nil, []string{"admin::is-creator"})}, []Permission{doc(nil, nil)}, 1},
		{"mais condições (OR) amplia", []Permission{doc(nil, []string{"admin::is-creator"})}, []Permission{doc(nil, []string{"admin::is-creator", "admin::same-department"})}, 1},
		{"recurso vazio equivale ao da ação", []Permission{{Action: read}}, []Permission{doc(nil, nil)}, 0},
		{"remoção nunca amplia", []Permission{doc(nil, nil)}, nil, 0},
	}
	for _, c := range cases {
		t.Run(c.nome, func(t *testing.T) {
			assert.Len(t, f.eng.Widened(c.before, c.after), c.widened)
		})
	}
}

func TestCovers(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	r := role("gestor")
	f.store.set(r.ID,
		Permission{Action: read, Subject: "document", Fields: []string{"titulo", "conteudo"}, Conditions: []string{"admin::same-department"}},
		Permission{Action: update, Subject: "document"})
	actor := adminWith(r)

	missing := func(p ...Permission) []Permission {
		t.Helper()
		m, err := f.eng.Covers(ctx, actor, p)
		require.NoError(t, err)
		return m
	}
	assert.Empty(t, missing(Permission{Action: update, Subject: "document", Fields: []string{"x"}, Conditions: []string{"admin::is-creator"}}), "o ator pode tudo em update, logo pode restringir")
	assert.Empty(t, missing(Permission{Action: read, Subject: "document", Fields: []string{"titulo"}, Conditions: []string{"admin::same-department"}}))
	assert.Len(t, missing(Permission{Action: read, Subject: "document", Fields: []string{"titulo"}}), 1, "sem condição é mais amplo que same-department")
	assert.Len(t, missing(Permission{Action: read, Subject: "document", Fields: []string{"segredo"}, Conditions: []string{"admin::same-department"}}), 1, "campo que o ator não lê")
	assert.Len(t, missing(Permission{Action: read, Subject: "document", Conditions: []string{"admin::same-department"}}), 1, "todos os campos")
	assert.Len(t, missing(Permission{Action: del, Subject: "document"}), 1, "ação que o ator não tem")

	su := adminWith(role(SuperAdminRoleCode))
	m, err := f.eng.Covers(ctx, su, []Permission{{Action: del, Subject: "document"}})
	require.NoError(t, err)
	assert.Empty(t, m, "Super Admin cobre tudo")
}

func TestCoversRolesECoversSubject(t *testing.T) {
	ctx := context.Background()
	f := newFixture(t, nil)
	small, big := role("pequena"), role("grande")
	f.store.set(small.ID, Permission{Action: read, Subject: "document"})
	f.store.set(big.ID, Permission{Action: read, Subject: "document"}, Permission{Action: del, Subject: "document"})

	m, err := f.eng.CoversRoles(ctx, adminWith(small), []uuid.UUID{small.ID, big.ID})
	require.NoError(t, err)
	require.Len(t, m, 1)
	assert.Equal(t, del, m[0].Action)

	m, err = f.eng.CoversSubject(ctx, adminWith(small), adminWith(big))
	require.NoError(t, err)
	assert.Len(t, m, 1, "quem tem menos poder não age sobre quem tem mais")
	m, err = f.eng.CoversSubject(ctx, adminWith(big), adminWith(small))
	require.NoError(t, err)
	assert.Empty(t, m)
	m, err = f.eng.CoversSubject(ctx, adminWith(small), adminWith())
	require.NoError(t, err)
	assert.Empty(t, m, "admin sem perfil não tem poder algum")
	m, err = f.eng.CoversSubject(ctx, adminWith(small), adminWith(role(SuperAdminRoleCode)))
	require.NoError(t, err)
	assert.NotEmpty(t, m, "ninguém além do Super Admin age sobre um Super Admin")
}
