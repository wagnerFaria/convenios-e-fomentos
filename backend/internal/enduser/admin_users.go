package enduser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

func fieldIssue(field, msg string) error {
	return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: field, Message: msg}}})
}

// ListUsersFilter são os filtros da listagem administrativa.
type ListUsersFilter struct {
	Q         string
	RoleID    *uuid.UUID
	Confirmed *bool
	Blocked   *bool
	Sort      httpapi.Sort
	Page      httpapi.Page
}

// UserSortColumns mapeia campos de ordenação da API para colunas.
var UserSortColumns = map[string]string{"username": "u.username", "email": "u.email", "createdAt": "u.created_at"}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ListUsers lista usuários finais. A consulta é dinâmica, mas valores vão como
// parâmetros e a coluna de ordenação vem de uma lista permitida.
func (s *Service) ListUsers(ctx context.Context, f ListUsersFilter) ([]UserView, int, error) {
	var where []string
	var args []any
	arg := func(v any) int { args = append(args, v); return len(args) }
	if f.Q != "" {
		n := arg("%" + escapeLike(f.Q) + "%")
		where = append(where, fmt.Sprintf("(u.username ILIKE $%[1]d OR u.email::text ILIKE $%[1]d)", n))
	}
	if f.RoleID != nil {
		where = append(where, fmt.Sprintf("u.role_id = $%d", arg(*f.RoleID)))
	}
	if f.Confirmed != nil {
		where = append(where, fmt.Sprintf("u.confirmed = $%d", arg(*f.Confirmed)))
	}
	if f.Blocked != nil {
		where = append(where, fmt.Sprintf("u.blocked = $%d", arg(*f.Blocked)))
	}
	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}
	var total int
	if err := s.pool.QueryRow(ctx, "SELECT count(*) FROM end_users u"+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `SELECT u.id, u.username, u.email, u.confirmed, u.blocked, u.provider, u.profile, u.created_at, u.updated_at,
		r.id, r.type, r.name FROM end_users u JOIN end_user_roles r ON r.id = u.role_id` + clause +
		fmt.Sprintf(" ORDER BY %s, u.id LIMIT %d OFFSET %d", f.Sort.SQL(), f.Page.PageSize, f.Page.Offset())
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []UserView{}
	for rows.Next() {
		var v UserView
		var profile []byte
		if err := rows.Scan(&v.ID, &v.Username, &v.Email, &v.Confirmed, &v.Blocked, &v.Provider, &profile, &v.CreatedAt, &v.UpdatedAt, &v.Role.ID, &v.Role.Type, &v.Role.Name); err != nil {
			return nil, 0, err
		}
		v.Profile = map[string]any{}
		_ = json.Unmarshal(profile, &v.Profile)
		out = append(out, v)
	}
	return out, total, rows.Err()
}

// GetUser devolve um usuário final.
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*UserView, error) {
	return s.viewByID(ctx, s.q, id)
}

// CreateUserInput são os dados de criação feita por um administrador.
type CreateUserInput struct {
	Username, Email, Password string
	RoleID                    *uuid.UUID // nil = role padrão
	Confirmed, Blocked        bool
}

// CreateUser cria um usuário final (respeita "uma conta por e-mail").
func (s *Service) CreateUser(ctx context.Context, actor AdminActor, in CreateUserInput) (*UserView, error) {
	if err := s.checkPassword(in.Password); err != nil {
		return nil, err
	}
	hashed, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	var out *UserView
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		st, err := s.settings(ctx, q)
		if err != nil {
			return err
		}
		roleID := st.DefaultRoleID
		if in.RoleID != nil {
			roleID = *in.RoleID
		}
		if _, err := q.GetEndUserRole(ctx, roleID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fieldIssue("roleId", "Role inexistente.")
			}
			return err
		}
		email := strings.TrimSpace(in.Email)
		if st.OneAccountPerEmail {
			if err := q.LockEmail(ctx, strings.ToLower(email)); err != nil {
				return err
			}
			n, err := q.CountEndUsersByEmail(ctx, email)
			if err != nil {
				return err
			}
			if n > 0 {
				return errTaken
			}
		}
		u, err := q.CreateEndUser(ctx, db.CreateEndUserParams{Username: strings.TrimSpace(in.Username), Email: email, PasswordHash: &hashed,
			Confirmed: in.Confirmed, Blocked: in.Blocked, RoleID: roleID, Profile: []byte("{}")})
		if database.IsUniqueViolation(err) {
			return errTaken
		}
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("end-user.create", "end_user", u.ID, nil, snapshotOf(u))); err != nil {
			return err
		}
		out, err = s.viewByID(ctx, q, u.ID)
		return err
	})
	return out, err
}

// UpdateUserInput são os dados editáveis; Password opcional redefine a senha e revoga as sessões.
type UpdateUserInput struct {
	Username, Email string
	Password        *string
}

// UpdateUser edita o usuário final.
func (s *Service) UpdateUser(ctx context.Context, actor AdminActor, id uuid.UUID, in UpdateUserInput) (*UserView, error) {
	var hashed *string
	if in.Password != nil {
		if err := s.checkPassword(*in.Password); err != nil {
			return nil, err
		}
		h, err := s.hasher.Hash(*in.Password)
		if err != nil {
			return nil, err
		}
		hashed = &h
	}
	var out *UserView
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetEndUserByID(ctx, id)
		if err != nil {
			return notFound(err)
		}
		st, err := s.settings(ctx, q)
		if err != nil {
			return err
		}
		email := strings.TrimSpace(in.Email)
		if st.OneAccountPerEmail && !strings.EqualFold(email, before.Email) {
			if err := q.LockEmail(ctx, strings.ToLower(email)); err != nil {
				return err
			}
			n, err := q.CountEndUsersByEmail(ctx, email)
			if err != nil {
				return err
			}
			if n > 0 {
				return errTaken
			}
		}
		after, err := q.UpdateEndUser(ctx, db.UpdateEndUserParams{ID: id, Username: strings.TrimSpace(in.Username), Email: email})
		if database.IsUniqueViolation(err) {
			return errTaken
		}
		if err != nil {
			return err
		}
		if hashed != nil {
			if err := q.SetEndUserPassword(ctx, db.SetEndUserPasswordParams{ID: id, PasswordHash: hashed}); err != nil {
				return err
			}
			if err := q.RevokeAllEndUserSessions(ctx, id); err != nil {
				return err
			}
		}
		if err := s.audit.Record(ctx, q, actor.entry("end-user.update", "end_user", id, snapshotOf(before), snapshotOf(after))); err != nil {
			return err
		}
		out, err = s.viewByID(ctx, q, id)
		return err
	})
	return out, err
}

// DeleteUser exclui o usuário final (as sessões saem por cascata).
func (s *Service) DeleteUser(ctx context.Context, actor AdminActor, id uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetEndUserByID(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if _, err := q.DeleteEndUser(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("end-user.delete", "end_user", id, snapshotOf(before), nil))
	})
}

// mutate aplica uma alteração simples de estado com auditoria.
func (s *Service) mutate(ctx context.Context, actor AdminActor, id uuid.UUID, action string, fn func(q *db.Queries) error) (*UserView, error) {
	var out *UserView
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetEndUserByID(ctx, id)
		if err != nil {
			return notFound(err)
		}
		if err := fn(q); err != nil {
			return err
		}
		after, err := q.GetEndUserByID(ctx, id)
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry(action, "end_user", id, snapshotOf(before), snapshotOf(after))); err != nil {
			return err
		}
		out, err = s.viewByID(ctx, q, id)
		return err
	})
	return out, err
}

// SetBlocked bloqueia ou desbloqueia. Bloquear revoga as sessões (efeito imediato).
func (s *Service) SetBlocked(ctx context.Context, actor AdminActor, id uuid.UUID, blocked bool) (*UserView, error) {
	action := "end-user.unblock"
	if blocked {
		action = "end-user.block"
	}
	return s.mutate(ctx, actor, id, action, func(q *db.Queries) error {
		if err := q.SetEndUserBlocked(ctx, db.SetEndUserBlockedParams{ID: id, Blocked: blocked}); err != nil {
			return err
		}
		if blocked {
			return q.RevokeAllEndUserSessions(ctx, id)
		}
		return nil
	})
}

// SetRole altera a role do usuário (uma única role por usuário).
func (s *Service) SetRole(ctx context.Context, actor AdminActor, id, roleID uuid.UUID) (*UserView, error) {
	return s.mutate(ctx, actor, id, "end-user.role-change", func(q *db.Queries) error {
		if _, err := q.GetEndUserRole(ctx, roleID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return fieldIssue("roleId", "Role inexistente.")
			}
			return err
		}
		return q.SetEndUserRole(ctx, db.SetEndUserRoleParams{ID: id, RoleID: roleID})
	})
}

// Confirm marca o e-mail como confirmado.
func (s *Service) Confirm(ctx context.Context, actor AdminActor, id uuid.UUID) (*UserView, error) {
	return s.mutate(ctx, actor, id, "end-user.confirm", func(q *db.Queries) error {
		return q.SetEndUserConfirmed(ctx, db.SetEndUserConfirmedParams{ID: id, Confirmed: true})
	})
}
