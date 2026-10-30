package project

import (
	"context"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
)

// Condições de visibilidade da equipe. Rascunho é do proponente (e dos membros que ele convidou): a equipe só enxerga projeto
// enviado ou arquivado.
//
// As condições de UMA permissão se combinam com OR, e as permissões de perfis diferentes também. Por isso o "E" entre status e
// departamento precisa estar DENTRO de uma única condição: listar `admin::same-department` ao lado de uma condição de status daria
// "mesmo departamento OU enviado", o contrário do que se quer.
const (
	// ConditionSubmitted: o projeto já foi enviado (ativo ou arquivado). Para Analista, Supervisor e Secretário.
	ConditionSubmitted = "admin::project-submitted"
	// ConditionSubmittedSameDepartment: projeto enviado E do departamento da pessoa. Para o Revisor do Departamento. Quem não
	// tem departamento no cadastro não enxerga nenhum projeto.
	ConditionSubmittedSameDepartment = "admin::project-submitted-same-department"
)

// submittedFilter casa os projetos que a equipe pode ver.
func submittedFilter() authz.Filter {
	return authz.In("status", StatusSubmitted, StatusArchived)
}

// registerVisibilityConditions registra as condições de visibilidade da equipe.
func registerVisibilityConditions(reg *authz.ConditionRegistry) {
	reg.MustRegister(
		authz.Condition{ID: ConditionSubmitted, Name: "Projeto enviado (não é rascunho)", Category: "default",
			Handler: func(context.Context, authz.Subject) (authz.Result, error) {
				return authz.Restrict(submittedFilter()), nil
			}},
		authz.Condition{ID: ConditionSubmittedSameDepartment, Name: "Projeto enviado do mesmo departamento", Category: "custom",
			Handler: func(_ context.Context, s authz.Subject) (authz.Result, error) {
				if s.DepartmentID == nil {
					return authz.Deny(), nil
				}
				return authz.Restrict(authz.And(submittedFilter(), authz.Eq(authz.FieldDepartmentID, *s.DepartmentID))), nil
			}},
	)
}
