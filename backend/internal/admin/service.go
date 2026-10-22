// Pacote admin trata dos administradores, suas roles e do pré-cadastro por CPF.
package admin

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/cpf"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Dados da role especial Super Admin.
const (
	SuperAdminCode        = authz.SuperAdminRoleCode
	superAdminName        = "Super Admin"
	superAdminDescription = "Acesso total ao sistema."
)

var (
	errCPFTaken = &httpapi.APIError{Status: 409, Code: "cpf_taken", Message: "Já existe um administrador com este CPF."}
	errBadCPF   = fieldIssue("cpf", "CPF inválido.")
)

// Service implementa as operações sobre administradores.
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	audit  audit.Recorder
	engine *authz.Engine
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, engine *authz.Engine) *Service {
	return &Service{pool: pool, q: db.New(pool), engine: engine}
}

// EnsureSuperAdminInput identifica o Super Admin: o CPF é obrigatório; nome e e-mail são
// opcionais porque o primeiro acesso com o MT Login os preenche.
type EnsureSuperAdminInput struct {
	CPF, FirstName, LastName, Email string
}

// EnsureSuperAdmin garante um administrador ativo com a role Super Admin para o CPF: cria o
// registro se não existir ou apenas concede a role ao existente. É idempotente.
func (s *Service) EnsureSuperAdmin(ctx context.Context, in EnsureSuperAdminInput, ip string) (db.Admin, error) {
	digits, err := cpf.Normalize(in.CPF)
	if err != nil {
		return db.Admin{}, errBadCPF
	}
	var admin db.Admin
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: SuperAdminCode, Name: superAdminName, Description: superAdminDescription, IsSystem: true})
		if err != nil {
			return err
		}
		admin, err = q.GetAdminByCPF(ctx, &digits)
		created := errors.Is(err, pgx.ErrNoRows)
		if created {
			admin, err = q.CreateAdmin(ctx, db.CreateAdminParams{Cpf: &digits, FirstName: in.FirstName, LastName: in.LastName, Email: strings.TrimSpace(in.Email), IsActive: true})
		}
		if err != nil {
			return err
		}
		if err := q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: admin.ID, RoleID: role.ID}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "system", Action: "admin.ensure-super-admin", Entity: "admin", EntityID: admin.ID.String(), IP: ip,
			Payload: map[string]any{"after": map[string]any{"cpf": cpf.Mask(digits), "created": created}}})
	})
	return admin, err
}

// RegisterInput são os dados do pré-cadastro: o administrador é criado antes do primeiro acesso
// e o login com o MT Login, pelo CPF, completa nome e e-mail.
type RegisterInput struct {
	CPF          string
	RoleIDs      []uuid.UUID
	DepartmentID *uuid.UUID
}

// Register pré-cadastra um administrador pelo CPF, já com as roles.
func (s *Service) Register(ctx context.Context, actor *auth.AdminPrincipal, in RegisterInput, ip string) (*UserView, error) {
	digits, err := cpf.Normalize(in.CPF)
	if err != nil {
		return nil, errBadCPF
	}
	roleIDs := slices.Clone(in.RoleIDs)
	slices.SortFunc(roleIDs, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	roleIDs = slices.Compact(roleIDs)

	var out *UserView
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		roles, err := q.GetAdminRolesByIDs(ctx, roleIDs)
		if err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return fieldIssue("roleIds", "Uma ou mais roles não existem.")
		}
		// Só um Super Admin pode conceder a role Super Admin (evita escalada de privilégio).
		if hasSuperRole(roles) && !SubjectOf(actor).IsSuperAdmin() {
			return httpapi.ErrForbidden()
		}
		if err := s.ensureCoversRoles(ctx, ActorOf(actor, ip), roleIDs); err != nil {
			return err
		}
		created, err := q.CreateAdmin(ctx, db.CreateAdminParams{Cpf: &digits, IsActive: true, DepartmentID: in.DepartmentID})
		switch {
		case database.IsUniqueViolation(err):
			return errCPFTaken
		case database.IsForeignKeyViolation(err):
			return fieldIssue("departmentId", "Departamento inexistente.")
		case err != nil:
			return err
		}
		for _, id := range roleIDs {
			if err := q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: created.ID, RoleID: id}); err != nil {
				return err
			}
		}
		if err := s.audit.Record(ctx, q, audit.Entry{ActorType: "admin", ActorID: actor.ID.String(), Action: "admin.register", Entity: "admin", EntityID: created.ID.String(), IP: ip,
			Payload: map[string]any{"after": map[string]any{"cpf": cpf.Mask(digits), "roleIds": roleIDs}}}); err != nil {
			return err
		}
		out, err = s.userView(ctx, q, created.ID)
		return err
	})
	return out, err
}

// RoleView é a role resumida exposta pela API.
type RoleView struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

// MeView é o administrador autenticado com as permissões efetivas (para o
// front esconder botões; a autorização real é sempre feita no servidor).
type MeView struct {
	ID           uuid.UUID                   `json:"id"`
	CPF          string                      `json:"cpf"`
	Email        string                      `json:"email"`
	FirstName    string                      `json:"firstName"`
	LastName     string                      `json:"lastName"`
	DepartmentID *uuid.UUID                  `json:"departmentId"`
	IsSuperAdmin bool                        `json:"isSuperAdmin"`
	Roles        []RoleView                  `json:"roles"`
	Permissions  []authz.EffectivePermission `json:"permissions"`
}

// Me monta a visão do administrador autenticado.
func (s *Service) Me(ctx context.Context, p *auth.AdminPrincipal) (*MeView, error) {
	subj := SubjectOf(p)
	perms, err := s.engine.Effective(ctx, subj)
	if err != nil {
		return nil, err
	}
	roles, err := s.q.ListRolesForAdmin(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	rv := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		rv = append(rv, RoleView{ID: r.ID, Code: r.Code, Name: r.Name})
	}
	return &MeView{ID: p.ID, CPF: cpf.Format(p.CPF), Email: p.Email, FirstName: p.FirstName, LastName: p.LastName, DepartmentID: p.DepartmentID,
		IsSuperAdmin: subj.IsSuperAdmin(), Roles: rv, Permissions: perms}, nil
}
