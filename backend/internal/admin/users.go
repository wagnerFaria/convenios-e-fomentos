package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Status do administrador, na ordem de precedência: blocked, pending (convite
// não aceito), inactive, active.
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusPending  = "pending"
	StatusBlocked  = "blocked"
)

// UserView é o administrador exposto pela API.
type UserView struct {
	ID           uuid.UUID  `json:"id"`
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	Email        string     `json:"email"`
	Status       string     `json:"status"`
	IsActive     bool       `json:"isActive"`
	Blocked      bool       `json:"blocked"`
	DepartmentID *uuid.UUID `json:"departmentId"`
	Roles        []RoleView `json:"roles"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

func statusOf(a db.Admin) string {
	switch {
	case a.Blocked:
		return StatusBlocked
	case a.PasswordHash == nil:
		return StatusPending
	case a.IsActive:
		return StatusActive
	}
	return StatusInactive
}

func viewOf(a db.Admin, roles []db.AdminRole) UserView {
	rv := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		rv = append(rv, RoleView{ID: r.ID, Code: r.Code, Name: r.Name})
	}
	return UserView{ID: a.ID, FirstName: a.FirstName, LastName: a.LastName, Email: a.Email, Status: statusOf(a), IsActive: a.IsActive,
		Blocked: a.Blocked, DepartmentID: a.DepartmentID, Roles: rv, CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt}
}

// GetUser devolve um administrador com suas roles.
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*UserView, error) {
	return s.userView(ctx, s.q, id)
}

func (s *Service) userView(ctx context.Context, q *db.Queries, id uuid.UUID) (*UserView, error) {
	a, err := q.GetAdminByID(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	roles, err := q.ListRolesForAdmin(ctx, id)
	if err != nil {
		return nil, err
	}
	v := viewOf(a, roles)
	return &v, nil
}

// ListUsersFilter são os filtros da listagem.
type ListUsersFilter struct {
	Q      string // busca em nome, sobrenome e e-mail
	RoleID *uuid.UUID
	Status string
	Sort   httpapi.Sort
	Page   httpapi.Page
}

// UserSortColumns mapeia campos de ordenação da API para colunas.
var UserSortColumns = map[string]string{
	"firstName": "a.first_name", "lastName": "a.last_name", "email": "a.email", "createdAt": "a.created_at",
}

// UserStatuses são os valores aceitos no filtro de status.
var UserStatuses = []string{StatusActive, StatusInactive, StatusPending, StatusBlocked}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ListUsers lista administradores com busca, filtros, ordenação e paginação.
// A consulta é montada à mão porque filtros e ordenação são dinâmicos, mas só
// concatena fragmentos fixos: valores vão como parâmetros e a coluna de ordenação
// vem de uma lista permitida.
func (s *Service) ListUsers(ctx context.Context, f ListUsersFilter) ([]UserView, int, error) {
	var where []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, strings.ReplaceAll(cond, "?", fmt.Sprintf("$%d", len(args))))
	}
	if f.Q != "" {
		args = append(args, "%"+escapeLike(f.Q)+"%")
		n := len(args)
		where = append(where, fmt.Sprintf(`(a.first_name ILIKE $%[1]d OR a.last_name ILIKE $%[1]d OR a.email::text ILIKE $%[1]d OR (a.first_name || ' ' || a.last_name) ILIKE $%[1]d)`, n))
	}
	if f.RoleID != nil {
		add(`EXISTS (SELECT 1 FROM admin_users_roles ur WHERE ur.admin_id = a.id AND ur.role_id = ?)`, *f.RoleID)
	}
	switch f.Status {
	case StatusBlocked:
		where = append(where, `a.blocked`)
	case StatusPending:
		where = append(where, `NOT a.blocked AND a.password_hash IS NULL`)
	case StatusInactive:
		where = append(where, `NOT a.blocked AND a.password_hash IS NOT NULL AND NOT a.is_active`)
	case StatusActive:
		where = append(where, `NOT a.blocked AND a.is_active`)
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM admins a"+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT a.id, a.first_name, a.last_name, a.email, a.is_active, a.blocked, a.password_hash IS NOT NULL,
		a.department_id, a.created_at, a.updated_at,
		COALESCE((SELECT jsonb_agg(jsonb_build_object('id', r.id, 'code', r.code, 'name', r.name) ORDER BY r.name)
		          FROM admin_roles r JOIN admin_users_roles ur ON ur.role_id = r.id WHERE ur.admin_id = a.id), '[]'::jsonb)
		FROM admins a` + clause +
		fmt.Sprintf(" ORDER BY %s, a.id LIMIT %d OFFSET %d", f.Sort.SQL(), f.Page.PageSize, f.Page.Offset())
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []UserView{}
	for rows.Next() {
		var v UserView
		var hasPassword bool
		var rolesJSON []byte
		if err := rows.Scan(&v.ID, &v.FirstName, &v.LastName, &v.Email, &v.IsActive, &v.Blocked, &hasPassword, &v.DepartmentID, &v.CreatedAt, &v.UpdatedAt, &rolesJSON); err != nil {
			return nil, 0, err
		}
		if err := json.Unmarshal(rolesJSON, &v.Roles); err != nil {
			return nil, 0, err
		}
		v.Status = statusOf(db.Admin{Blocked: v.Blocked, IsActive: v.IsActive, PasswordHash: passwordMarker(hasPassword)})
		out = append(out, v)
	}
	return out, total, rows.Err()
}

func passwordMarker(has bool) *string {
	if !has {
		return nil
	}
	s := ""
	return &s
}

// UpdateUserInput são os dados editáveis de um administrador.
type UpdateUserInput struct {
	FirstName    string
	LastName     string
	Email        string
	DepartmentID *uuid.UUID
}

type userSnapshot struct {
	FirstName    string     `json:"firstName"`
	LastName     string     `json:"lastName"`
	Email        string     `json:"email"`
	DepartmentID *uuid.UUID `json:"departmentId"`
	IsActive     bool       `json:"isActive"`
	Roles        []string   `json:"roles,omitempty"`
}

func snapshotOf(a db.Admin) userSnapshot {
	return userSnapshot{FirstName: a.FirstName, LastName: a.LastName, Email: a.Email, DepartmentID: a.DepartmentID, IsActive: a.IsActive}
}

// UpdateUser edita os dados do administrador.
func (s *Service) UpdateUser(ctx context.Context, actor Actor, id uuid.UUID, in UpdateUserInput) (*UserView, error) {
	var out *UserView
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetAdminByID(ctx, id)
		if err != nil {
			return notFound(err)
		}
		after, err := q.UpdateAdmin(ctx, db.UpdateAdminParams{ID: id, FirstName: in.FirstName, LastName: in.LastName, Email: strings.TrimSpace(in.Email), DepartmentID: in.DepartmentID})
		switch {
		case database.IsUniqueViolation(err):
			return errEmailTaken
		case database.IsForeignKeyViolation(err):
			return fieldIssue("departmentId", "Departamento inexistente.")
		case err != nil:
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("admin.update", "admin", id, snapshotOf(before), snapshotOf(after))); err != nil {
			return err
		}
		out, err = s.userView(ctx, q, id)
		return err
	})
	return out, err
}

// withSuperGuard executa fn em transação e desfaz tudo se a operação tirar o
// último Super Admin ativo. O lock consultivo serializa operações concorrentes,
// evitando que duas remoções simultâneas passem juntas pela verificação.
func (s *Service) withSuperGuard(ctx context.Context, fn func(q *db.Queries) error) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if err := q.LockSuperAdmins(ctx); err != nil {
			return err
		}
		before, err := q.CountActiveSuperAdmins(ctx, SuperAdminCode)
		if err != nil {
			return err
		}
		if err := fn(q); err != nil {
			return err
		}
		after, err := q.CountActiveSuperAdmins(ctx, SuperAdminCode)
		if err != nil {
			return err
		}
		if before > 0 && after == 0 {
			return errLastSuperAdmin
		}
		return nil
	})
}

// SetActive ativa ou desativa o administrador. Desativar revoga as sessões e
// não é permitido sobre si mesmo nem sobre o último Super Admin ativo.
func (s *Service) SetActive(ctx context.Context, actor Actor, id uuid.UUID, active bool) (*UserView, error) {
	if !active && actor.ID == id {
		return nil, errCannotModifySelf
	}
	var out *UserView
	err := s.withSuperGuard(ctx, func(q *db.Queries) error {
		before, err := q.GetAdminByID(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if active && before.PasswordHash == nil {
			return errInvitationPending
		}
		if err := q.SetAdminActive(ctx, db.SetAdminActiveParams{ID: id, IsActive: active}); err != nil {
			return err
		}
		if !active {
			if err := q.RevokeAllAdminSessions(ctx, id); err != nil {
				return err
			}
		}
		action := "admin.deactivate"
		if active {
			action = "admin.activate"
		}
		after := before
		after.IsActive = active
		if err := s.audit.Record(ctx, q, actor.entry(action, "admin", id, snapshotOf(before), snapshotOf(after))); err != nil {
			return err
		}
		out, err = s.userView(ctx, q, id)
		return err
	})
	return out, err
}

// MaxBatch é o limite de itens em operações em lote.
const MaxBatch = 100

// DeleteUsers exclui um ou mais administradores de uma vez (tudo ou nada).
func (s *Service) DeleteUsers(ctx context.Context, actor Actor, ids []uuid.UUID) error {
	ids = uniqueIDs(ids)
	if len(ids) == 0 || len(ids) > MaxBatch {
		return fieldIssue("ids", fmt.Sprintf("Informe de 1 a %d administradores.", MaxBatch))
	}
	if slices.Contains(ids, actor.ID) {
		return errCannotModifySelf
	}
	return s.withSuperGuard(ctx, func(q *db.Queries) error {
		found, err := q.GetAdminsByIDs(ctx, ids)
		if err != nil {
			return err
		}
		if len(found) != len(ids) {
			return httpapi.ErrNotFound()
		}
		if _, err := q.DeleteAdminsByIDs(ctx, ids); err != nil {
			return err
		}
		for _, a := range found {
			if err := s.audit.Record(ctx, q, actor.entry("admin.delete", "admin", a.ID, snapshotOf(a), nil)); err != nil {
				return err
			}
		}
		return nil
	})
}

func uniqueIDs(ids []uuid.UUID) []uuid.UUID {
	seen := map[uuid.UUID]bool{}
	out := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func roleCodes(roles []db.AdminRole) []string {
	out := make([]string, len(roles))
	for i, r := range roles {
		out[i] = r.Code
	}
	return out
}

func hasSuperRole(roles []db.AdminRole) bool {
	return slices.ContainsFunc(roles, func(r db.AdminRole) bool { return r.Code == SuperAdminCode })
}

// AddRoles atribui roles ao administrador. Só um Super Admin concede a role Super Admin.
func (s *Service) AddRoles(ctx context.Context, actor Actor, id uuid.UUID, roleIDs []uuid.UUID) (*UserView, error) {
	roleIDs = uniqueIDs(roleIDs)
	var out *UserView
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := q.GetAdminByID(ctx, id); err != nil {
			return notFound(err)
		}
		roles, err := q.GetAdminRolesByIDs(ctx, roleIDs)
		if err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return fieldIssue("roleIds", "Uma ou mais roles não existem.")
		}
		if hasSuperRole(roles) && !actor.Subject.IsSuperAdmin() {
			return httpapi.ErrForbidden()
		}
		current, err := q.ListRolesForAdmin(ctx, id)
		if err != nil {
			return err
		}
		for _, r := range roles {
			if err := q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: id, RoleID: r.ID}); err != nil {
				return err
			}
		}
		out, err = s.userView(ctx, q, id)
		if err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("admin.roles-add", "admin", id,
			map[string]any{"roles": roleCodes(current)}, map[string]any{"roles": codesOfView(out.Roles)}))
	})
	return out, err
}

func codesOfView(rs []RoleView) []string {
	out := make([]string, len(rs))
	for i, r := range rs {
		out[i] = r.Code
	}
	return out
}

// RemoveRole retira uma role do administrador. Ele precisa manter ao menos uma
// role, e o último Super Admin ativo não pode perder a role Super Admin.
func (s *Service) RemoveRole(ctx context.Context, actor Actor, id, roleID uuid.UUID) (*UserView, error) {
	var out *UserView
	err := s.withSuperGuard(ctx, func(q *db.Queries) error {
		if _, err := q.GetAdminByID(ctx, id); err != nil {
			return notFound(err)
		}
		role, err := q.GetAdminRole(ctx, roleID)
		if err != nil {
			return notFound(err)
		}
		if role.Code == SuperAdminCode && !actor.Subject.IsSuperAdmin() {
			return httpapi.ErrForbidden()
		}
		current, err := q.ListRolesForAdmin(ctx, id)
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(current, func(r db.AdminRole) bool { return r.ID == roleID }) {
			return httpapi.ErrNotFound()
		}
		if len(current) == 1 {
			return fieldIssue("roleId", "O administrador precisa ter ao menos uma role.")
		}
		if _, err := q.RemoveAdminRole(ctx, db.RemoveAdminRoleParams{AdminID: id, RoleID: roleID}); err != nil {
			return err
		}
		out, err = s.userView(ctx, q, id)
		if err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("admin.roles-remove", "admin", id,
			map[string]any{"roles": roleCodes(current)}, map[string]any{"roles": codesOfView(out.Roles)}))
	})
	return out, err
}
