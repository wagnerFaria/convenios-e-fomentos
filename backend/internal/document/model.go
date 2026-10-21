// Pacote document é o recurso de negócio de exemplo ("documentos"). Ele existe para
// demonstrar, de ponta a ponta, o motor de permissões: permissões por ação, por
// campo e por condição na API administrativa, e permissões por ação com policies na
// API pública. Serve de modelo para novos recursos protegidos (ver README).
package document

import (
	"time"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

// Valores do campo status.
const (
	StatusDraft     = "rascunho"
	StatusPublished = "publicado"
)

// Resource é o nome do recurso nas permissões (subject).
const Resource = "document"

// Document é o documento com todos os campos. As tags alimentam o authz.Schema:
// "json" = nome lógico (usado em permissões, filtros e respostas), "db" = coluna e
// authz:"system" = campo de sistema (sempre legível, nunca gravável pelo cliente).
type Document struct {
	ID                uuid.UUID  `json:"id" db:"id" authz:"system"`
	Titulo            string     `json:"titulo" db:"titulo"`
	Conteudo          string     `json:"conteudo" db:"conteudo"`
	Status            string     `json:"status" db:"status"`
	ObservacaoInterna string     `json:"observacao_interna" db:"observacao_interna"`
	DepartmentID      *uuid.UUID `json:"department_id" db:"department_id"`
	PublishedAt       *time.Time `json:"published_at" db:"published_at" authz:"system"`
	CreatedByID       *uuid.UUID `json:"created_by_id" db:"created_by_id" authz:"system"`
	UpdatedByID       *uuid.UUID `json:"updated_by_id" db:"updated_by_id" authz:"system"`
	OwnerUserID       *uuid.UUID `json:"owner_user_id" db:"owner_user_id" authz:"system"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at" authz:"system"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at" authz:"system"`
}

func fromDB(d db.Document) Document { return Document(d) }

// Getter expõe os campos do documento para avaliar filtros de permissão em memória.
func (d Document) Getter() authz.Getter {
	m := map[string]any{
		"id": d.ID, "titulo": d.Titulo, "conteudo": d.Conteudo, "status": d.Status, "observacao_interna": d.ObservacaoInterna,
		"department_id": d.DepartmentID, "published_at": d.PublishedAt, "created_by_id": d.CreatedByID,
		"updated_by_id": d.UpdatedByID, "owner_user_id": d.OwnerUserID, "created_at": d.CreatedAt, "updated_at": d.UpdatedAt,
	}
	return func(field string) (any, bool) { v, ok := m[field]; return v, ok }
}

// Schema descreve os campos do documento para o motor de permissões.
var Schema = func() *authz.Schema {
	s, err := authz.NewSchema(Resource, Document{})
	if err != nil {
		panic(err)
	}
	return s
}()

// PublicDocument é a visão da API pública: não tem observacao_interna, status,
// department_id nem campos de autoria. O que não existe na struct não pode vazar.
type PublicDocument struct {
	ID          uuid.UUID  `json:"id"`
	Titulo      string     `json:"titulo"`
	Conteudo    string     `json:"conteudo"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func publicOf(d Document) PublicDocument {
	return PublicDocument{ID: d.ID, Titulo: d.Titulo, Conteudo: d.Conteudo, PublishedAt: d.PublishedAt, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt}
}

// Ações administrativas e da API pública do recurso.
const (
	ActionAdminCreate  = "admin::content.document.create"
	ActionAdminRead    = "admin::content.document.read"
	ActionAdminUpdate  = "admin::content.document.update"
	ActionAdminDelete  = "admin::content.document.delete"
	ActionAdminPublish = "admin::content.document.publish"

	ActionAPIFind    = "api::document.find"
	ActionAPIFindOne = "api::document.findOne"
	ActionAPICreate  = "api::document.create"
	ActionAPIUpdate  = "api::document.update"
	ActionAPIDelete  = "api::document.delete"
)

// Register registra as ações e o schema do recurso no Action Registry.
func Register(reg *authz.Registry) {
	admin := func(id, name string, fields, conds bool) authz.Action {
		return authz.Action{ID: id, Domain: authz.DomainAdmin, Category: authz.CategoryCollectionTypes, Name: name, Resource: Resource,
			SupportsFields: fields, SupportsConditions: conds, SupportsLocales: true}
	}
	api := func(id, name, method, route string) authz.Action {
		return authz.Action{ID: id, Domain: authz.DomainAPI, Category: authz.CategoryApplication, Name: name, Resource: Resource, Method: method, Route: route}
	}
	reg.MustRegister(
		admin(ActionAdminCreate, "Criar documentos", true, false),
		admin(ActionAdminRead, "Ler documentos", true, true),
		admin(ActionAdminUpdate, "Atualizar documentos", true, true),
		admin(ActionAdminDelete, "Excluir documentos", false, true),
		admin(ActionAdminPublish, "Publicar documentos", false, true),
		api(ActionAPIFind, "Listar documentos publicados", "GET", "/api/documents"),
		api(ActionAPIFindOne, "Detalhar documento publicado", "GET", "/api/documents/{id}"),
		api(ActionAPICreate, "Criar documento", "POST", "/api/documents"),
		api(ActionAPIUpdate, "Atualizar o próprio documento", "PUT", "/api/documents/{id}"),
		api(ActionAPIDelete, "Excluir o próprio documento", "DELETE", "/api/documents/{id}"),
	)
	reg.RegisterSchema(Resource, Schema)
}
