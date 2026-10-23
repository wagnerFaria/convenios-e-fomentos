package rules

import "github.com/wagner/convenios-e-fomentos/backend/internal/authz"

// Ações administrativas do recurso. A leitura é aberta ao papel de proponente: o cadastro do projeto
// mostra prazos e limites vigentes.
const (
	Resource     = "settings.rule-parameters"
	ActionRead   = "admin::settings.rule-parameters.read"
	ActionCreate = "admin::settings.rule-parameters.create"
	ActionUpdate = "admin::settings.rule-parameters.update"
	ActionDelete = "admin::settings.rule-parameters.delete"
)

// Register registra as ações no motor de permissões.
func Register(reg *authz.Registry) {
	act := func(id, name string) authz.Action {
		return authz.Action{ID: id, Domain: authz.DomainAdmin, Category: authz.CategorySettings, Name: name, Resource: Resource}
	}
	reg.MustRegister(
		act(ActionCreate, "Criar parâmetros de regras"),
		act(ActionRead, "Ler parâmetros de regras"),
		act(ActionUpdate, "Atualizar parâmetros de regras"),
		act(ActionDelete, "Excluir parâmetros de regras"),
	)
}
