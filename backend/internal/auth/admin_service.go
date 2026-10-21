package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

// Erros voltados ao cliente. As mensagens de login são genéricas de propósito:
// não revelam se o e-mail existe.
var (
	errInvalidCredentials = &httpapi.APIError{Status: 401, Code: "invalid_credentials", Message: "E-mail ou senha inválidos."}
	errInvalidRefresh     = &httpapi.APIError{Status: 401, Code: "invalid_refresh_token", Message: "Sessão inválida ou expirada. Faça login novamente."}
	errInvalidOneTime     = &httpapi.APIError{Status: 400, Code: "invalid_token", Message: "Token inválido, expirado ou já utilizado."}
)

const (
	subjectAdmin    = "admin"
	purposeInvite   = "invite"
	purposeReset    = "reset"
	maxUserAgent    = 255
	tokenTypeBearer = "Bearer"
)

// AdminAuthConfig reúne os parâmetros do serviço.
type AdminAuthConfig struct {
	RefreshMaxTTL  time.Duration
	RefreshIdleTTL time.Duration
	ResetTokenTTL  time.Duration
	PublicURL      string
}

// AdminAuth implementa login, sessões, convite aceito e redefinição de senha dos administradores.
type AdminAuth struct {
	pool    *pgxpool.Pool
	q       *db.Queries
	hasher  Hasher
	jwt     *JWTIssuer
	mail    mailer.Mailer
	audit   audit.Recorder
	cfg     AdminAuthConfig
	log     *slog.Logger
	dummy   string // hash descartável para igualar o tempo de login de e-mails inexistentes
	nowFunc func() time.Time
}

// NewAdminAuth cria o serviço.
func NewAdminAuth(pool *pgxpool.Pool, hasher Hasher, jwt *JWTIssuer, mail mailer.Mailer, cfg AdminAuthConfig, log *slog.Logger) (*AdminAuth, error) {
	dummy, err := hasher.Hash("senha-descartavel")
	if err != nil {
		return nil, err
	}
	return &AdminAuth{pool: pool, q: db.New(pool), hasher: hasher, jwt: jwt, mail: mail, cfg: cfg, log: log, dummy: dummy, nowFunc: time.Now}, nil
}

// TokenPair é o resultado de login/refresh.
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int    `json:"expiresIn"` // segundos
}

// SessionMeta descreve o cliente que abre a sessão.
type SessionMeta struct{ UserAgent, IP string }

// Trimmed limita o tamanho do user agent gravado.
func (m SessionMeta) Trimmed() SessionMeta {
	if len(m.UserAgent) > maxUserAgent {
		m.UserAgent = m.UserAgent[:maxUserAgent]
	}
	return m
}

// Login autentica por e-mail e senha e abre uma nova família de sessão.
func (s *AdminAuth) Login(ctx context.Context, email, password string, meta SessionMeta) (*TokenPair, error) {
	admin, err := s.q.GetAdminByEmail(ctx, strings.TrimSpace(email))
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	hash := s.dummy
	if err == nil && admin.PasswordHash != nil {
		hash = *admin.PasswordHash
	}
	// A verificação sempre roda, para que o tempo não revele a existência do e-mail.
	ok := s.hasher.Verify(password, hash)
	if err != nil || admin.PasswordHash == nil || !ok || !admin.IsActive || admin.Blocked {
		return nil, errInvalidCredentials
	}
	now := s.nowFunc()
	return s.createSession(ctx, s.q, admin.ID, uuid.New(), now, meta)
}

// createSession grava um novo refresh token na família e emite o par de tokens.
func (s *AdminAuth) createSession(ctx context.Context, q *db.Queries, adminID, family uuid.UUID, familyStart time.Time, meta SessionMeta) (*TokenPair, error) {
	now := s.nowFunc()
	expires := now.Add(s.cfg.RefreshIdleTTL)
	if abs := familyStart.Add(s.cfg.RefreshMaxTTL); abs.Before(expires) {
		expires = abs
	}
	if !expires.After(now) {
		return nil, errInvalidRefresh
	}
	meta = meta.Trimmed()
	plain, hash, err := NewOpaqueToken()
	if err != nil {
		return nil, err
	}
	if _, err = q.CreateAdminSession(ctx, db.CreateAdminSessionParams{
		AdminID: adminID, FamilyID: family, FamilyStartedAt: familyStart,
		RefreshTokenHash: hash, UserAgent: meta.UserAgent, Ip: meta.IP, ExpiresAt: expires,
	}); err != nil {
		return nil, fmt.Errorf("criando sessão: %w", err)
	}
	access, err := s.jwt.Issue(adminID, family)
	if err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: plain, TokenType: tokenTypeBearer, ExpiresIn: int(s.jwt.TTL().Seconds())}, nil
}

// Refresh troca um refresh token por um novo par (rotação). Se um token já
// usado for reapresentado, toda a família de sessões é revogada.
func (s *AdminAuth) Refresh(ctx context.Context, refreshToken string, meta SessionMeta) (*TokenPair, error) {
	hash := HashToken(refreshToken)
	var pair *TokenPair
	var rejected bool
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.ConsumeAdminSession(ctx, hash)
		if errors.Is(err, pgx.ErrNoRows) {
			old, err := q.GetAdminSessionByTokenHash(ctx, hash)
			if err == nil {
				// Reutilização detectada: revoga a família e confirma a transação.
				s.log.WarnContext(ctx, "reutilização de refresh token detectada", "admin_id", old.AdminID, "family_id", old.FamilyID)
				if err := q.RevokeAdminSessionFamily(ctx, old.FamilyID); err != nil {
					return err
				}
				_ = s.audit.Record(ctx, q, audit.Entry{ActorType: "admin", ActorID: old.AdminID.String(), Action: "auth.refresh-reuse", Entity: "admin_session", EntityID: old.FamilyID.String(), IP: meta.IP})
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			rejected = true
			return nil
		}
		if err != nil {
			return err
		}
		admin, err := q.GetAdminByID(ctx, row.AdminID)
		if err != nil {
			return err
		}
		if !row.ExpiresAt.After(s.nowFunc()) || !admin.IsActive || admin.Blocked {
			if err := q.RevokeAdminSessionFamily(ctx, row.FamilyID); err != nil {
				return err
			}
			rejected = true
			return nil
		}
		pair, err = s.createSession(ctx, q, row.AdminID, row.FamilyID, row.FamilyStartedAt, meta)
		if errors.Is(err, errInvalidRefresh) {
			rejected = true
			return nil
		}
		return err
	})
	if err != nil {
		return nil, err
	}
	if rejected {
		return nil, errInvalidRefresh
	}
	return pair, nil
}

// AdminPrincipal é o administrador autenticado na requisição.
type AdminPrincipal struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Email     string
	FirstName string
	LastName  string
	// DepartmentID e Roles alimentam o motor de permissões.
	DepartmentID *uuid.UUID
	Roles        []authz.Role
}

// Authenticate valida o access token e confirma, no banco, que a sessão segue
// ativa e o administrador não foi desativado ou bloqueado (efeito imediato).
func (s *AdminAuth) Authenticate(ctx context.Context, accessToken string) (*AdminPrincipal, error) {
	claims, err := s.jwt.Parse(accessToken)
	if err != nil {
		return nil, httpapi.ErrUnauthorized()
	}
	adminID, err := claims.SubjectID()
	if err != nil {
		return nil, httpapi.ErrUnauthorized()
	}
	row, err := s.q.AuthenticateAdminSession(ctx, db.AuthenticateAdminSessionParams{FamilyID: claims.SessionID, AdminID: adminID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrUnauthorized()
	}
	if err != nil {
		return nil, err
	}
	roles, err := s.q.ListRolesForAdmin(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	p := &AdminPrincipal{ID: row.ID, SessionID: claims.SessionID, Email: row.Email, FirstName: row.FirstName, LastName: row.LastName, DepartmentID: row.DepartmentID}
	for _, r := range roles {
		p.Roles = append(p.Roles, authz.Role{ID: r.ID, Code: r.Code})
	}
	return p, nil
}

// Logout revoga a sessão atual.
func (s *AdminAuth) Logout(ctx context.Context, p *AdminPrincipal) error {
	return s.q.RevokeAdminSessionFamily(ctx, p.SessionID)
}

// SessionView é uma sessão ativa exposta pela API.
type SessionView struct {
	ID         uuid.UUID `json:"id"`
	UserAgent  string    `json:"userAgent"`
	IP         string    `json:"ip"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUsedAt time.Time `json:"lastUsedAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	Current    bool      `json:"current"`
}

// ListSessions lista as sessões ativas do administrador.
func (s *AdminAuth) ListSessions(ctx context.Context, p *AdminPrincipal) ([]SessionView, error) {
	rows, err := s.q.ListActiveAdminSessions(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	out := make([]SessionView, 0, len(rows))
	for _, r := range rows {
		out = append(out, SessionView{ID: r.FamilyID, UserAgent: r.UserAgent, IP: r.Ip, CreatedAt: r.FamilyStartedAt,
			LastUsedAt: r.LastUsedAt, ExpiresAt: r.ExpiresAt, Current: r.FamilyID == p.SessionID})
	}
	return out, nil
}

// RevokeSession revoga uma sessão do próprio administrador.
func (s *AdminAuth) RevokeSession(ctx context.Context, p *AdminPrincipal, id uuid.UUID) error {
	n, err := s.q.RevokeAdminSessionFamilyOfAdmin(ctx, db.RevokeAdminSessionFamilyOfAdminParams{FamilyID: id, AdminID: p.ID})
	if err != nil {
		return err
	}
	if n == 0 {
		return httpapi.ErrNotFound()
	}
	return nil
}

// RevokeAllSessions revoga todas as sessões; com keepCurrent preserva a atual.
func (s *AdminAuth) RevokeAllSessions(ctx context.Context, p *AdminPrincipal, keepCurrent bool) error {
	if keepCurrent {
		return s.q.RevokeOtherAdminSessions(ctx, db.RevokeOtherAdminSessionsParams{AdminID: p.ID, FamilyID: p.SessionID})
	}
	return s.q.RevokeAllAdminSessions(ctx, p.ID)
}

// ForgotPassword envia o link de redefinição. Sempre retorna sucesso para o
// chamador, exista ou não o e-mail.
func (s *AdminAuth) ForgotPassword(ctx context.Context, email string) error {
	admin, err := s.q.GetAdminByEmail(ctx, strings.TrimSpace(email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !admin.IsActive || admin.Blocked {
		return nil
	}
	plain, hash, err := NewOpaqueToken()
	if err != nil {
		return err
	}
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if err := q.InvalidateOneTimeTokens(ctx, db.InvalidateOneTimeTokensParams{SubjectType: subjectAdmin, SubjectID: admin.ID, Purpose: purposeReset}); err != nil {
			return err
		}
		return q.CreateOneTimeToken(ctx, db.CreateOneTimeTokenParams{SubjectType: subjectAdmin, SubjectID: admin.ID, Purpose: purposeReset, TokenHash: hash, ExpiresAt: s.nowFunc().Add(s.cfg.ResetTokenTTL)})
	})
	if err != nil {
		return err
	}
	link := strings.TrimRight(s.cfg.PublicURL, "/") + "/admin/reset-password?token=" + plain
	if err := s.mail.Send(ctx, mailer.Message{To: admin.Email, Subject: "Redefinição de senha",
		Body: "Para redefinir sua senha acesse: " + link + "\nO link expira em breve e só pode ser usado uma vez."}); err != nil {
		s.log.ErrorContext(ctx, "falha ao enviar e-mail de redefinição", "error", err)
	}
	return nil
}

func (s *AdminAuth) checkPassword(password string) error {
	if err := s.hasher.CheckPolicy(password); err != nil {
		return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: "password", Message: strings.TrimPrefix(err.Error(), ErrPasswordPolicy.Error()+": ")}}})
	}
	return nil
}

// ResetPassword define a nova senha com o token de redefinição e revoga as sessões.
func (s *AdminAuth) ResetPassword(ctx context.Context, token, password string) error {
	if err := s.checkPassword(password); err != nil {
		return err
	}
	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		t, err := q.ConsumeOneTimeToken(ctx, db.ConsumeOneTimeTokenParams{TokenHash: HashToken(token), Purpose: purposeReset})
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidOneTime
		}
		if err != nil {
			return err
		}
		if err := q.SetAdminPassword(ctx, db.SetAdminPasswordParams{ID: t.SubjectID, PasswordHash: &hashed}); err != nil {
			return err
		}
		if err := q.RevokeAllAdminSessions(ctx, t.SubjectID); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "admin", ActorID: t.SubjectID.String(), Action: "auth.password-reset", Entity: "admin", EntityID: t.SubjectID.String()})
	})
}

// AcceptInvite ativa a conta convidada, define a senha e já abre uma sessão.
func (s *AdminAuth) AcceptInvite(ctx context.Context, token, password string, meta SessionMeta) (*TokenPair, error) {
	if err := s.checkPassword(password); err != nil {
		return nil, err
	}
	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	var pair *TokenPair
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		t, err := q.ConsumeOneTimeToken(ctx, db.ConsumeOneTimeTokenParams{TokenHash: HashToken(token), Purpose: purposeInvite})
		if errors.Is(err, pgx.ErrNoRows) {
			return errInvalidOneTime
		}
		if err != nil {
			return err
		}
		admin, err := q.GetAdminByID(ctx, t.SubjectID)
		if err != nil {
			return err
		}
		if admin.Blocked {
			return errInvalidOneTime
		}
		if err := q.SetAdminPassword(ctx, db.SetAdminPasswordParams{ID: admin.ID, PasswordHash: &hashed}); err != nil {
			return err
		}
		if err := q.ActivateAdmin(ctx, admin.ID); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, audit.Entry{ActorType: "admin", ActorID: admin.ID.String(), Action: "admin.invite-accepted", Entity: "admin", EntityID: admin.ID.String(), IP: meta.IP}); err != nil {
			return err
		}
		pair, err = s.createSession(ctx, q, admin.ID, uuid.New(), s.nowFunc(), meta)
		return err
	})
	return pair, err
}
