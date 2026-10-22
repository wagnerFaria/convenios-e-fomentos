package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/slug"
)

// RoleSummary é a role exposta nas listagens.
type RoleSummary struct {
	ID           uuid.UUID `json:"id"`
	Code         string    `json:"code"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	IsSystem     bool      `json:"isSystem"`
	IsSuperAdmin bool      `json:"isSuperAdmin"` // acesso total; não tem permissões editáveis
	UserCount    int64     `json:"userCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// PermissionView é uma permissão de role exposta pela API. Fields nulo = todos os campos.
type PermissionView struct {
	Action     string         `json:"action"`
	Subject    string         `json:"subject,omitempty"`
	Fields     []string       `json:"fields"`
	Conditions []string       `json:"conditions"`
	Properties map[string]any `json:"properties"`
}

// RoleDetail é a role com suas permissões.
type RoleDetail struct {
	RoleSummary
	Permissions []PermissionView `json:"permissions"`
}

func summaryOf(r db.AdminRole, count int64) RoleSummary {
	return RoleSummary{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem,
		IsSuperAdmin: r.Code == SuperAdminCode, UserCount: count, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func permViews(ps []authz.Permission) []PermissionView {
	out := make([]PermissionView, 0, len(ps))
	for _, p := range ps {
		c := p.Conditions
		if c == nil {
			c = []string{}
		}
		props := p.Properties
		if props == nil {
			props = map[string]any{}
		}
		out = append(out, PermissionView{Action: p.Action, Subject: p.Subject, Fields: p.Fields, Conditions: c, Properties: props})
	}
	return out
}

// ListRoles lista as roles com a contagem de administradores.
func (s *Service) ListRoles(ctx context.Context) ([]RoleSummary, error) {
	rows, err := s.q.ListAdminRolesWithCount(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, summaryOf(db.AdminRole{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, r.UserCount))
	}
	return out, nil
}

func (s *Service) roleDetail(ctx context.Context, q *db.Queries, id uuid.UUID) (*RoleDetail, error) {
	r, err := q.GetAdminRoleWithCount(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	perms, err := authz.DBStore{Q: q}.RolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleDetail{
		RoleSummary: summaryOf(db.AdminRole{ID: r.ID, Code: r.Code, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, r.UserCount),
		Permissions: permViews(perms),
	}, nil
}

// GetRole devolve a role com permissões e contagem de administradores.
func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (*RoleDetail, error) {
	return s.roleDetail(ctx, s.q, id)
}

// uniqueRoleCode gera um code único a partir do nome (sufixo -2, -3... se preciso).
func uniqueRoleCode(ctx context.Context, q *db.Queries, name string) (string, error) {
	base := slug.Make(name)
	for i := 1; i <= 100; i++ {
		code := base
		if i > 1 {
			code = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := q.AdminRoleCodeExists(ctx, code)
		if err != nil {
			return "", err
		}
		if !exists {
			return code, nil
		}
	}
	return "", httpapi.ErrConflict("role_code_conflict", "Não foi possível gerar um código único para a role.")
}

func roleSnapshot(r db.AdminRole) map[string]any {
	return map[string]any{"code": r.Code, "name": r.Name, "description": r.Description}
}

// CreateRole cria uma role sem permissões.
func (s *Service) CreateRole(ctx context.Context, actor Actor, name, description string) (*RoleDetail, error) {
	var out *RoleDetail
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		code, err := uniqueRoleCode(ctx, q, name)
		if err != nil {
			return err
		}
		r, err := q.CreateAdminRole(ctx, db.CreateAdminRoleParams{Code: code, Name: name, Description: description})
		if database.IsUniqueViolation(err) {
			return httpapi.ErrConflict("role_code_conflict", "Já existe uma role com este código. Tente novamente.")
		}
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("admin-role.create", "admin_role", r.ID, nil, roleSnapshot(r))); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, r.ID)
		return err
	})
	return out, err
}

// UpdateRole edita nome e descrição (o code é imutável).
func (s *Service) UpdateRole(ctx context.Context, actor Actor, id uuid.UUID, name, description string) (*RoleDetail, error) {
	var out *RoleDetail
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetAdminRoleForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		after, err := q.UpdateAdminRole(ctx, db.UpdateAdminRoleParams{ID: id, Name: name, Description: description})
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("admin-role.update", "admin_role", id, roleSnapshot(before), roleSnapshot(after))); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, id)
		return err
	})
	return out, err
}

func insertPermission(ctx context.Context, q *db.Queries, roleID uuid.UUID, p authz.Permission) error {
	var fields []byte
	if p.Fields != nil {
		b, err := json.Marshal(p.Fields)
		if err != nil {
			return err
		}
		fields = b
	}
	conds, err := json.Marshal(append([]string{}, p.Conditions...))
	if err != nil {
		return err
	}
	props := p.Properties
	if props == nil {
		props = map[string]any{}
	}
	propsJSON, err := json.Marshal(props)
	if err != nil {
		return err
	}
	var subject *string
	if p.Subject != "" {
		subject = &p.Subject
	}
	_, err = q.CreateAdminPermission(ctx, db.CreateAdminPermissionParams{RoleID: roleID, Action: p.Action, Subject: subject, Fields: fields, Conditions: conds, Properties: propsJSON})
	return err
}

// ReplacePermissions substitui, em uma transação, todas as permissões da role.
// A role Super Admin não pode ser alterada. O cache é invalidado após o commit.
func (s *Service) ReplacePermissions(ctx context.Context, actor Actor, id uuid.UUID, in []authz.Permission) (*RoleDetail, error) {
	role, err := s.q.GetAdminRole(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	if role.Code == SuperAdminCode {
		return nil, errSuperAdminImmutable
	}
	if !actor.Subject.IsSuperAdmin() && ownsRole(actor, id) {
		return nil, errCannotModifyOwnRole
	}
	perms, issues := s.engine.NormalizePermissions(in)
	if len(issues) > 0 {
		fields := make([]httpapi.FieldError, len(issues))
		for i, is := range issues {
			fields[i] = httpapi.FieldError{Field: is.Path, Message: is.Message}
		}
		return nil, httpapi.ErrValidation(map[string]any{"fields": fields})
	}

	var out *RoleDetail
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := q.GetAdminRoleForUpdate(ctx, id); err != nil {
			return notFound(err)
		}
		before, err := authz.DBStore{Q: q}.RolePermissions(ctx, id)
		if err != nil {
			return err
		}
		// Só se concede ou amplia o que o próprio ator possui; reduzir é sempre permitido.
		if err := s.ensureCovers(ctx, actor, s.engine.Widened(before, perms)); err != nil {
			return err
		}
		if err := q.DeleteAdminPermissionsByRole(ctx, id); err != nil {
			return err
		}
		for _, p := range perms {
			if err := insertPermission(ctx, q, id, p); err != nil {
				return err
			}
		}
		if err := s.audit.Record(ctx, q, actor.entry("admin-role.permissions-replace", "admin_role", id,
			map[string]any{"permissions": permViews(before)}, map[string]any{"permissions": permViews(perms)})); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	s.invalidate(ctx, id)
	return out, nil
}

// invalidate limpa o cache de permissões; a falha de propagação para outras
// instâncias não desfaz a alteração, mas é registrada.
func (s *Service) invalidate(ctx context.Context, ids ...uuid.UUID) {
	if err := s.engine.Invalidate(ctx, ids...); err != nil {
		slog.ErrorContext(ctx, "falha ao propagar invalidação de permissões", "error", err)
	}
}

// DuplicateRole copia nome (com sufixo), descrição e permissões.
func (s *Service) DuplicateRole(ctx context.Context, actor Actor, id uuid.UUID) (*RoleDetail, error) {
	var out *RoleDetail
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		src, err := q.GetAdminRole(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if src.Code == SuperAdminCode {
			return errSuperAdminImmutable
		}
		perms, err := authz.DBStore{Q: q}.RolePermissions(ctx, id)
		if err != nil {
			return err
		}
		if err := s.ensureCovers(ctx, actor, perms); err != nil {
			return err
		}
		name := src.Name + " (cópia)"
		code, err := uniqueRoleCode(ctx, q, name)
		if err != nil {
			return err
		}
		dup, err := q.CreateAdminRole(ctx, db.CreateAdminRoleParams{Code: code, Name: name, Description: src.Description})
		if database.IsUniqueViolation(err) {
			return httpapi.ErrConflict("role_code_conflict", "Já existe uma role com este código. Tente novamente.")
		}
		if err != nil {
			return err
		}
		for _, p := range perms {
			if err := insertPermission(ctx, q, dup.ID, p); err != nil {
				return err
			}
		}
		if err := s.audit.Record(ctx, q, actor.entry("admin-role.duplicate", "admin_role", dup.ID,
			map[string]any{"sourceRoleId": src.ID}, map[string]any{"role": roleSnapshot(dup), "permissions": permViews(perms)})); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, dup.ID)
		return err
	})
	return out, err
}

// DeleteRole exclui a role. Não é permitido para o Super Admin nem para roles
// atribuídas a algum administrador (409).
func (s *Service) DeleteRole(ctx context.Context, actor Actor, id uuid.UUID) error {
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		role, err := q.GetAdminRoleForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if role.Code == SuperAdminCode {
			return errSuperAdminImmutable
		}
		n, err := q.CountAdminsWithRole(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return errRoleInUse(n)
		}
		perms, err := authz.DBStore{Q: q}.RolePermissions(ctx, id)
		if err != nil {
			return err
		}
		if _, err := q.DeleteAdminRole(ctx, id); err != nil {
			// Atribuição concorrente após a contagem: a FK RESTRICT protege.
			if database.IsForeignKeyViolation(err) {
				return errRoleInUse(1)
			}
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("admin-role.delete", "admin_role", id,
			map[string]any{"role": roleSnapshot(role), "permissions": permViews(perms)}, nil))
	})
	if err != nil {
		return err
	}
	s.invalidate(ctx, id)
	return nil
}
