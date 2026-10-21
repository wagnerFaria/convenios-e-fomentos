package enduser

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

const (
	subjectEndUser     = "end_user"
	purposeConfirm     = "confirmation"
	purposeReset       = "reset"
	maxExtraValueBytes = 500
)

// RegisterInput é o cadastro público. Extra traz os campos adicionais permitidos pela configuração.
type RegisterInput struct {
	Username, Email, Password string
	Extra                     map[string]any
}

// RegisterResult: com confirmação obrigatória só User é preenchido; senão também Tokens.
type RegisterResult struct {
	User   UserView        `json:"user"`
	Tokens *auth.TokenPair `json:"tokens,omitempty"`
}

func (s *Service) checkPassword(password string) error {
	if err := s.hasher.CheckPolicy(password); err != nil {
		return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: "password", Message: strings.TrimPrefix(err.Error(), auth.ErrPasswordPolicy.Error()+": ")}}})
	}
	return nil
}

// Register cria a conta respeitando as configurações avançadas.
func (s *Service) Register(ctx context.Context, in RegisterInput, meta auth.SessionMeta) (*RegisterResult, error) {
	st, err := s.settings(ctx, s.q)
	if err != nil {
		return nil, err
	}
	if !st.RegistrationEnabled {
		return nil, errRegistrationDisabled
	}
	if err := s.checkPassword(in.Password); err != nil {
		return nil, err
	}
	hashed, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}
	profile, err := json.Marshal(nonNil(in.Extra))
	if err != nil {
		return nil, err
	}
	var user db.EndUser
	var role db.EndUserRole
	var confirmToken string
	var pair *auth.TokenPair
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
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
		role, err = q.GetEndUserRole(ctx, st.DefaultRoleID)
		if err != nil {
			return fmt.Errorf("role padrão: %w", err)
		}
		user, err = q.CreateEndUser(ctx, db.CreateEndUserParams{Username: strings.TrimSpace(in.Username), Email: email, PasswordHash: &hashed,
			Confirmed: !st.EmailConfirmationRequired, RoleID: role.ID, Profile: profile})
		if database.IsUniqueViolation(err) {
			return errTaken
		}
		if err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, audit.Entry{ActorType: subjectEndUser, ActorID: user.ID.String(), Action: "end-user.register", Entity: "end_user", EntityID: user.ID.String(), IP: meta.IP}); err != nil {
			return err
		}
		if st.EmailConfirmationRequired {
			confirmToken, err = s.newOneTimeToken(ctx, q, user.ID, purposeConfirm, s.cfg.ConfirmTTL)
			return err
		}
		pair, err = s.createSession(ctx, q, user.ID, uuid.New(), s.now(), meta)
		return err
	})
	if err != nil {
		return nil, err
	}
	if confirmToken != "" {
		s.sendConfirmation(ctx, user, confirmToken)
	}
	return &RegisterResult{User: viewOf(user, role), Tokens: pair}, nil
}

func nonNil(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return m
}

func (s *Service) now() time.Time { return time.Now() }

// newOneTimeToken invalida os tokens anteriores do mesmo tipo e cria um novo (devolve o valor em claro).
func (s *Service) newOneTimeToken(ctx context.Context, q *db.Queries, userID uuid.UUID, purpose string, ttl time.Duration) (string, error) {
	plain, hash, err := auth.NewOpaqueToken()
	if err != nil {
		return "", err
	}
	if err := q.InvalidateOneTimeTokens(ctx, db.InvalidateOneTimeTokensParams{SubjectType: subjectEndUser, SubjectID: userID, Purpose: purpose}); err != nil {
		return "", err
	}
	err = q.CreateOneTimeToken(ctx, db.CreateOneTimeTokenParams{SubjectType: subjectEndUser, SubjectID: userID, Purpose: purpose, TokenHash: hash, ExpiresAt: s.now().Add(ttl)})
	return plain, err
}

func (s *Service) sendTemplate(ctx context.Context, key string, u db.EndUser, token, link string) {
	t, err := s.q.GetEmailTemplate(ctx, key)
	if err != nil {
		s.log.ErrorContext(ctx, "template de e-mail indisponível", "key", key, "error", err)
		return
	}
	subject, body, err := renderTemplate(t, TemplateData{User: TemplateUser{Username: u.Username, Email: u.Email}, Token: token, URL: link, ServerURL: strings.TrimRight(s.cfg.PublicURL, "/")})
	if err != nil {
		s.log.ErrorContext(ctx, "falha ao renderizar o template de e-mail", "key", key, "error", err)
		return
	}
	if err := s.mail.Send(ctx, mailer.Message{To: u.Email, From: t.SenderEmail, FromName: t.SenderName, ReplyTo: t.ReplyTo, Subject: subject, HTML: body}); err != nil {
		s.log.ErrorContext(ctx, "falha ao enviar e-mail", "key", key, "error", err)
	}
}

func (s *Service) sendConfirmation(ctx context.Context, u db.EndUser, token string) {
	link := strings.TrimRight(s.cfg.PublicURL, "/") + "/api/auth/email-confirmation?confirmation=" + url.QueryEscape(token)
	s.sendTemplate(ctx, TemplateEmailConfirmation, u, token, link)
}

func (s *Service) sendReset(ctx context.Context, st Settings, u db.EndUser, token string) {
	base := st.ResetPasswordURL
	if base == "" {
		base = strings.TrimRight(s.cfg.PublicURL, "/") + "/reset-password"
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	s.sendTemplate(ctx, TemplateResetPassword, u, token, base+sep+"token="+url.QueryEscape(token))
}

// createSession grava um novo refresh token na família e emite o par de tokens.
func (s *Service) createSession(ctx context.Context, q *db.Queries, userID, family uuid.UUID, familyStart time.Time, meta auth.SessionMeta) (*auth.TokenPair, error) {
	now := s.now()
	expires := now.Add(s.cfg.RefreshIdleTTL)
	if abs := familyStart.Add(s.cfg.RefreshMaxTTL); abs.Before(expires) {
		expires = abs
	}
	if !expires.After(now) {
		return nil, errInvalidRefresh
	}
	meta = meta.Trimmed()
	plain, hash, err := auth.NewOpaqueToken()
	if err != nil {
		return nil, err
	}
	if _, err = q.CreateEndUserSession(ctx, db.CreateEndUserSessionParams{EndUserID: userID, FamilyID: family, FamilyStartedAt: familyStart,
		RefreshTokenHash: hash, UserAgent: meta.UserAgent, Ip: meta.IP, ExpiresAt: expires}); err != nil {
		return nil, fmt.Errorf("criando sessão: %w", err)
	}
	access, err := s.jwt.Issue(userID, family)
	if err != nil {
		return nil, err
	}
	return &auth.TokenPair{AccessToken: access, RefreshToken: plain, TokenType: "Bearer", ExpiresIn: int(s.jwt.TTL().Seconds())}, nil
}

// LoginResult é o resultado de um login.
type LoginResult struct {
	User   UserView        `json:"user"`
	Tokens *auth.TokenPair `json:"tokens"`
}

// Login autentica por e-mail ou nome de usuário. A senha é sempre verificada,
// exista a conta ou não; só depois de acertar a senha o cliente é informado de
// bloqueio ou falta de confirmação.
func (s *Service) Login(ctx context.Context, identifier, password string, meta auth.SessionMeta) (*LoginResult, error) {
	candidates, err := s.q.GetEndUsersByIdentifier(ctx, strings.TrimSpace(identifier))
	if err != nil {
		return nil, err
	}
	var matched *db.EndUser
	for i := range candidates {
		c := candidates[i]
		if c.PasswordHash != nil && c.Provider == "local" && s.hasher.Verify(password, *c.PasswordHash) {
			matched = &c
			break
		}
	}
	if len(candidates) == 0 {
		s.hasher.Verify(password, s.dummy) // iguala o tempo de resposta
	}
	if matched == nil {
		return nil, errInvalidCredentials
	}
	if matched.Blocked {
		return nil, errBlocked
	}
	st, err := s.settings(ctx, s.q)
	if err != nil {
		return nil, err
	}
	if st.EmailConfirmationRequired && !matched.Confirmed {
		return nil, errNotConfirmed
	}
	pair, err := s.createSession(ctx, s.q, matched.ID, uuid.New(), s.now(), meta)
	if err != nil {
		return nil, err
	}
	v, err := s.viewByID(ctx, s.q, matched.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResult{User: *v, Tokens: pair}, nil
}

// Refresh troca o refresh token por um novo par; reutilização revoga a família inteira.
func (s *Service) Refresh(ctx context.Context, refreshToken string, meta auth.SessionMeta) (*auth.TokenPair, error) {
	hash := auth.HashToken(refreshToken)
	var pair *auth.TokenPair
	var rejected bool
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		row, err := q.ConsumeEndUserSession(ctx, hash)
		if errors.Is(err, pgx.ErrNoRows) {
			old, err := q.GetEndUserSessionByTokenHash(ctx, hash)
			if err == nil {
				s.log.WarnContext(ctx, "reutilização de refresh token detectada", "end_user_id", old.EndUserID, "family_id", old.FamilyID)
				if err := q.RevokeEndUserSessionFamily(ctx, old.FamilyID); err != nil {
					return err
				}
				_ = s.audit.Record(ctx, q, audit.Entry{ActorType: subjectEndUser, ActorID: old.EndUserID.String(), Action: "auth.refresh-reuse", Entity: "end_user_session", EntityID: old.FamilyID.String(), IP: meta.IP})
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			rejected = true
			return nil
		}
		if err != nil {
			return err
		}
		u, err := q.GetEndUserByID(ctx, row.EndUserID)
		if err != nil {
			return err
		}
		st, err := s.settings(ctx, q)
		if err != nil {
			return err
		}
		if !row.ExpiresAt.After(s.now()) || u.Blocked || (st.EmailConfirmationRequired && !u.Confirmed) {
			rejected = true
			return q.RevokeEndUserSessionFamily(ctx, row.FamilyID)
		}
		pair, err = s.createSession(ctx, q, row.EndUserID, row.FamilyID, row.FamilyStartedAt, meta)
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

// Principal é o usuário final autenticado na requisição.
type Principal struct {
	ID        uuid.UUID
	SessionID uuid.UUID
	Username  string
	Email     string
	RoleID    uuid.UUID
	RoleType  string
}

// Subject converte o usuário no sujeito do motor de permissões.
func (p *Principal) Subject() authz.Subject {
	return authz.Subject{Type: authz.SubjectEndUser, ID: p.ID, Roles: []authz.Role{{ID: p.RoleID, Code: p.RoleType}}}
}

// Authenticate valida o access token (audience api) e confirma no banco que a
// sessão está ativa e a conta não foi bloqueada (efeito imediato).
func (s *Service) Authenticate(ctx context.Context, accessToken string) (*Principal, error) {
	claims, err := s.jwt.Parse(accessToken)
	if err != nil {
		return nil, httpapi.ErrUnauthorized()
	}
	userID, err := claims.SubjectID()
	if err != nil {
		return nil, httpapi.ErrUnauthorized()
	}
	row, err := s.q.AuthenticateEndUserSession(ctx, db.AuthenticateEndUserSessionParams{FamilyID: claims.SessionID, EndUserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrUnauthorized()
	}
	if err != nil {
		return nil, err
	}
	return &Principal{ID: row.ID, SessionID: claims.SessionID, Username: row.Username, Email: row.Email, RoleID: row.RoleID, RoleType: row.RoleType}, nil
}

// Logout revoga a sessão atual.
func (s *Service) Logout(ctx context.Context, p *Principal) error {
	return s.q.RevokeEndUserSessionFamily(ctx, p.SessionID)
}

// Me devolve os dados do usuário autenticado.
func (s *Service) Me(ctx context.Context, p *Principal) (*UserView, error) {
	return s.viewByID(ctx, s.q, p.ID)
}

// ConfirmEmail consome o token de confirmação e devolve a URL de redirecionamento configurada ("" se não houver).
func (s *Service) ConfirmEmail(ctx context.Context, token string) (string, error) {
	var redirect string
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		t, err := q.ConsumeOneTimeToken(ctx, db.ConsumeOneTimeTokenParams{TokenHash: auth.HashToken(token), Purpose: purposeConfirm})
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && t.SubjectType != subjectEndUser) {
			return errInvalidOneTime
		}
		if err != nil {
			return err
		}
		if err := q.SetEndUserConfirmed(ctx, db.SetEndUserConfirmedParams{ID: t.SubjectID, Confirmed: true}); err != nil {
			return err
		}
		st, err := s.settings(ctx, q)
		if err != nil {
			return err
		}
		redirect = st.EmailConfirmationRedirectURL
		return s.audit.Record(ctx, q, audit.Entry{ActorType: subjectEndUser, ActorID: t.SubjectID.String(), Action: "end-user.email-confirmed", Entity: "end_user", EntityID: t.SubjectID.String()})
	})
	return redirect, err
}

// SendEmailConfirmation reenvia a confirmação. Sempre "sucesso" para o chamador,
// exista a conta ou não, para não revelar quais e-mails estão cadastrados.
func (s *Service) SendEmailConfirmation(ctx context.Context, email string) error {
	users, err := s.q.GetEndUsersByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Confirmed || u.Blocked || u.Provider != "local" {
			continue
		}
		token, err := s.newOneTimeToken(ctx, s.q, u.ID, purposeConfirm, s.cfg.ConfirmTTL)
		if err != nil {
			return err
		}
		s.sendConfirmation(ctx, u, token)
	}
	return nil
}

// ForgotPassword envia o link de redefinição (sempre "sucesso" para o chamador).
func (s *Service) ForgotPassword(ctx context.Context, email string) error {
	users, err := s.q.GetEndUsersByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}
	st, err := s.settings(ctx, s.q)
	if err != nil {
		return err
	}
	for _, u := range users {
		if u.Blocked || u.Provider != "local" {
			continue
		}
		token, err := s.newOneTimeToken(ctx, s.q, u.ID, purposeReset, s.cfg.ResetTTL)
		if err != nil {
			return err
		}
		s.sendReset(ctx, st, u, token)
	}
	return nil
}

// ResetPassword define a nova senha com o token e revoga todas as sessões.
func (s *Service) ResetPassword(ctx context.Context, token, password string) error {
	if err := s.checkPassword(password); err != nil {
		return err
	}
	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		t, err := q.ConsumeOneTimeToken(ctx, db.ConsumeOneTimeTokenParams{TokenHash: auth.HashToken(token), Purpose: purposeReset})
		if errors.Is(err, pgx.ErrNoRows) || (err == nil && t.SubjectType != subjectEndUser) {
			return errInvalidOneTime
		}
		if err != nil {
			return err
		}
		if err := q.SetEndUserPassword(ctx, db.SetEndUserPasswordParams{ID: t.SubjectID, PasswordHash: &hashed}); err != nil {
			return err
		}
		if err := q.RevokeAllEndUserSessions(ctx, t.SubjectID); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: subjectEndUser, ActorID: t.SubjectID.String(), Action: "end-user.password-reset", Entity: "end_user", EntityID: t.SubjectID.String()})
	})
}

// ChangePassword troca a senha exigindo a atual e revoga as demais sessões.
func (s *Service) ChangePassword(ctx context.Context, p *Principal, current, next string) error {
	u, err := s.q.GetEndUserByID(ctx, p.ID)
	if err != nil {
		return err
	}
	if u.PasswordHash == nil || !s.hasher.Verify(current, *u.PasswordHash) {
		return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: "currentPassword", Message: "Senha atual incorreta."}}})
	}
	if err := s.checkPassword(next); err != nil {
		return err
	}
	hashed, err := s.hasher.Hash(next)
	if err != nil {
		return err
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if err := q.SetEndUserPassword(ctx, db.SetEndUserPasswordParams{ID: p.ID, PasswordHash: &hashed}); err != nil {
			return err
		}
		if err := q.RevokeOtherEndUserSessions(ctx, db.RevokeOtherEndUserSessionsParams{EndUserID: p.ID, FamilyID: p.SessionID}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: subjectEndUser, ActorID: p.ID.String(), Action: "end-user.password-change", Entity: "end_user", EntityID: p.ID.String()})
	})
}
