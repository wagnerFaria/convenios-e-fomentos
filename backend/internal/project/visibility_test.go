package project

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
)

func visibilityResult(t *testing.T, id string, s authz.Subject) authz.Result {
	t.Helper()
	reg := authz.NewConditionRegistry()
	registerVisibilityConditions(reg)
	c, ok := reg.Get(id)
	require.True(t, ok, "condição %s registrada", id)
	r, err := c.Handler(context.Background(), s)
	require.NoError(t, err)
	return r
}

func TestCondicaoProjetoEnviado(t *testing.T) {
	f := visibilityResult(t, ConditionSubmitted, authz.Subject{}).Filter()
	for status, want := range map[string]bool{StatusDraft: false, StatusSubmitted: true, StatusArchived: true, "": false} {
		assert.Equal(t, want, f.Match(Ref{Status: status}.Getter()), "status %q", status)
	}
}

func TestCondicaoProjetoEnviadoDoMesmoDepartamento(t *testing.T) {
	mine, other := uuid.New(), uuid.New()
	f := visibilityResult(t, ConditionSubmittedSameDepartment, authz.Subject{DepartmentID: &mine}).Filter()

	cases := []struct {
		name   string
		status string
		dept   *uuid.UUID
		want   bool
	}{
		{"enviado do meu departamento", StatusSubmitted, &mine, true},
		{"arquivado do meu departamento", StatusArchived, &mine, true},
		{"rascunho do meu departamento", StatusDraft, &mine, false},
		{"enviado de outro departamento", StatusSubmitted, &other, false},
		{"enviado sem departamento indicado", StatusSubmitted, nil, false},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, f.Match(Ref{Status: c.status, DepartmentID: c.dept}.Getter()), c.name)
	}

	t.Run("quem não tem departamento não enxerga nada", func(t *testing.T) {
		r := visibilityResult(t, ConditionSubmittedSameDepartment, authz.Subject{})
		assert.True(t, r.Filter().IsFalse())
		assert.False(t, r.Filter().Match(Ref{Status: StatusSubmitted, DepartmentID: &mine}.Getter()))
	})

	t.Run("vira SQL parametrizado com E", func(t *testing.T) {
		sql, args, err := f.ToSQL(Schema.Column, 1)
		require.NoError(t, err)
		assert.Contains(t, sql, "status IN")
		assert.Contains(t, sql, " AND ")
		assert.Contains(t, sql, "department_id =")
		assert.Equal(t, []any{StatusSubmitted, StatusArchived, mine}, args)
	})
}

// Documenta por que o E fica dentro de uma condição: duas condições na mesma permissão se combinam com OR, e o resultado deixaria
// o Revisor ler um rascunho do próprio departamento.
func TestSomarDuasCondicoesDariaOU(t *testing.T) {
	mine := uuid.New()
	subject := authz.Subject{DepartmentID: &mine}
	sameDept := authz.Eq(authz.FieldDepartmentID, mine)
	submitted := visibilityResult(t, ConditionSubmitted, subject).Filter()

	draftOfMyDept := Ref{Status: StatusDraft, DepartmentID: &mine}.Getter()
	assert.True(t, authz.Or(sameDept, submitted).Match(draftOfMyDept), "OR deixa passar o rascunho (o erro que a condição composta evita)")
	assert.False(t, visibilityResult(t, ConditionSubmittedSameDepartment, subject).Filter().Match(draftOfMyDept))
}
