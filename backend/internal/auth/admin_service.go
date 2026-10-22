package auth

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/cpf"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

var (
	errAccountDisabled = &httpapi.APIError{Status: 403, Code: "account_disabled", Message: "Seu acesso está desativado. Procure um administrador do sistema."}
	errNotConfigured   = &httpapi.APIError{Status: 503, Code: "login_unavailable", Message: "O login pelo MT Login não está configurado neste servidor."}
	errMissingCPF      = &httpapi.APIError{Status: 403, Code: "invalid_identity", Message: "Não foi possível identificar o CPF da sua conta do MT Login."}
)

// loginTouchInterval é o intervalo mínimo entre gravações do último acesso: evita um UPDATE por requisição.
const loginTouchInterval = 5 * time.Minute

// AdminAuth autentica administradores com tokens do MT Login. O CPF do token identifica o
// administrador; quem ainda não existe na base é criado, sem roles, no primeiro acesso.
type AdminAuth struct {
	pool     *pgxpool.Pool
	q        *db.Queries
	verifier *OIDCVerifier
	audit    audit.Recorder
	log      *slog.Logger
	nowFunc  func() time.Time
}

// NewAdminAuth cria o serviço.
func NewAdminAuth(pool *pgxpool.Pool, verifier *OIDCVerifier, log *slog.Logger) *AdminAuth {
	return &AdminAuth{pool: pool, q: db.New(pool), verifier: verifier, log: log, nowFunc: time.Now}
}

// AdminPrincipal é o administrador autenticado na requisição.
type AdminPrincipal struct {
	ID        uuid.UUID
	CPF       string // 11 dígitos
	Email     string
	FirstName string
	LastName  string
	// DepartmentID e Roles alimentam o motor de permissões.
	DepartmentID *uuid.UUID
	Roles        []authz.Role
}

// Authenticate valida o access token do MT Login e resolve o administrador pelo CPF, criando-o no
// primeiro acesso e atualizando nome e e-mail quando mudam. Confirma, no banco, que a conta segue
// ativa e não foi bloqueada, de modo que desativar ou bloquear vale de imediato.
func (s *AdminAuth) Authenticate(ctx context.Context, accessToken string) (*AdminPrincipal, error) {
	if s.verifier == nil {
		return nil, errNotConfigured
	}
	identity, err := s.verifier.Verify(accessToken)
	switch {
	case errors.Is(err, ErrMissingCPF):
		return nil, errMissingCPF
	case err != nil:
		return nil, httpapi.ErrUnauthorized()
	}
	admin, err := s.resolve(ctx, identity)
	if err != nil {
		return nil, err
	}
	if !admin.IsActive || admin.Blocked {
		return nil, errAccountDisabled
	}
	roles, err := s.q.ListRolesForAdmin(ctx, admin.ID)
	if err != nil {
		return nil, err
	}
	p := &AdminPrincipal{ID: admin.ID, CPF: identity.CPF, Email: admin.Email, FirstName: admin.FirstName, LastName: admin.LastName, DepartmentID: admin.DepartmentID}
	for _, r := range roles {
		p.Roles = append(p.Roles, authz.Role{ID: r.ID, Code: r.Code})
	}
	return p, nil
}

// resolve devolve o administrador do CPF, criando-o ou sincronizando seus dados conforme preciso.
func (s *AdminAuth) resolve(ctx context.Context, id *Identity) (db.Admin, error) {
	admin, err := s.q.GetAdminByCPF(ctx, &id.CPF)
	if errors.Is(err, pgx.ErrNoRows) {
		return s.provision(ctx, id)
	}
	if err != nil {
		return db.Admin{}, err
	}
	if !s.needsSync(admin, id) {
		return admin, nil
	}
	if err := s.q.SyncAdminFromLogin(ctx, db.SyncAdminFromLoginParams{ID: admin.ID, FirstName: id.FirstName, LastName: id.LastName, Email: id.Email}); err != nil {
		return db.Admin{}, err
	}
	admin.FirstName, admin.LastName, admin.Email = id.FirstName, id.LastName, id.Email
	return admin, nil
}

func (s *AdminAuth) needsSync(a db.Admin, id *Identity) bool {
	changed := a.FirstName != id.FirstName || a.LastName != id.LastName || a.Email != id.Email
	return changed || a.LastLoginAt == nil || s.nowFunc().Sub(*a.LastLoginAt) >= loginTouchInterval
}

// provision cria o administrador do primeiro acesso: ativo e sem nenhuma role.
func (s *AdminAuth) provision(ctx context.Context, id *Identity) (db.Admin, error) {
	var created db.Admin
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		var err error
		created, err = q.ProvisionAdminFromLogin(ctx, db.ProvisionAdminFromLoginParams{Cpf: &id.CPF, FirstName: id.FirstName, LastName: id.LastName, Email: id.Email})
		if errors.Is(err, pgx.ErrNoRows) {
			// Outra requisição do mesmo primeiro acesso criou o registro antes.
			created, err = q.GetAdminByCPF(ctx, &id.CPF)
			return err
		}
		if err != nil {
			return err
		}
		return s.audit.Record(ctx, q, audit.Entry{ActorType: "system", Action: "admin.auto-provisioned", Entity: "admin", EntityID: created.ID.String(),
			Payload: map[string]any{"after": map[string]any{"cpf": cpf.Mask(id.CPF), "email": id.Email}}})
	})
	return created, err
}
