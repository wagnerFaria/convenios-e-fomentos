package enduser

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/slug"
)

// RoleSummary é a role de usuário final exposta nas listagens.
type RoleSummary struct {
	ID          uuid.UUID `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsSystem    bool      `json:"isSystem"`
	UserCount   int64     `json:"userCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// PermissionView é uma ação liberada à role, com a rota vinculada.
type PermissionView struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Method string `json:"method"`
	Route  string `json:"route"`
}

// RoleDetail é a role com suas permissões.
type RoleDetail struct {
	RoleSummary
	Permissions []PermissionView `json:"permissions"`
}

func summary(r db.EndUserRole, count int64) RoleSummary {
	return RoleSummary{ID: r.ID, Type: r.Type, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem, UserCount: count, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}

func (s *Service) permissionViews(actions []string) []PermissionView {
	out := make([]PermissionView, 0, len(actions))
	for _, id := range actions {
		// Ações removidas do código somem da resposta (a sincronização apaga as linhas).
		if a, ok := s.engine.Actions().Get(id); ok && a.Domain == authz.DomainAPI {
			out = append(out, PermissionView{Action: a.ID, Name: a.Name, Method: a.Method, Route: a.Route})
		}
	}
	return out
}

// ListRoles lista as roles com a contagem de usuários.
func (s *Service) ListRoles(ctx context.Context) ([]RoleSummary, error) {
	rows, err := s.q.ListEndUserRolesWithCount(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]RoleSummary, len(rows))
	for i, r := range rows {
		out[i] = summary(db.EndUserRole{ID: r.ID, Type: r.Type, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, r.UserCount)
	}
	return out, nil
}

func (s *Service) roleDetail(ctx context.Context, q *db.Queries, id uuid.UUID) (*RoleDetail, error) {
	r, err := q.GetEndUserRoleWithCount(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	actions, err := q.ListEndUserPermissionActions(ctx, id)
	if err != nil {
		return nil, err
	}
	return &RoleDetail{
		RoleSummary: summary(db.EndUserRole{ID: r.ID, Type: r.Type, Name: r.Name, Description: r.Description, IsSystem: r.IsSystem, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}, r.UserCount),
		Permissions: s.permissionViews(actions),
	}, nil
}

// GetRole devolve a role com permissões (e a rota de cada ação).
func (s *Service) GetRole(ctx context.Context, id uuid.UUID) (*RoleDetail, error) {
	return s.roleDetail(ctx, s.q, id)
}

// validateActions confere se cada ação existe e é do domínio api.
func (s *Service) validateActions(actions []string) ([]string, error) {
	var issues []httpapi.FieldError
	for i, id := range actions {
		if a, ok := s.engine.Actions().Get(id); !ok || a.Domain != authz.DomainAPI {
			issues = append(issues, httpapi.FieldError{Field: fmt.Sprintf("permissions[%d]", i), Message: "Ação desconhecida ou fora do domínio api: " + id})
		}
	}
	if len(issues) > 0 {
		return nil, httpapi.ErrValidation(map[string]any{"fields": issues})
	}
	return slices.Compact(slices.Sorted(slices.Values(actions))), nil
}

func roleSnap(r db.EndUserRole, actions []string) map[string]any {
	return map[string]any{"type": r.Type, "name": r.Name, "description": r.Description, "permissions": actions}
}

// CreateRole cria uma role customizada com as ações informadas.
func (s *Service) CreateRole(ctx context.Context, actor AdminActor, name, description string, actions []string) (*RoleDetail, error) {
	actions, err := s.validateActions(actions)
	if err != nil {
		return nil, err
	}
	var out *RoleDetail
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		typ, err := uniqueType(ctx, q, name)
		if err != nil {
			return err
		}
		r, err := q.CreateEndUserRole(ctx, db.CreateEndUserRoleParams{Type: typ, Name: name, Description: description})
		if database.IsUniqueViolation(err) {
			return httpapi.ErrConflict("role_type_conflict", "Já existe uma role com este tipo. Tente novamente.")
		}
		if err != nil {
			return err
		}
		for _, a := range actions {
			if err := q.InsertEndUserPermission(ctx, db.InsertEndUserPermissionParams{RoleID: r.ID, Action: a}); err != nil {
				return err
			}
		}
		if err := s.audit.Record(ctx, q, actor.entry("end-user-role.create", "end_user_role", r.ID, nil, roleSnap(r, actions))); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, r.ID)
		return err
	})
	return out, err
}

func uniqueType(ctx context.Context, q *db.Queries, name string) (string, error) {
	base := customPrefix + slug.Make(name)
	for i := 1; i <= 100; i++ {
		t := base
		if i > 1 {
			t = fmt.Sprintf("%s-%d", base, i)
		}
		exists, err := q.EndUserRoleTypeExists(ctx, t)
		if err != nil {
			return "", err
		}
		if !exists {
			return t, nil
		}
	}
	return "", httpapi.ErrConflict("role_type_conflict", "Não foi possível gerar um tipo único para a role.")
}

// UpdateRole edita nome/descrição e, se actions != nil, substitui as permissões.
func (s *Service) UpdateRole(ctx context.Context, actor AdminActor, id uuid.UUID, name, description string, actions *[]string) (*RoleDetail, error) {
	var validated []string
	if actions != nil {
		v, err := s.validateActions(*actions)
		if err != nil {
			return nil, err
		}
		validated = v
	}
	var out *RoleDetail
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetEndUserRoleForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		beforeActions, err := q.ListEndUserPermissionActions(ctx, id)
		if err != nil {
			return err
		}
		after, err := q.UpdateEndUserRole(ctx, db.UpdateEndUserRoleParams{ID: id, Name: name, Description: description})
		if err != nil {
			return err
		}
		afterActions := beforeActions
		if actions != nil {
			if err := q.DeleteEndUserPermissionsByRole(ctx, id); err != nil {
				return err
			}
			for _, a := range validated {
				if err := q.InsertEndUserPermission(ctx, db.InsertEndUserPermissionParams{RoleID: id, Action: a}); err != nil {
					return err
				}
			}
			afterActions = validated
		}
		if err := s.audit.Record(ctx, q, actor.entry("end-user-role.update", "end_user_role", id, roleSnap(before, beforeActions), roleSnap(after, afterActions))); err != nil {
			return err
		}
		out, err = s.roleDetail(ctx, q, id)
		return err
	})
	if err != nil {
		return nil, err
	}
	if actions != nil {
		s.invalidate(ctx, id)
	}
	return out, nil
}

func (s *Service) invalidate(ctx context.Context, ids ...uuid.UUID) {
	if err := s.engine.Invalidate(ctx, ids...); err != nil {
		slog.ErrorContext(ctx, "falha ao propagar invalidação de permissões", "error", err)
	}
}

var (
	errRoleProtected = httpapi.ErrConflict("role_protected", "As roles Public e Authenticated fazem parte do sistema e não podem ser excluídas.")
	errRoleIsDefault = httpapi.ErrConflict("role_is_default", "Esta é a role padrão de novos usuários. Escolha outra nas configurações avançadas antes de excluí-la.")
)

// DeleteRole exclui a role customizada. Os usuários vinculados passam para a role
// Public na mesma transação. Public/Authenticated e a role padrão não podem ser excluídas.
func (s *Service) DeleteRole(ctx context.Context, actor AdminActor, id uuid.UUID) error {
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		role, err := q.GetEndUserRoleForUpdate(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if role.IsSystem {
			return errRoleProtected
		}
		st, err := s.settings(ctx, q)
		if err != nil {
			return err
		}
		if st.DefaultRoleID == id {
			return errRoleIsDefault
		}
		public, err := q.GetEndUserRoleByType(ctx, RolePublic)
		if err != nil {
			return err
		}
		actions, err := q.ListEndUserPermissionActions(ctx, id)
		if err != nil {
			return err
		}
		moved, err := q.MoveEndUsersToRole(ctx, db.MoveEndUsersToRoleParams{RoleID: id, RoleID_2: public.ID})
		if err != nil {
			return err
		}
		if _, err := q.DeleteEndUserRole(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("end-user-role.delete", "end_user_role", id, roleSnap(role, actions), map[string]any{"usersMovedToPublic": moved}))
	})
	if err != nil {
		return err
	}
	s.invalidate(ctx, id)
	return nil
}
