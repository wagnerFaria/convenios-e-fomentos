// Pacote admin trata dos administradores, suas roles e do fluxo de convite.
package admin

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Dados da role especial Super Admin.
const (
	SuperAdminCode        = authz.SuperAdminRoleCode
	superAdminName        = "Super Admin"
	superAdminDescription = "Acesso total ao sistema."
)

var (
	// ErrSetupDone indica que o registro inicial já foi feito.
	errSetupDone  = &httpapi.APIError{Status: 403, Code: "setup_completed", Message: "O registro inicial já foi concluído."}
	errEmailTaken = &httpapi.APIError{Status: 409, Code: "email_taken", Message: "Já existe um administrador com este e-mail."}
)

// Config reúne os parâmetros do serviço.
type Config struct {
	InviteTTL time.Duration
	PublicURL string
}

// Service implementa as operações sobre administradores.
type Service struct {
	pool   *pgxpool.Pool
	q      *db.Queries
	hasher auth.Hasher
	audit  audit.Recorder
	cfg    Config
	engine *authz.Engine
}

// NewService cria o serviço.
func NewService(pool *pgxpool.Pool, hasher auth.Hasher, cfg Config, engine *authz.Engine) *Service {
	return &Service{pool: pool, q: db.New(pool), hasher: hasher, cfg: cfg, engine: engine}
}

// CreateSuperAdminInput são os dados do primeiro (ou de mais um) Super Admin.
type CreateSuperAdminInput struct {
	FirstName, LastName, Email, Password string
}

func passwordError(err error) error {
	return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{
		{Field: "password", Message: strings.TrimPrefix(err.Error(), auth.ErrPasswordPolicy.Error()+": ")}}})
}

// CreateSuperAdmin cria um administrador ativo com a role Super Admin. Com
// onlyIfNone, só funciona enquanto não existir nenhum administrador (registro
// inicial); a verificação é protegida por lock consultivo contra corridas.
func (s *Service) CreateSuperAdmin(ctx context.Context, in CreateSuperAdminInput, onlyIfNone bool, ip string) (db.Admin, error) {
	if err := s.hasher.CheckPolicy(in.Password); err != nil {
		return db.Admin{}, passwordError(err)
	}
	hashed, err := s.hasher.Hash(in.Password)
	if err != nil {
		return db.Admin{}, err
	}
	var created db.Admin
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if onlyIfNone {
			if err := q.LockAdminSetup(ctx); err != nil {
				return err
			}
			n, err := q.CountAdmins(ctx)
			if err != nil {
				return err
			}
			if n > 0 {
				return errSetupDone
			}
		}
		role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: SuperAdminCode, Name: superAdminName, Description: superAdminDescription, IsSystem: true})
		if err != nil {
			return err
		}
		created, err = q.CreateAdmin(ctx, db.CreateAdminParams{FirstName: in.FirstName, LastName: in.LastName, Email: strings.TrimSpace(in.Email), PasswordHash: &hashed, IsActive: true})
		if err != nil {
			if database.IsUniqueViolation(err) {
				return errEmailTaken
			}
			return err
		}
		if err := q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: created.ID, RoleID: role.ID}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "system", Action: "admin.create-super-admin", Entity: "admin", EntityID: created.ID.String(), IP: ip,
			Payload: map[string]any{"after": map[string]any{"email": created.Email}}})
	})
	return created, err
}

// ResetPasswordByEmail redefine a senha (uso da CLI) e revoga as sessões.
func (s *Service) ResetPasswordByEmail(ctx context.Context, email, password string) error {
	if err := s.hasher.CheckPolicy(password); err != nil {
		return passwordError(err)
	}
	hashed, err := s.hasher.Hash(password)
	if err != nil {
		return err
	}
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		a, err := q.GetAdminByEmail(ctx, strings.TrimSpace(email))
		if errors.Is(err, pgx.ErrNoRows) {
			return httpapi.ErrNotFound()
		}
		if err != nil {
			return err
		}
		if err := q.SetAdminPassword(ctx, db.SetAdminPasswordParams{ID: a.ID, PasswordHash: &hashed}); err != nil {
			return err
		}
		if err := q.RevokeAllAdminSessions(ctx, a.ID); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "system", Action: "admin.password-reset-cli", Entity: "admin", EntityID: a.ID.String()})
	})
}

// InviteInput são os dados de um convite.
type InviteInput struct {
	FirstName    string
	LastName     string
	Email        string
	RoleIDs      []uuid.UUID
	DepartmentID *uuid.UUID
}

// InviteResult é o resultado de um convite: a URL contém o token, exibido apenas aqui.
type InviteResult struct {
	Admin     db.Admin
	InviteURL string
	ExpiresAt time.Time
}

// Invite cria o administrador inativo com as roles e gera o token de convite.
func (s *Service) Invite(ctx context.Context, actor *auth.AdminPrincipal, in InviteInput, ip string) (*InviteResult, error) {
	roleIDs := slices.Clone(in.RoleIDs)
	slices.SortFunc(roleIDs, func(a, b uuid.UUID) int { return strings.Compare(a.String(), b.String()) })
	roleIDs = slices.Compact(roleIDs)

	plain, hash, err := auth.NewOpaqueToken()
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(s.cfg.InviteTTL)
	var created db.Admin
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		roles, err := q.GetAdminRolesByIDs(ctx, roleIDs)
		if err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return fieldIssue("roleIds", "Uma ou mais roles não existem.")
		}
		// Só um Super Admin pode conceder a role Super Admin (evita escalada de privilégio).
		if hasSuperRole(roles) && !SubjectOf(actor).IsSuperAdmin() {
			return httpapi.ErrForbidden()
		}
		created, err = q.CreateAdmin(ctx, db.CreateAdminParams{FirstName: in.FirstName, LastName: in.LastName, Email: strings.TrimSpace(in.Email), IsActive: false, DepartmentID: in.DepartmentID})
		if err != nil {
			if database.IsUniqueViolation(err) {
				return errEmailTaken
			}
			return err
		}
		for _, id := range roleIDs {
			if err := q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: created.ID, RoleID: id}); err != nil {
				return err
			}
		}
		if err := q.CreateOneTimeToken(ctx, db.CreateOneTimeTokenParams{SubjectType: "admin", SubjectID: created.ID, Purpose: "invite", TokenHash: hash, ExpiresAt: expires}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "admin", ActorID: actor.ID.String(), Action: "admin.invite", Entity: "admin", EntityID: created.ID.String(), IP: ip,
			Payload: map[string]any{"after": map[string]any{"email": created.Email, "roleIds": roleIDs}}})
	})
	if err != nil {
		return nil, err
	}
	return &InviteResult{Admin: created, ExpiresAt: expires, InviteURL: strings.TrimRight(s.cfg.PublicURL, "/") + "/admin/accept-invite?token=" + plain}, nil
}

// RoleView é a role resumida exposta pela API.
type RoleView struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

// MeView é o administrador autenticado com as permissões efetivas (para o
// front esconder botões; a autorização real é sempre feita no servidor).
type MeView struct {
	ID           uuid.UUID                   `json:"id"`
	Email        string                      `json:"email"`
	FirstName    string                      `json:"firstName"`
	LastName     string                      `json:"lastName"`
	DepartmentID *uuid.UUID                  `json:"departmentId"`
	IsSuperAdmin bool                        `json:"isSuperAdmin"`
	Roles        []RoleView                  `json:"roles"`
	Permissions  []authz.EffectivePermission `json:"permissions"`
}

// Me monta a visão do administrador autenticado.
func (s *Service) Me(ctx context.Context, p *auth.AdminPrincipal) (*MeView, error) {
	subj := SubjectOf(p)
	perms, err := s.engine.Effective(ctx, subj)
	if err != nil {
		return nil, err
	}
	roles, err := s.q.ListRolesForAdmin(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	rv := make([]RoleView, 0, len(roles))
	for _, r := range roles {
		rv = append(rv, RoleView{ID: r.ID, Code: r.Code, Name: r.Name})
	}
	return &MeView{ID: p.ID, Email: p.Email, FirstName: p.FirstName, LastName: p.LastName, DepartmentID: p.DepartmentID,
		IsSuperAdmin: subj.IsSuperAdmin(), Roles: rv, Permissions: perms}, nil
}
