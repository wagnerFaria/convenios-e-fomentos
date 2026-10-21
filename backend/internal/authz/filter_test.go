package authz

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func cols(allowed ...string) func(string) (string, bool) {
	return func(f string) (string, bool) {
		for _, a := range allowed {
			if a == f {
				return "t_" + f, true
			}
		}
		return "", false
	}
}

func TestFilterToSQL(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		nome string
		f    Filter
		sql  string
		args []any
	}{
		{"verdadeiro", FilterTrue(), "TRUE", nil},
		{"falso", FilterFalse(), "FALSE", nil},
		{"eq", Eq("a", 1), "t_a = $1", []any{1}},
		{"neq", Neq("a", "x"), "t_a <> $1", []any{"x"}},
		{"in", In("a", 1, 2, 3), "t_a IN ($1, $2, $3)", []any{1, 2, 3}},
		{"in vazio é falso", In("a"), "FALSE", nil},
		{"eq nulo é falso", Eq("a", nil), "FALSE", nil},
		{"and", And(Eq("a", 1), Neq("b", 2)), "(t_a = $1 AND t_b <> $2)", []any{1, 2}},
		{"or", Or(Eq("a", 1), Eq("b", id)), "(t_a = $1 OR t_b = $2)", []any{1, id}},
		{"aninhado", And(Or(Eq("a", 1), Eq("b", 2)), Eq("a", 3)), "((t_a = $1 OR t_b = $2) AND t_a = $3)", []any{1, 2, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			sql, args, err := tt.f.ToSQL(cols("a", "b"), 1)
			require.NoError(t, err)
			assert.Equal(t, tt.sql, sql)
			assert.Equal(t, tt.args, args)
		})
	}
}

func TestFilterToSQLNumeracaoInicial(t *testing.T) {
	sql, args, err := And(Eq("a", 1), Eq("b", 2)).ToSQL(cols("a", "b"), 4)
	require.NoError(t, err)
	assert.Equal(t, "(t_a = $4 AND t_b = $5)", sql)
	assert.Len(t, args, 2)
}

func TestFilterNaoPermiteCampoForaDaLista(t *testing.T) {
	_, _, err := Eq("senha; DROP TABLE x", 1).ToSQL(cols("a"), 1)
	assert.Error(t, err)
	_, _, err = And(Eq("a", 1), Eq("oculto", 2)).ToSQL(cols("a"), 1)
	assert.Error(t, err)
}

func TestFilterValoresNuncaVaoParaOSQL(t *testing.T) {
	evil := "x' OR '1'='1"
	sql, args, err := Eq("a", evil).ToSQL(cols("a"), 1)
	require.NoError(t, err)
	assert.NotContains(t, sql, evil)
	assert.Equal(t, []any{evil}, args)
}

func TestFilterSimplificacao(t *testing.T) {
	assert.True(t, And().IsTrue())
	assert.True(t, Or().IsFalse())
	assert.True(t, And(FilterTrue(), FilterTrue()).IsTrue())
	assert.True(t, And(Eq("a", 1), FilterFalse()).IsFalse())
	assert.True(t, Or(Eq("a", 1), FilterTrue()).IsTrue())
	assert.Equal(t, Eq("a", 1), And(FilterTrue(), Eq("a", 1)))
	assert.Equal(t, Eq("a", 1), Or(FilterFalse(), Eq("a", 1)))
	// Achatamento de and/or aninhados.
	assert.Len(t, And(And(Eq("a", 1), Eq("b", 2)), Eq("c", 3)).Children, 3)
	assert.Len(t, Or(Or(Eq("a", 1), Eq("b", 2)), Eq("c", 3)).Children, 3)
}

func TestFilterMatch(t *testing.T) {
	id := uuid.New()
	idp := &id
	rec := record(map[string]any{"a": 1, "b": "x", "u": id, "p": idp, "nulo": nil, "nilptr": (*uuid.UUID)(nil)})
	tests := []struct {
		nome string
		f    Filter
		want bool
	}{
		{"eq", Eq("a", 1), true},
		{"eq falso", Eq("a", 2), false},
		{"neq", Neq("a", 2), true},
		{"in", In("b", "y", "x"), true},
		{"in sem", In("b", "y"), false},
		{"uuid x ponteiro", Eq("u", idp), true},
		{"ponteiro x uuid", Eq("p", id), true},
		{"campo ausente", Eq("zz", 1), false},
		{"neq em ausente", Neq("zz", 1), false},
		{"campo nulo", Eq("nulo", 1), false},
		{"neq em nulo", Neq("nulo", 1), false},
		{"ponteiro nulo", Eq("nilptr", id), false},
		{"and", And(Eq("a", 1), Eq("b", "x")), true},
		{"and falha", And(Eq("a", 1), Eq("b", "y")), false},
		{"or", Or(Eq("a", 9), Eq("b", "x")), true},
		{"or falha", Or(Eq("a", 9), Eq("b", "y")), false},
		{"true", FilterTrue(), true},
		{"false", FilterFalse(), false},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) { assert.Equal(t, tt.want, tt.f.Match(rec)) })
	}
}

func TestFilterOperadorDesconhecido(t *testing.T) {
	bad := Filter{kind: kindLeaf, Field: "a", Op: "like", Values: []any{1}}
	_, _, err := bad.ToSQL(cols("a"), 1)
	assert.Error(t, err)
	assert.False(t, bad.Match(record(map[string]any{"a": 1})))
}
