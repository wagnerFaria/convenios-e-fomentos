package authz

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDoc struct {
	ID        uuid.UUID `json:"id" db:"id" authz:"system"`
	Titulo    string    `json:"titulo" db:"titulo"`
	Conteudo  string    `json:"conteudo"`
	Interna   string    `json:"observacao_interna" db:"observacao_interna"`
	CreatedBy uuid.UUID `json:"created_by_id" db:"created_by_id" authz:"system"`
	Oculto    string    `json:"-"`
	SemTag    string
}

func newDocSchema(t *testing.T) *Schema {
	t.Helper()
	s, err := NewSchema("document", testDoc{})
	require.NoError(t, err)
	return s
}

func TestSchemaAPartirDeTags(t *testing.T) {
	s := newDocSchema(t)
	assert.Equal(t, []string{"id", "titulo", "conteudo", "observacao_interna", "created_by_id"}, s.Fields())
	assert.Equal(t, []string{"titulo", "conteudo", "observacao_interna"}, s.Configurable())
	col, ok := s.Column("conteudo")
	assert.True(t, ok)
	assert.Equal(t, "conteudo", col, "coluna padrão = nome JSON")
	_, ok = s.Column("Oculto")
	assert.False(t, ok)
	assert.True(t, s.IsSystem("id"))
	assert.False(t, s.IsSystem("titulo"))
	assert.Equal(t, []string{"x", "y"}, s.UnknownFields([]string{"y", "titulo", "x"}))

	_, err := NewSchema("p", &testDoc{})
	assert.NoError(t, err, "aceita ponteiro")
	_, err = NewSchema("x", 1)
	assert.Error(t, err)
	_, err = NewSchema("x", nil)
	assert.Error(t, err)
	_, err = NewSchema("x", struct {
		A string `json:"a" db:"a; drop"`
	}{})
	assert.Error(t, err, "coluna fora do padrão de identificador")
}

func TestFiltroDeCamposNaLeitura(t *testing.T) {
	s := newDocSchema(t)
	doc := testDoc{ID: uuid.New(), Titulo: "t", Conteudo: "c", Interna: "segredo", CreatedBy: uuid.New()}

	m, err := s.Project(doc, Fields("titulo"))
	require.NoError(t, err)
	assert.Contains(t, m, "titulo")
	assert.NotContains(t, m, "conteudo")
	assert.NotContains(t, m, "observacao_interna")
	assert.Contains(t, m, "id", "campos de sistema são sempre legíveis")
	assert.Contains(t, m, "created_by_id")

	m, _ = s.Project(doc, AllFields())
	assert.Len(t, m, 5)
	m, _ = s.Project(doc, FieldSet{})
	assert.Len(t, m, 2, "sem campos permitidos sobram só os de sistema")
}

func TestRejeicaoDeCamposNaEscrita(t *testing.T) {
	s := newDocSchema(t)
	allowed := Fields("titulo", "conteudo")
	assert.Empty(t, s.RejectedWrite([]string{"titulo", "conteudo"}, allowed))
	assert.Equal(t, []string{"observacao_interna"}, s.RejectedWrite([]string{"titulo", "observacao_interna"}, allowed))
	// Campos de sistema nunca são graváveis, nem com permissão total.
	assert.Equal(t, []string{"created_by_id", "id"}, s.RejectedWrite([]string{"id", "titulo", "created_by_id", "id"}, AllFields()))
}

func TestFieldSet(t *testing.T) {
	assert.False(t, FieldSet{}.Allows("a"))
	assert.True(t, AllFields().Allows("qualquer"))
	assert.Equal(t, []string{"a", "b"}, Fields("a").Union(Fields("b", "a")).Names())
	assert.True(t, Fields("a").Union(AllFields()).IsAll())
	assert.True(t, fieldSetFromList(nil).IsAll())
	assert.False(t, fieldSetFromList([]string{}).Allows("a"), "lista vazia não permite nenhum campo")
}
