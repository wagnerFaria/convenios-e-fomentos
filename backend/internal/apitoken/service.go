// Pacote apitoken implementa os tokens de API: integrações sistema a sistema na
// API pública, sem usuário associado.
//
// O token tem o formato "gat_" + 43 caracteres (256 bits aleatórios) e é exibido
// uma única vez, ao criar ou regenerar. No banco fica apenas o HMAC-SHA256 dele
// (chave API_TOKEN_SALT).
package apitoken

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Prefix identifica tokens de API no cabeçalho Authorization.
const Prefix = "gat_"

const (
	randomBytes   = 32
	hintLen       = 4
	touchInterval = time.Minute
	touchTimeout  = 5 * time.Second
	maxPerms      = 500
)

// Durações aceitas na criação.
const (
	Lifespan7d        = "7d"
	Lifespan30d       = "30d"
	Lifespan90d       = "90d"
	LifespanUnlimited = "unlimited"
)

var lifespanDays = map[string]*int{
	Lifespan7d: ptr(7), Lifespan30d: ptr(30), Lifespan90d: ptr(90), LifespanUnlimited: nil,
}

func ptr(i int) *int { return &i }

var (
	errNameTaken = httpapi.ErrConflict("name_taken", "Já existe um token ativo com este nome.")
	errRevoked   = httpapi.ErrConflict("token_revoked", "O token foi revogado e não pode ser alterado nem regenerado.")
)

// Service implementa a regra de negócio dos tokens de API.
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	engine *authz.Engine
	audit  audit.Recorder
	salt   []byte
	log    *slog.Logger
	now    func() time.Time

	touchMu   sync.Mutex
	lastTouch map[uuid.UUID]time.Time
	wg        sync.WaitGroup
}

// NewService cria o serviço; salt é o API_TOKEN_SALT.
func NewService(pool *pgxpool.Pool, engine *authz.Engine, salt string, log *slog.Logger) *Service {
	return &Service{pool: pool, q: db.New(pool), engine: engine, salt: []byte(salt), log: log, now: time.Now, lastTouch: map[uuid.UUID]time.Time{}}
}

// ResetThrottleForTest limpa o controle em memória de last_used_at (somente para testes).
func (s *Service) ResetThrottleForTest() {
	s.touchMu.Lock()
	clear(s.lastTouch)
	s.touchMu.Unlock()
}

// Wait espera as atualizações assíncronas de last_used_at terminarem (encerramento e testes).
func (s *Service) Wait() { s.wg.Wait() }

// Hash calcula o HMAC-SHA256 (hex) do token com o salt da aplicação.
func (s *Service) Hash(raw string) string {
	m := hmac.New(sha256.New, s.salt)
	m.Write([]byte(raw))
	return hex.EncodeToString(m.Sum(nil))
}

func (s *Service) generate() (raw, hash, hint string, err error) {
	b := make([]byte, randomBytes)
	if _, err = rand.Read(b); err != nil {
		return "", "", "", fmt.Errorf("gerando token: %w", err)
	}
	raw = Prefix + base64.RawURLEncoding.EncodeToString(b)
	return raw, s.Hash(raw), raw[len(raw)-hintLen:], nil
}

// Status do token.
const (
	StatusActive  = "active"
	StatusExpired = "expired"
	StatusRevoked = "revoked"
)

// View é o token exposto pela API. Nunca inclui o valor nem o hash.
type View struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Permissions []string   `json:"permissions"`
	Hint        string     `json:"hint"` // últimos caracteres, ex.: "gat_…a1b2"
	Lifespan    string     `json:"lifespan"`
	Status      string     `json:"status"`
	ExpiresAt   *time.Time `json:"expiresAt"`
	LastUsedAt  *time.Time `json:"lastUsedAt"`
	CreatedBy   *uuid.UUID `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	RevokedAt   *time.Time `json:"revokedAt"`
}

// Created é o resultado de criar/regenerar: o único momento em que o valor aparece.
type Created struct {
	View
	Token string `json:"token"`
}

func (s *Service) view(t db.ApiToken) View {
	perms := []string{}
	_ = json.Unmarshal(t.Permissions, &perms)
	lifespan := LifespanUnlimited
	if t.LifespanDays != nil {
		lifespan = fmt.Sprintf("%dd", *t.LifespanDays)
	}
	status := StatusActive
	switch {
	case t.RevokedAt != nil:
		status = StatusRevoked
	case t.ExpiresAt != nil && !t.ExpiresAt.After(s.now()):
		status = StatusExpired
	}
	return View{ID: t.ID, Name: t.Name, Description: t.Description, Type: t.Type, Permissions: perms, Hint: Prefix + "…" + t.Hint,
		Lifespan: lifespan, Status: status, ExpiresAt: t.ExpiresAt, LastUsedAt: t.LastUsedAt, CreatedBy: t.CreatedBy,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt, RevokedAt: t.RevokedAt}
}

// AdminActor é o administrador que executa a operação (auditoria e created_by).
type AdminActor struct {
	ID uuid.UUID
	IP string
}

func (a AdminActor) entry(action string, id uuid.UUID, before, after any) audit.Entry {
	payload := map[string]any{}
	if before != nil {
		payload["before"] = before
	}
	if after != nil {
		payload["after"] = after
	}
	return audit.Entry{ActorType: "admin", ActorID: a.ID.String(), Action: action, Entity: "api_token", EntityID: id.String(), Payload: payload, IP: a.IP}
}

// snapshot é o estado gravado na auditoria; jamais contém o token ou o hash.
func snapshot(v View) map[string]any {
	return map[string]any{"name": v.Name, "description": v.Description, "type": v.Type, "permissions": v.Permissions,
		"lifespan": v.Lifespan, "expiresAt": v.ExpiresAt, "status": v.Status}
}

// Input são os dados editáveis de um token.
type Input struct {
	Name        string
	Description string
	Type        string
	Permissions []string
}

// normalize valida o tipo e as ações. Só o tipo custom aceita (e exige) ações api::.
func (s *Service) normalize(in Input) ([]string, error) {
	var issues []httpapi.FieldError
	switch in.Type {
	case authz.TokenReadOnly, authz.TokenFullAccess:
		if len(in.Permissions) > 0 {
			issues = append(issues, httpapi.FieldError{Field: "permissions", Message: "Somente tokens do tipo custom aceitam lista de ações."})
		}
		return []string{}, validation(issues)
	case authz.TokenCustom:
	default:
		return nil, validation([]httpapi.FieldError{{Field: "type", Message: "Use read-only, full-access ou custom."}})
	}
	if len(in.Permissions) == 0 {
		issues = append(issues, httpapi.FieldError{Field: "permissions", Message: "Informe ao menos uma ação para tokens custom."})
	}
	if len(in.Permissions) > maxPerms {
		issues = append(issues, httpapi.FieldError{Field: "permissions", Message: "Ações demais."})
	}
	for i, id := range in.Permissions {
		if a, ok := s.engine.Actions().Get(id); !ok || a.Domain != authz.DomainAPI {
			issues = append(issues, httpapi.FieldError{Field: fmt.Sprintf("permissions[%d]", i), Message: "Ação desconhecida ou fora do domínio api: " + id})
		}
	}
	if len(issues) > 0 {
		return nil, validation(issues)
	}
	return slices.Compact(slices.Sorted(slices.Values(in.Permissions))), nil
}

func validation(issues []httpapi.FieldError) error {
	if len(issues) == 0 {
		return nil
	}
	return httpapi.ErrValidation(map[string]any{"fields": issues})
}

// Create gera um token novo. O valor em claro só é devolvido aqui.
func (s *Service) Create(ctx context.Context, actor AdminActor, in Input, lifespan string) (*Created, error) {
	days, ok := lifespanDays[lifespan]
	if !ok {
		return nil, validation([]httpapi.FieldError{{Field: "lifespan", Message: "Use 7d, 30d, 90d ou unlimited."}})
	}
	perms, err := s.normalize(in)
	if err != nil {
		return nil, err
	}
	permsJSON, _ := json.Marshal(perms)
	raw, hash, hint, err := s.generate()
	if err != nil {
		return nil, err
	}
	var expires *time.Time
	if days != nil {
		t := s.now().Add(time.Duration(*days) * 24 * time.Hour)
		expires = &t
	}
	var out *Created
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		t, err := q.CreateAPIToken(ctx, db.CreateAPITokenParams{Name: in.Name, Description: in.Description, Type: in.Type, TokenHash: hash, Hint: hint,
			Permissions: permsJSON, LifespanDays: intPtr32(days), ExpiresAt: expires, CreatedBy: &actor.ID})
		if database.IsUniqueViolation(err) {
			return errNameTaken
		}
		if err != nil {
			return err
		}
		v := s.view(t)
		out = &Created{View: v, Token: raw}
		return s.audit.Record(ctx, q, actor.entry("api-token.create", t.ID, nil, snapshot(v)))
	})
	return out, err
}

func intPtr32(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p) //nolint:gosec // 7, 30 ou 90
	return &v
}

// List lista os tokens (sem os revogados, salvo includeRevoked).
func (s *Service) List(ctx context.Context, includeRevoked bool) ([]View, error) {
	rows, err := s.q.ListAPITokens(ctx, includeRevoked)
	if err != nil {
		return nil, err
	}
	out := make([]View, len(rows))
	for i, r := range rows {
		out[i] = s.view(r)
	}
	return out, nil
}

// Get devolve um token.
func (s *Service) Get(ctx context.Context, id uuid.UUID) (*View, error) {
	t, err := s.q.GetAPIToken(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrNotFound()
	}
	if err != nil {
		return nil, err
	}
	v := s.view(t)
	return &v, nil
}

// Update edita nome, descrição, tipo e ações. A mudança vale na próxima requisição.
func (s *Service) Update(ctx context.Context, actor AdminActor, id uuid.UUID, in Input) (*View, error) {
	perms, err := s.normalize(in)
	if err != nil {
		return nil, err
	}
	permsJSON, _ := json.Marshal(perms)
	var out *View
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetAPITokenForUpdate(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.ErrNotFound()
		}
		if err != nil {
			return err
		}
		if before.RevokedAt != nil {
			return errRevoked
		}
		after, err := q.UpdateAPIToken(ctx, db.UpdateAPITokenParams{ID: id, Name: in.Name, Description: in.Description, Type: in.Type, Permissions: permsJSON})
		if database.IsUniqueViolation(err) {
			return errNameTaken
		}
		if err != nil {
			return err
		}
		v := s.view(after)
		out = &v
		return s.audit.Record(ctx, q, actor.entry("api-token.update", id, snapshot(s.view(before)), snapshot(v)))
	})
	return out, err
}

// Regenerate troca o valor do token (o anterior deixa de valer na hora) e reinicia a
// validade conforme a duração original.
func (s *Service) Regenerate(ctx context.Context, actor AdminActor, id uuid.UUID) (*Created, error) {
	raw, hash, hint, err := s.generate()
	if err != nil {
		return nil, err
	}
	var out *Created
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetAPITokenForUpdate(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.ErrNotFound()
		}
		if err != nil {
			return err
		}
		if before.RevokedAt != nil {
			return errRevoked
		}
		var expires *time.Time
		if before.LifespanDays != nil {
			t := s.now().Add(time.Duration(*before.LifespanDays) * 24 * time.Hour)
			expires = &t
		}
		after, err := q.RegenerateAPIToken(ctx, db.RegenerateAPITokenParams{ID: id, TokenHash: hash, Hint: hint, ExpiresAt: expires})
		if err != nil {
			return err
		}
		v := s.view(after)
		out = &Created{View: v, Token: raw}
		return s.audit.Record(ctx, q, actor.entry("api-token.regenerate", id, snapshot(s.view(before)), snapshot(v)))
	})
	return out, err
}

// Revoke invalida o token definitivamente (efeito imediato).
func (s *Service) Revoke(ctx context.Context, actor AdminActor, id uuid.UUID) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		before, err := q.GetAPITokenForUpdate(ctx, id)
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.ErrNotFound()
		}
		if err != nil {
			return err
		}
		if before.RevokedAt != nil {
			return httpapi.ErrNotFound()
		}
		if _, err := q.RevokeAPIToken(ctx, id); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("api-token.revoke", id, snapshot(s.view(before)), map[string]any{"status": StatusRevoked}))
	})
}

// Authenticate valida o token recebido na API pública e devolve o sujeito. Revogação,
// expiração e mudanças de tipo/ações valem na hora, pois cada requisição consulta o banco.
func (s *Service) Authenticate(ctx context.Context, raw string) (authz.Subject, error) {
	// Formato: prefixo + 43 caracteres base64url. Evita consulta ao banco para lixo óbvio.
	if !strings.HasPrefix(raw, Prefix) || len(raw) != len(Prefix)+base64.RawURLEncoding.EncodedLen(randomBytes) {
		return authz.Subject{}, httpapi.ErrUnauthorized()
	}
	t, err := s.q.GetActiveAPITokenByHash(ctx, s.Hash(raw))
	if errors.Is(err, pgx.ErrNoRows) {
		return authz.Subject{}, httpapi.ErrUnauthorized()
	}
	if err != nil {
		return authz.Subject{}, err
	}
	perms := []string{}
	if err := json.Unmarshal(t.Permissions, &perms); err != nil {
		return authz.Subject{}, fmt.Errorf("permissões do token %s: %w", t.ID, err)
	}
	s.touch(t.ID)
	return authz.Subject{Type: authz.SubjectAPIToken, ID: t.ID, APIToken: &authz.TokenGrant{Type: t.Type, Actions: perms}}, nil
}

// touch atualiza last_used_at fora do caminho da requisição e no máximo uma vez por minuto por token.
func (s *Service) touch(id uuid.UUID) {
	now := s.now()
	s.touchMu.Lock()
	if last, ok := s.lastTouch[id]; ok && now.Sub(last) < touchInterval {
		s.touchMu.Unlock()
		return
	}
	s.lastTouch[id] = now
	s.touchMu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), touchTimeout)
		defer cancel()
		if err := s.q.TouchAPIToken(ctx, id); err != nil {
			s.log.Warn("falha ao atualizar last_used_at do token de API", "token_id", id, "error", err)
		}
	}()
}
