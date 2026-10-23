// Pacote catalog mantém os cadastros de referência do cadastro de projetos: tipos de entidade,
// municípios (IBGE, somente leitura) e parlamentares.
package catalog

import "github.com/wagner/convenios-e-fomentos/backend/internal/authz"

// Ações administrativas. A leitura de cada catálogo é aberta ao papel de proponente, porque o
// cadastro do projeto usa as listas; a escrita é da administração.
const (
	ResourceEntityTypes     = "catalog.entity-types"
	ResourceMunicipalities  = "catalog.municipalities"
	ResourceParliamentarian = "catalog.parliamentarians"

	ActionEntityTypesRead   = "admin::catalog.entity-types.read"
	ActionEntityTypesCreate = "admin::catalog.entity-types.create"
	ActionEntityTypesUpdate = "admin::catalog.entity-types.update"
	ActionEntityTypesDelete = "admin::catalog.entity-types.delete"

	ActionMunicipalitiesRead = "admin::catalog.municipalities.read"

	ActionParliamentariansRead   = "admin::catalog.parliamentarians.read"
	ActionParliamentariansCreate = "admin::catalog.parliamentarians.create"
	ActionParliamentariansUpdate = "admin::catalog.parliamentarians.update"
	ActionParliamentariansDelete = "admin::catalog.parliamentarians.delete"
)

// Register registra as ações no motor de permissões.
func Register(reg *authz.Registry) {
	act := func(id, resource, name string) authz.Action {
		return authz.Action{ID: id, Domain: authz.DomainAdmin, Category: authz.CategorySettings, Name: name, Resource: resource}
	}
	reg.MustRegister(
		act(ActionEntityTypesCreate, ResourceEntityTypes, "Criar tipos de entidade"),
		act(ActionEntityTypesRead, ResourceEntityTypes, "Ler tipos de entidade"),
		act(ActionEntityTypesUpdate, ResourceEntityTypes, "Atualizar tipos de entidade"),
		act(ActionEntityTypesDelete, ResourceEntityTypes, "Excluir tipos de entidade"),
		act(ActionMunicipalitiesRead, ResourceMunicipalities, "Ler municípios"),
		act(ActionParliamentariansCreate, ResourceParliamentarian, "Criar parlamentares"),
		act(ActionParliamentariansRead, ResourceParliamentarian, "Ler parlamentares"),
		act(ActionParliamentariansUpdate, ResourceParliamentarian, "Atualizar parlamentares"),
		act(ActionParliamentariansDelete, ResourceParliamentarian, "Excluir parlamentares"),
	)
}
