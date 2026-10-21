package enduser

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

// Config reúne os parâmetros do serviço.
type Config struct {
	RefreshMaxTTL  time.Duration
	RefreshIdleTTL time.Duration
	ResetTTL       time.Duration
	ConfirmTTL     time.Duration
	PublicURL      string
	DefaultSender  string // remetente padrão dos templates (SMTP_FROM)
}

// Service implementa a regra de negócio de usuários finais.
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	hasher auth.Hasher
	jwt    *auth.JWTIssuer
	mail   mailer.Mailer
	engine *authz.Engine
	audit  audit.Recorder
	cfg    Config
	log    *slog.Logger
	dummy  string // hash descartável para igualar o tempo de login

	publicMu   sync.Mutex
	publicRole uuid.UUID
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, hasher auth.Hasher, jwt *auth.JWTIssuer, mail mailer.Mailer, engine *authz.Engine, cfg Config, log *slog.Logger) (*Service, error) {
	dummy, err := hasher.Hash("senha-descartavel")
	if err != nil {
		return nil, err
	}
	return &Service{pool: pool, q: db.New(pool), hasher: hasher, jwt: jwt, mail: mail, engine: engine, cfg: cfg, log: log, dummy: dummy}, nil
}

// EnsureDefaults cria, de forma idempotente, o que o sistema exige para funcionar:
// as roles Public e Authenticated, as configurações padrão e os templates de e-mail.
// Nunca sobrescreve alterações já feitas.
func (s *Service) EnsureDefaults(ctx context.Context) error {
	if _, err := s.q.UpsertEndUserRole(ctx, db.UpsertEndUserRoleParams{Type: RolePublic, Name: "Public", Description: "Aplicada a toda requisição sem token.", IsSystem: true}); err != nil {
		return err
	}
	authRole, err := s.q.UpsertEndUserRole(ctx, db.UpsertEndUserRoleParams{Type: RoleAuthenticated, Name: "Authenticated", Description: "Role padrão de novos usuários.", IsSystem: true})
	if err != nil {
		return err
	}
	if _, err := s.q.GetSetting(ctx, settingsKey); err != nil {
		st, _ := marshalSettings(defaultSettings(authRole.ID))
		if _, err := s.q.UpsertSetting(ctx, db.UpsertSettingParams{Key: settingsKey, Value: st}); err != nil {
			return err
		}
	}
	for _, t := range defaultTemplates(s.cfg.DefaultSender) {
		if err := s.q.InsertEmailTemplateIfMissing(ctx, db.InsertEmailTemplateIfMissingParams{
			Key: t.Key, SenderName: t.SenderName, SenderEmail: t.SenderEmail, ReplyTo: t.ReplyTo, Subject: t.Subject, BodyHtml: t.BodyHtml}); err != nil {
			return err
		}
	}
	return nil
}

// PublicRoleID devolve o ID da role Public (imutável, então fica em memória após a primeira leitura).
func (s *Service) PublicRoleID(ctx context.Context) (uuid.UUID, error) {
	s.publicMu.Lock()
	defer s.publicMu.Unlock()
	if s.publicRole != uuid.Nil {
		return s.publicRole, nil
	}
	r, err := s.q.GetEndUserRoleByType(ctx, RolePublic)
	if err != nil {
		return uuid.Nil, err
	}
	s.publicRole = r.ID
	return r.ID, nil
}

// AdminActor é o administrador que executa uma operação administrativa (para auditoria).
type AdminActor struct {
	ID uuid.UUID
	IP string
}

func (a AdminActor) entry(action, entity string, id uuid.UUID, before, after any) audit.Entry {
	return a.entryKey(action, entity, id.String(), before, after)
}

func (a AdminActor) entryKey(action, entity, entityID string, before, after any) audit.Entry {
	payload := map[string]any{}
	if before != nil {
		payload["before"] = before
	}
	if after != nil {
		payload["after"] = after
	}
	return audit.Entry{ActorType: "admin", ActorID: a.ID.String(), Action: action, Entity: entity, EntityID: entityID, Payload: payload, IP: a.IP}
}
