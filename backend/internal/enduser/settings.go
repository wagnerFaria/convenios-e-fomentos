// Pacote enduser implementa usuários finais (Users & Permissions): cadastro e
// autenticação na API pública, roles Public/Authenticated/customizadas,
// configurações avançadas, templates de e-mail e a administração de tudo isso.
package enduser

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

const settingsKey = "enduser.advanced"

// Tipos das roles padrão.
const (
	RolePublic        = "public"
	RoleAuthenticated = "authenticated"
	customPrefix      = "custom-"
)

// Settings são as configurações avançadas de usuários finais.
type Settings struct {
	DefaultRoleID                uuid.UUID `json:"defaultRoleId"`
	OneAccountPerEmail           bool      `json:"oneAccountPerEmail"`
	RegistrationEnabled          bool      `json:"registrationEnabled"`
	EmailConfirmationRequired    bool      `json:"emailConfirmationRequired"`
	ResetPasswordURL             string    `json:"resetPasswordUrl"`
	EmailConfirmationRedirectURL string    `json:"emailConfirmationRedirectUrl"`
	// Campos extras aceitos no cadastro (equivalente ao register.allowedFields do Strapi); ficam em profile.
	RegisterAllowedFields []string `json:"registerAllowedFields"`
}

func defaultSettings(authenticated uuid.UUID) Settings {
	return Settings{DefaultRoleID: authenticated, OneAccountPerEmail: true, RegistrationEnabled: true, RegisterAllowedFields: []string{}}
}

// Campos que o cadastro já trata ou que nunca podem ser definidos pelo cliente.
var reservedFields = []string{"id", "username", "email", "password", "passwordHash", "password_hash", "confirmed", "blocked", "role", "roleId", "role_id", "provider", "profile", "createdAt", "updatedAt", "created_at", "updated_at"}

var fieldNameRe = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]{0,49}$`)

const maxAllowedFields = 20

// GetSettings lê as configurações; sem registro usa os padrões.
func (s *Service) GetSettings(ctx context.Context) (Settings, error) { return s.settings(ctx, s.q) }

func (s *Service) settings(ctx context.Context, q *db.Queries) (Settings, error) {
	var st Settings
	row, err := q.GetSetting(ctx, settingsKey)
	if err == nil {
		if err := json.Unmarshal(row.Value, &st); err != nil {
			return Settings{}, err
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return Settings{}, err
	} else {
		st = defaultSettings(uuid.Nil)
	}
	if st.DefaultRoleID == uuid.Nil {
		auth, err := q.GetEndUserRoleByType(ctx, RoleAuthenticated)
		if err != nil {
			return Settings{}, err
		}
		st.DefaultRoleID = auth.ID
	}
	if st.RegisterAllowedFields == nil {
		st.RegisterAllowedFields = []string{}
	}
	return st, nil
}

func validURL(raw string) bool {
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// UpdateSettings valida e grava as configurações, com auditoria.
func (s *Service) UpdateSettings(ctx context.Context, actor AdminActor, in Settings) (Settings, error) {
	var issues []httpapi.FieldError
	bad := func(field, msg string) { issues = append(issues, httpapi.FieldError{Field: field, Message: msg}) }
	if !validURL(in.ResetPasswordURL) {
		bad("resetPasswordUrl", "Informe uma URL http(s) absoluta.")
	}
	if !validURL(in.EmailConfirmationRedirectURL) {
		bad("emailConfirmationRedirectUrl", "Informe uma URL http(s) absoluta.")
	}
	if len(in.RegisterAllowedFields) > maxAllowedFields {
		bad("registerAllowedFields", "Máximo de 20 campos.")
	}
	for _, f := range in.RegisterAllowedFields {
		if !fieldNameRe.MatchString(f) || slices.Contains(reservedFields, f) {
			bad("registerAllowedFields", "Nome de campo inválido ou reservado: "+f)
		}
	}
	role, err := s.q.GetEndUserRole(ctx, in.DefaultRoleID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		bad("defaultRoleId", "Role inexistente.")
	case err != nil:
		return Settings{}, err
	case role.Type == RolePublic:
		bad("defaultRoleId", "A role Public não pode ser a padrão de novos usuários.")
	}
	if len(issues) > 0 {
		return Settings{}, httpapi.ErrValidation(map[string]any{"fields": issues})
	}
	in.RegisterAllowedFields = slices.Compact(slices.Sorted(slices.Values(in.RegisterAllowedFields)))
	if in.RegisterAllowedFields == nil {
		in.RegisterAllowedFields = []string{}
	}

	before, err := s.GetSettings(ctx)
	if err != nil {
		return Settings{}, err
	}
	value, err := json.Marshal(in)
	if err != nil {
		return Settings{}, err
	}
	if _, err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: settingsKey, Value: value}); err != nil {
		return Settings{}, err
	}
	if err := s.audit.Record(ctx, s.q, actor.entry("end-user-settings.update", "settings", uuid.Nil, before, in)); err != nil {
		return Settings{}, err
	}
	return in, nil
}

func marshalSettings(st Settings) ([]byte, error) { return json.Marshal(st) }
