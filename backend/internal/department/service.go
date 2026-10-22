// Pacote department administra os departamentos, que delimitam o alcance da condição
// `admin::same-department` e agrupam administradores e documentos.
package department

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// View é o departamento exposto pela API, com o que está vinculado a ele.
type View struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	UserCount     int64     `json:"userCount"`
	DocumentCount int64     `json:"documentCount"`
}

// Actor é o administrador que executa a operação, para a auditoria.
type Actor struct {
	ID uuid.UUID
	IP string
}

func (a Actor) entry(action string, id uuid.UUID, before, after any) audit.Entry {
	payload := map[string]any{}
	if before != nil {
		payload["before"] = before
	}
	if after != nil {
		payload["after"] = after
	}
	return audit.Entry{ActorType: "admin", ActorID: a.ID.String(), Action: action, Entity: "department", EntityID: id.String(), Payload: payload, IP: a.IP}
}

func errNameTaken() error {
	return httpapi.ErrConflict("department_name_taken", "Já existe um departamento com este nome.")
}

func errInUse(users, documents int64) error {
	e := httpapi.ErrConflict("department_in_use", "O departamento tem administradores ou documentos vinculados e não pode ser excluído. Mude a lotação deles antes.")
	e.Details = map[string]any{"userCount": users, "documentCount": documents}
	return e
}

func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.ErrNotFound()
	}
	return err
}

// cleanName normaliza e valida o nome.
func cleanName(name string) (string, error) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || len([]rune(name)) > 100 {
		return "", httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: "name", Message: "Informe um nome de até 100 caracteres."}}})
	}
	return name, nil
}

// Service implementa as operações sobre departamentos.
type Service struct {
	pool  *pgxpool.Pool
	q     *db.Queries
	audit audit.Recorder
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool) *Service { return &Service{pool: pool, q: db.New(pool)} }

// List devolve todos os departamentos por nome.
func (s *Service) List(ctx context.Context) ([]View, error) {
	rows, err := s.q.ListDepartmentsWithCounts(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]View, len(rows))
	for i, r := range rows {
		out[i] = View{ID: r.ID, Name: r.Name, UserCount: r.UserCount, DocumentCount: r.DocumentCount}
	}
	return out, nil
}

// Get devolve um departamento.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*View, error) {
	return s.view(ctx, s.q, id)
}

func (s *Service) view(ctx context.Context, q *db.Queries, id uuid.UUID) (*View, error) {
	r, err := q.GetDepartmentWithCounts(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	return &View{ID: r.ID, Name: r.Name, UserCount: r.UserCount, DocumentCount: r.DocumentCount}, nil
}

// Create cria um departamento.
func (s *Service) Create(ctx context.Context, actor Actor, name string) (*View, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, err
	}
	var out *View
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		d, err := q.CreateDepartment(ctx, name)
		if database.IsUniqueViolation(err) {
			return errNameTaken()
		}
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("department.create", d.ID, nil, map[string]any{"name": d.Name})); err != nil {
			return err
		}
		out, err = s.view(ctx, q, d.ID)
		return err
	})
	return out, err
}

// Update renomeia um departamento.
func (s *Service) Update(ctx context.Context, actor Actor, id uuid.UUID, name string) (*View, error) {
	name, err := cleanName(name)
	if err != nil {
		return nil, err
	}
	var out *View
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetDepartmentForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		after, err := q.UpdateDepartment(ctx, db.UpdateDepartmentParams{ID: id, Name: name})
		if database.IsUniqueViolation(err) {
			return errNameTaken()
		}
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("department.update", id, map[string]any{"name": before.Name}, map[string]any{"name": after.Name})); err != nil {
			return err
		}
		out, err = s.view(ctx, q, id)
		return err
	})
	return out, err
}

// Delete exclui o departamento se nenhum administrador ou documento depender dele. Sem essa
// checagem a exclusão deixaria os vinculados sem departamento e mudaria em silêncio o alcance
// de quem tem permissões `admin::same-department`.
func (s *Service) Delete(ctx context.Context, actor Actor, id uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		d, err := q.GetDepartmentForUpdate(ctx, id) // a trava impede novos vínculos durante a exclusão
		if err != nil {
			return notFound(err)
		}
		v, err := s.view(ctx, q, id)
		if err != nil {
			return err
		}
		if v.UserCount > 0 || v.DocumentCount > 0 {
			return errInUse(v.UserCount, v.DocumentCount)
		}
		if _, err := q.DeleteDepartment(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("department.delete", id, map[string]any{"name": d.Name}, nil))
	})
}
