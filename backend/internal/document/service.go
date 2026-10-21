package document

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

const (
	maxTitulo     = 200
	maxConteudo   = 100_000
	maxObservacao = 5_000
)

// Service implementa a regra de negócio dos documentos.
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	engine *authz.Engine
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, engine *authz.Engine) *Service {
	return &Service{pool: pool, q: db.New(pool), engine: engine}
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.ErrNotFound()
	}
	return err
}

func fieldErrors(msgs ...httpapi.FieldError) error {
	return httpapi.ErrValidation(map[string]any{"fields": msgs})
}

// Get devolve o documento ou 404.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (Document, error) {
	d, err := s.q.GetDocument(ctx, id)
	if err != nil {
		return Document{}, notFound(err)
	}
	return fromDB(d), nil
}

// ListParams são os parâmetros da listagem (administrativa ou pública).
type ListParams struct {
	// Perm restringe os registros pela permissão (admin). FilterTrue = sem restrição.
	Perm          authz.Filter
	OnlyPublished bool
	Q             string
	Status        string
	DepartmentID  *uuid.UUID
	Sort          httpapi.Sort
	Page          httpapi.Page
}

// AdminSortColumns e PublicSortColumns são as listas permitidas de ordenação.
var (
	AdminSortColumns  = map[string]string{"titulo": "titulo", "status": "status", "created_at": "created_at", "updated_at": "updated_at", "published_at": "published_at"}
	PublicSortColumns = map[string]string{"titulo": "titulo", "published_at": "published_at", "created_at": "created_at"}
)

const selectColumns = `id, titulo, conteudo, status, observacao_interna, department_id, published_at, created_by_id, updated_by_id, owner_user_id, created_at, updated_at`

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// List lista documentos combinando SEMPRE o filtro de permissão com os filtros do
// usuário. O filtro de permissão vira cláusula parametrizada apenas com as colunas do
// Schema; a coluna de ordenação vem de uma lista permitida.
func (s *Service) List(ctx context.Context, p ListParams) ([]Document, int, error) {
	var conds []string
	var args []any
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }

	if !p.Perm.IsTrue() {
		clause, pargs, err := p.Perm.ToSQL(Schema.Column, 1)
		if err != nil {
			return nil, 0, err
		}
		args = append(args, pargs...)
		conds = append(conds, clause)
	}
	if p.OnlyPublished {
		conds = append(conds, "status = '"+StatusPublished+"'")
	}
	if p.Status != "" {
		conds = append(conds, "status = "+arg(p.Status))
	}
	if p.DepartmentID != nil {
		conds = append(conds, "department_id = "+arg(*p.DepartmentID))
	}
	if p.Q != "" {
		ph := arg("%" + escapeLike(p.Q) + "%")
		conds = append(conds, fmt.Sprintf("(titulo ILIKE %[1]s OR conteudo ILIKE %[1]s)", ph))
	}
	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM documents"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	// Desempate por id mantém a paginação estável.
	query := fmt.Sprintf("SELECT %s FROM documents%s ORDER BY %s NULLS LAST, id LIMIT %d OFFSET %d", selectColumns, where, p.Sort.SQL(), p.Page.PageSize, p.Page.Offset())
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []Document{}
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Titulo, &d.Conteudo, &d.Status, &d.ObservacaoInterna, &d.DepartmentID, &d.PublishedAt, &d.CreatedByID, &d.UpdatedByID, &d.OwnerUserID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

// Patch traz os campos editáveis de um documento. Keys lista os campos realmente
// enviados (para distinguir ausente de null, ex.: department_id: null limpa o campo).
type Patch struct {
	Titulo            *string
	Conteudo          *string
	Status            *string
	ObservacaoInterna *string
	DepartmentID      *uuid.UUID
	Keys              []string
}

// Validate confere tamanhos e valores dos campos enviados.
func (p Patch) Validate(requireTitulo bool) error {
	var issues []httpapi.FieldError
	bad := func(f, m string) { issues = append(issues, httpapi.FieldError{Field: f, Message: m}) }
	if p.Titulo == nil && requireTitulo {
		bad("titulo", "Campo obrigatório.")
	}
	if p.Titulo != nil && (strings.TrimSpace(*p.Titulo) == "" || len(*p.Titulo) > maxTitulo) {
		bad("titulo", fmt.Sprintf("Informe de 1 a %d caracteres.", maxTitulo))
	}
	if p.Conteudo != nil && len(*p.Conteudo) > maxConteudo {
		bad("conteudo", "Conteúdo longo demais.")
	}
	if p.ObservacaoInterna != nil && len(*p.ObservacaoInterna) > maxObservacao {
		bad("observacao_interna", "Observação longa demais.")
	}
	if p.Status != nil && *p.Status != StatusDraft && *p.Status != StatusPublished {
		bad("status", "Use rascunho ou publicado.")
	}
	if len(issues) > 0 {
		return fieldErrors(issues...)
	}
	return nil
}

func (p Patch) has(key string) bool {
	for _, k := range p.Keys {
		if k == key {
			return true
		}
	}
	return false
}

// AdminActor é o administrador que age sobre o documento.
type AdminActor struct {
	ID           uuid.UUID
	Subject      authz.Subject
	DepartmentID *uuid.UUID
}

func rejected(fields []string) error {
	list := make([]httpapi.FieldError, len(fields))
	for i, f := range fields {
		list[i] = httpapi.FieldError{Field: f, Message: "Você não tem permissão para gravar este campo."}
	}
	return fieldErrors(list...)
}

// CreateByAdmin cria um documento como rascunho. allowed são os campos que o
// administrador pode gravar (decisão da ação create). Sem department_id, usa o
// departamento do administrador.
func (s *Service) CreateByAdmin(ctx context.Context, actor AdminActor, p Patch, allowed authz.FieldSet) (Document, error) {
	if p.Status != nil || p.has("status") {
		return Document{}, fieldErrors(httpapi.FieldError{Field: "status", Message: "Documentos nascem como rascunho; use a ação de publicar."})
	}
	if err := p.Validate(true); err != nil {
		return Document{}, err
	}
	if bad := Schema.RejectedWrite(p.Keys, allowed); len(bad) > 0 {
		return Document{}, rejected(bad)
	}
	dept := p.DepartmentID
	if !p.has("department_id") {
		dept = actor.DepartmentID
	}
	row, err := s.q.InsertDocument(ctx, db.InsertDocumentParams{Titulo: *p.Titulo, Conteudo: deref(p.Conteudo), ObservacaoInterna: deref(p.ObservacaoInterna),
		DepartmentID: dept, CreatedByID: &actor.ID})
	if database.IsForeignKeyViolation(err) {
		return Document{}, fieldErrors(httpapi.FieldError{Field: "department_id", Message: "Departamento inexistente."})
	}
	if err != nil {
		return Document{}, err
	}
	return fromDB(row), nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// UpdateByAdmin edita o documento. upd é a decisão da ação update: o registro precisa
// passar pelo filtro (403 se não) e só os campos que ela libera para ESTE registro podem
// ser enviados (422 com a lista dos recusados). Mudar o status exige também a ação publish.
func (s *Service) UpdateByAdmin(ctx context.Context, actor AdminActor, id uuid.UUID, p Patch, upd authz.Decision) (Document, error) {
	if err := p.Validate(false); err != nil {
		return Document{}, err
	}
	var out Document
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.GetDocumentForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		cur := fromDB(row)
		if !upd.AllowsRecord(cur.Getter()) {
			return httpapi.ErrForbidden()
		}
		if bad := Schema.RejectedWrite(p.Keys, upd.FieldsFor(cur.Getter())); len(bad) > 0 {
			return rejected(bad)
		}
		if p.Status != nil && *p.Status != cur.Status {
			pub, err := s.engine.Can(ctx, actor.Subject, ActionAdminPublish)
			if err != nil {
				return err
			}
			if !pub.AllowsRecord(cur.Getter()) {
				e := httpapi.ErrForbidden()
				e.Details = map[string]string{"action": ActionAdminPublish}
				return e
			}
		}
		next := cur
		if p.Titulo != nil {
			next.Titulo = *p.Titulo
		}
		if p.Conteudo != nil {
			next.Conteudo = *p.Conteudo
		}
		if p.ObservacaoInterna != nil {
			next.ObservacaoInterna = *p.ObservacaoInterna
		}
		if p.has("department_id") {
			next.DepartmentID = p.DepartmentID
		}
		if p.Status != nil {
			applyStatus(&next, *p.Status)
		}
		saved, err := q.UpdateDocument(ctx, updateParams(next, &actor.ID))
		if database.IsForeignKeyViolation(err) {
			return fieldErrors(httpapi.FieldError{Field: "department_id", Message: "Departamento inexistente."})
		}
		if err != nil {
			return err
		}
		out = fromDB(saved)
		return nil
	})
	return out, err
}

func applyStatus(d *Document, status string) {
	if status == d.Status {
		return
	}
	d.Status = status
	if status == StatusPublished {
		now := time.Now()
		d.PublishedAt = &now
	} else {
		d.PublishedAt = nil
	}
}

func updateParams(d Document, updatedBy *uuid.UUID) db.UpdateDocumentParams {
	return db.UpdateDocumentParams{ID: d.ID, Titulo: d.Titulo, Conteudo: d.Conteudo, Status: d.Status, ObservacaoInterna: d.ObservacaoInterna,
		DepartmentID: d.DepartmentID, PublishedAt: d.PublishedAt, UpdatedByID: updatedBy}
}

// SetPublished publica ou despublica. pub é a decisão da ação publish (403 se o registro não passar no filtro).
func (s *Service) SetPublished(ctx context.Context, actor AdminActor, id uuid.UUID, publish bool, pub authz.Decision) (Document, error) {
	var out Document
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.GetDocumentForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		cur := fromDB(row)
		if !pub.AllowsRecord(cur.Getter()) {
			return httpapi.ErrForbidden()
		}
		next := cur
		status := StatusDraft
		if publish {
			status = StatusPublished
		}
		applyStatus(&next, status)
		saved, err := q.UpdateDocument(ctx, updateParams(next, &actor.ID))
		if err != nil {
			return err
		}
		out = fromDB(saved)
		return nil
	})
	return out, err
}

// DeleteByAdmin exclui o documento; del é a decisão da ação delete.
func (s *Service) DeleteByAdmin(ctx context.Context, id uuid.UUID, del authz.Decision) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.GetDocumentForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if !del.AllowsRecord(fromDB(row).Getter()) {
			return httpapi.ErrForbidden()
		}
		_, err = q.DeleteDocument(ctx, id)
		return err
	})
}

// CreateByUser cria um rascunho pertencente ao usuário final (owner = quem chamou;
// nulo quando a chamada vem de um token de API).
func (s *Service) CreateByUser(ctx context.Context, owner *uuid.UUID, p Patch) (Document, error) {
	if err := p.Validate(true); err != nil {
		return Document{}, err
	}
	row, err := s.q.InsertDocument(ctx, db.InsertDocumentParams{Titulo: *p.Titulo, Conteudo: deref(p.Conteudo), OwnerUserID: owner})
	if err != nil {
		return Document{}, err
	}
	return fromDB(row), nil
}

// UpdateByUser edita título/conteúdo. A checagem de propriedade é feita pela policy is-owner na rota.
func (s *Service) UpdateByUser(ctx context.Context, id uuid.UUID, p Patch) (Document, error) {
	if err := p.Validate(false); err != nil {
		return Document{}, err
	}
	var out Document
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.GetDocumentForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		next := fromDB(row)
		if p.Titulo != nil {
			next.Titulo = *p.Titulo
		}
		if p.Conteudo != nil {
			next.Conteudo = *p.Conteudo
		}
		saved, err := q.UpdateDocument(ctx, updateParams(next, next.UpdatedByID))
		if err != nil {
			return err
		}
		out = fromDB(saved)
		return nil
	})
	return out, err
}

// DeleteByUser exclui o documento (a propriedade é checada pela policy na rota).
func (s *Service) DeleteByUser(ctx context.Context, id uuid.UUID) error {
	n, err := s.q.DeleteDocument(ctx, id)
	if err != nil {
		return err
	}
	if n == 0 {
		return httpapi.ErrNotFound()
	}
	return nil
}
