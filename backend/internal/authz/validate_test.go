package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizePermissions(t *testing.T) {
	f := newFixture(t, nil)
	s, err := NewSchema("document", testDoc{})
	require.NoError(t, err)
	f.eng.actions.RegisterSchema("document", s)
	f.eng.actions.MustRegister(Action{ID: "admin::settings.users.read", Domain: DomainAdmin, Category: CategorySettings, Name: "Ler", Resource: "settings.users"})
	got, ok := f.eng.actions.Schema("document")
	require.True(t, ok)
	assert.Same(t, s, got)
	_, ok = f.eng.actions.Schema("nada")
	assert.False(t, ok)

	// Normalização: subject padrão, campos ordenados e sem repetição, condições nunca nulas.
	out, issues := f.eng.NormalizePermissions([]Permission{
		{Action: read, Fields: []string{"titulo", "conteudo", "titulo"}, Conditions: []string{"admin::same-department", "admin::is-creator", "admin::is-creator"}},
		{Action: "admin::settings.users.read"},
	})
	require.Empty(t, issues)
	require.Len(t, out, 2)
	assert.Equal(t, "document", out[0].Subject)
	assert.Equal(t, []string{"conteudo", "titulo"}, out[0].Fields)
	assert.Equal(t, []string{"admin::is-creator", "admin::same-department"}, out[0].Conditions)
	assert.NotNil(t, out[1].Conditions)
	assert.NotNil(t, out[1].Properties)
	assert.Nil(t, out[1].Fields, "nil = todos os campos")

	tests := []struct {
		nome string
		in   Permission
		path string
	}{
		{"ação desconhecida", Permission{Action: "admin::x.y.z"}, "permissions[0].action"},
		{"ação de outro domínio", Permission{Action: "api::x.find"}, "permissions[0].action"},
		{"recurso divergente", Permission{Action: read, Subject: "outro"}, "permissions[0].subject"},
		{"campo inexistente", Permission{Action: read, Fields: []string{"nao_existe"}}, "permissions[0].fields"},
		{"condição desconhecida", Permission{Action: read, Conditions: []string{"admin::nada"}}, "permissions[0].conditions"},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			_, issues := f.eng.NormalizePermissions([]Permission{tt.in})
			require.Len(t, issues, 1)
			assert.Equal(t, tt.path, issues[0].Path)
		})
	}

	// Ação sem suporte a campos/condições.
	f.eng.actions.MustRegister(Action{ID: "admin::content.document.noopts", Domain: DomainAdmin, Category: CategoryCollectionTypes, Name: "x", Resource: "document"})
	_, issues = f.eng.NormalizePermissions([]Permission{{Action: "admin::content.document.noopts", Fields: []string{"titulo"}, Conditions: []string{"admin::is-creator"}}})
	assert.Len(t, issues, 2)

	_, issues = f.eng.NormalizePermissions([]Permission{{Action: read}, {Action: read, Subject: "document"}})
	require.Len(t, issues, 1, "repetida após normalizar o subject")
}

func TestNormalizeMantemListaVaziaDeCampos(t *testing.T) {
	f := newFixture(t, nil)
	out, issues := f.eng.NormalizePermissions([]Permission{{Action: read, Fields: []string{}}})
	require.Empty(t, issues)
	require.NotNil(t, out[0].Fields, "[] (nenhum campo) não pode virar nil (todos os campos)")
	assert.Empty(t, out[0].Fields)
	d := fieldSetFromList(out[0].Fields)
	assert.False(t, d.IsAll())
}
