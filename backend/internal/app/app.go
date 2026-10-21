// Pacote app faz a injeção manual de dependências, compartilhada pelo servidor e pelos testes.
package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/api"
	"github.com/wagner/convenios-e-fomentos/backend/internal/admin"
	"github.com/wagner/convenios-e-fomentos/backend/internal/apitoken"
	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/config"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/document"
	"github.com/wagner/convenios-e-fomentos/backend/internal/enduser"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

// Hasher cria o hasher de senhas a partir da configuração.
func Hasher(cfg *config.Config) auth.Hasher {
	return auth.Hasher{Memory: cfg.Argon2MemoryKiB, Iterations: cfg.Argon2Iterations, Parallelism: cfg.Argon2Parallel, MinLength: cfg.PasswordMinLength}
}

// AdminService cria o serviço de administradores sem motor de permissões
// (uso da CLI, que não avalia permissões).
func AdminService(pool *pgxpool.Pool) *admin.Service {
	return admin.NewService(pool, nil)
}

// App agrupa o roteador e as peças que precisam de ciclo de vida.
type App struct {
	Handler http.Handler
	// Router permite registrar rotas dos recursos (Fase 7) e injetar rotas nos testes.
	Router  *chi.Mux
	Authz   *authz.Engine
	Actions *authz.Registry
	// Identifier e EndUsers servem para montar as rotas públicas dos recursos (Fase 7).
	Identifier *enduser.Identifier
	EndUsers   *enduser.Service
	APITokens  *apitoken.Service

	log  *slog.Logger
	pool *pgxpool.Pool
}

// New monta todo o grafo de dependências.
func New(ctx context.Context, cfg *config.Config, log *slog.Logger, pool *pgxpool.Pool, mail mailer.Mailer) (*App, error) {
	trusted, err := httpapi.ParseCIDRs(cfg.TrustedProxies)
	if err != nil {
		return nil, err
	}
	q := db.New(pool)

	// Motor de permissões: ações e condições são registradas em código.
	actions := authz.NewRegistry()
	authz.RegisterCoreActions(actions)
	document.Register(actions) // recurso de exemplo
	conds := authz.NewConditionRegistry()
	store := authz.DBStore{Q: q}
	authz.RegisterStandardConditions(conds, store)
	engine := authz.NewEngine(actions, conds, store, authz.PGBroadcaster{Pool: pool, Log: log}, log)
	engine.SetEndUserStore(authz.EndUserStore{Q: q})

	// Sem JWKS_URI (CLI) não há como validar tokens de administrador; o servidor da API exige a
	// configuração antes de subir (Config.ValidateOIDC).
	var oidc *auth.OIDCVerifier
	if cfg.JWKSURI != "" {
		oidc, err = auth.NewOIDCVerifier(ctx, auth.OIDCConfig{JWKSURI: cfg.JWKSURI, Issuer: cfg.OIDCIssuer, ClientID: cfg.OIDCClientID, CPFClaim: cfg.OIDCCPFClaim}, log)
		if err != nil {
			return nil, err
		}
	}
	adminAuth := auth.NewAdminAuth(pool, oidc, log)
	adminSvc := admin.NewService(pool, engine)
	adminHandler := admin.NewHandler(adminSvc, adminAuth, engine)

	apiJWT := auth.NewJWTIssuer(cfg.JWTAPISecret, auth.AudienceAPI, cfg.AccessTokenTTL)
	endUserSvc, err := enduser.NewService(pool, Hasher(cfg), apiJWT, mail, engine, enduser.Config{
		RefreshMaxTTL: cfg.RefreshTokenMaxTTL, RefreshIdleTTL: cfg.RefreshTokenIdleTTL, ResetTTL: cfg.ResetTokenTTL,
		ConfirmTTL: cfg.EmailConfirmationTTL, PublicURL: cfg.PublicURL, DefaultSender: cfg.SMTPFrom,
	}, log)
	if err != nil {
		return nil, err
	}
	tokenSvc := apitoken.NewService(pool, engine, cfg.APITokenSalt, log)
	identifier := &enduser.Identifier{Svc: endUserSvc, Engine: engine, Tokens: tokenSvc}
	endUserHandler := enduser.NewHandler(endUserSvc, identifier, cfg.RateLimitAuthPerMin)
	endUserAdmin := enduser.NewAdminHandler(endUserSvc, adminHandler.Guard)
	tokenAdmin := apitoken.NewHandler(tokenSvc, adminHandler.Guard)
	documentSvc := document.NewService(pool, engine)
	documentAdmin := document.NewAdminHandler(documentSvc, engine, adminHandler.Guard)
	documentPublic := document.NewPublicHandler(documentSvc, identifier)

	r := httpapi.NewRouter(httpapi.RouterOptions{
		Log: log, DB: pool, AllowedOrigins: cfg.CORSAllowedOrigins, MaxBodyBytes: cfg.MaxBodyBytes, GlobalRatePerMinute: cfg.RateLimitGlobalPerMin,
		OpenAPISpec: api.OpenAPI, TrustedProxies: trusted,
		Routes: func(r chi.Router) {
			adminHandler.Routes(r)
			audit.NewHandler(q).Routes(r, adminHandler.Guard)
			endUserAdmin.Routes(r)
			tokenAdmin.Routes(r)
			documentAdmin.Routes(r)
			documentPublic.Routes(r)
			endUserHandler.Routes(r)
		},
	})
	mux := r.(*chi.Mux)
	return &App{Handler: mux, Router: mux, Authz: engine, Actions: actions, Identifier: identifier, EndUsers: endUserSvc, APITokens: tokenSvc, log: log, pool: pool}, nil
}

// Start sincroniza o registro de ações com o banco e passa a escutar as
// invalidações de cache vindas de outras instâncias, até o contexto terminar.
func (a *App) Start(ctx context.Context) error {
	if err := a.EndUsers.EnsureDefaults(ctx); err != nil {
		return err
	}
	if err := authz.SyncEndUserActions(ctx, db.New(a.pool), a.Actions, a.Authz, a.log); err != nil {
		return err
	}
	if err := authz.SyncActions(ctx, db.New(a.pool), a.Actions, a.Authz, a.log); err != nil {
		return err
	}
	go func() {
		if err := a.Authz.Listen(ctx); err != nil {
			a.log.ErrorContext(ctx, "escuta de invalidações encerrada", "error", err)
		}
	}()
	return nil
}
