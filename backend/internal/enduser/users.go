package enduser

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

var (
	errInvalidCredentials   = &httpapi.APIError{Status: 401, Code: "invalid_credentials", Message: "Identificador ou senha inválidos."}
	errInvalidRefresh       = &httpapi.APIError{Status: 401, Code: "invalid_refresh_token", Message: "Sessão inválida ou expirada. Faça login novamente."}
	errInvalidOneTime       = &httpapi.APIError{Status: 400, Code: "invalid_token", Message: "Token inválido, expirado ou já utilizado."}
	errTaken                = &httpapi.APIError{Status: 409, Code: "already_taken", Message: "E-mail ou nome de usuário já em uso."}
	errRegistrationDisabled = &httpapi.APIError{Status: 403, Code: "registration_disabled", Message: "O cadastro de novos usuários está desabilitado."}
	errBlocked              = &httpapi.APIError{Status: 403, Code: "account_blocked", Message: "Esta conta está bloqueada."}
	errNotConfirmed         = &httpapi.APIError{Status: 403, Code: "email_not_confirmed", Message: "Confirme seu e-mail para entrar."}
)

// RoleRef é a role resumida de um usuário.
type RoleRef struct {
	ID   uuid.UUID `json:"id"`
	Type string    `json:"type"`
	Name string    `json:"name"`
}

// UserView é o usuário final exposto pela API. Nunca inclui o hash da senha.
type UserView struct {
	ID        uuid.UUID      `json:"id"`
	Username  string         `json:"username"`
	Email     string         `json:"email"`
	Confirmed bool           `json:"confirmed"`
	Blocked   bool           `json:"blocked"`
	Provider  string         `json:"provider"`
	Role      RoleRef        `json:"role"`
	Profile   map[string]any `json:"profile"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func viewOf(u db.EndUser, role db.EndUserRole) UserView {
	profile := map[string]any{}
	_ = json.Unmarshal(u.Profile, &profile)
	return UserView{ID: u.ID, Username: u.Username, Email: u.Email, Confirmed: u.Confirmed, Blocked: u.Blocked, Provider: u.Provider,
		Role: RoleRef{ID: role.ID, Type: role.Type, Name: role.Name}, Profile: profile, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

// userSnapshot é o estado de um usuário gravado na auditoria (sem senha).
type userSnapshot struct {
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Confirmed bool      `json:"confirmed"`
	Blocked   bool      `json:"blocked"`
	RoleID    uuid.UUID `json:"roleId"`
}

func snapshotOf(u db.EndUser) userSnapshot {
	return userSnapshot{Username: u.Username, Email: u.Email, Confirmed: u.Confirmed, Blocked: u.Blocked, RoleID: u.RoleID}
}

func (s *Service) viewByID(ctx context.Context, q *db.Queries, id uuid.UUID) (*UserView, error) {
	u, err := q.GetEndUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrNotFound()
	}
	if err != nil {
		return nil, err
	}
	role, err := q.GetEndUserRole(ctx, u.RoleID)
	if err != nil {
		return nil, err
	}
	v := viewOf(u, role)
	return &v, nil
}

// notFound converte "sem linhas" em 404.
func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.ErrNotFound()
	}
	return err
}
