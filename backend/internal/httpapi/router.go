package httpapi

import (
	"context"
	_ "embed"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Pinger verifica a disponibilidade de uma dependência (ex.: o banco).
type Pinger interface {
	Ping(ctx context.Context) error
}

// RouterOptions agrupa as dependências do roteador.
type RouterOptions struct {
	Log            *slog.Logger
	DB             Pinger
	AllowedOrigins []string
	MaxBodyBytes   int64
	// GlobalRatePerMinute limita cada IP em todas as rotas (0 desliga).
	GlobalRatePerMinute int
	OpenAPISpec         []byte
	TrustedProxies      []*net.IPNet
	// Routes registra as rotas dos domínios (admin, api...).
	Routes func(r chi.Router)
}

// APIPrefix é o prefixo público de todas as rotas da API.
const APIPrefix = "/api"

// NewRouter monta o roteador com os middlewares globais e as rotas de infraestrutura.
func NewRouter(o RouterOptions) http.Handler {
	r := chi.NewRouter()
	r.Use(RequestID, ClientIPMiddleware(o.TrustedProxies), RateLimitByIP(o.GlobalRatePerMinute), AccessLog(o.Log), Recoverer(o.Log), SecurityHeaders, BodyLimit(o.MaxBodyBytes))
	r.Use(middleware.GetHead)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: o.AllowedOrigins,
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type", requestIDHeader},
		ExposedHeaders: []string{requestIDHeader},
		MaxAge:         300,
	}))

	// Toda a API fica sob /api: o proxy reverso (HAProxy) encaminha esse prefixo ao backend.
	r.Route(APIPrefix, func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			JSON(w, http.StatusOK, map[string]string{"status": "ok"}, nil)
		})
		r.Get("/ready", func(w http.ResponseWriter, req *http.Request) {
			ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
			defer cancel()
			if err := o.DB.Ping(ctx); err != nil {
				o.Log.WarnContext(req.Context(), "banco indisponível", "error", err)
				Error(w, req, &APIError{Status: 503, Code: "not_ready", Message: "Banco de dados indisponível."})
				return
			}
			JSON(w, http.StatusOK, map[string]string{"status": "ready"}, nil)
		})
		docs := docsHandler(o.OpenAPISpec)
		r.Get("/docs", docs.page)
		r.Get("/docs/openapi.yaml", docs.spec)

		if o.Routes != nil {
			o.Routes(r)
		}
	})

	r.NotFound(func(w http.ResponseWriter, req *http.Request) { Error(w, req, ErrNotFound()) })
	r.MethodNotAllowed(func(w http.ResponseWriter, req *http.Request) {
		Error(w, req, &APIError{Status: 405, Code: "method_not_allowed", Message: "Método não permitido."})
	})
	return r
}
