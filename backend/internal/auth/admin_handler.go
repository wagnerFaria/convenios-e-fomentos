package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

type principalKey struct{}

// AdminFrom devolve o administrador autenticado no contexto.
func AdminFrom(ctx context.Context) (*AdminPrincipal, bool) {
	p, ok := ctx.Value(principalKey{}).(*AdminPrincipal)
	return p, ok
}

// RequireAdmin exige um access token de administrador válido (audience admin).
func (s *AdminAuth) RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		token, found := strings.CutPrefix(h, "Bearer ")
		if !found || token == "" {
			httpapi.Error(w, r, httpapi.ErrUnauthorized())
			return
		}
		p, err := s.Authenticate(r.Context(), token)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), principalKey{}, p)))
	})
}

// AdminHandler expõe as rotas /admin/auth.
type AdminHandler struct {
	svc           *AdminAuth
	ratePerMinute int
}

// NewAdminHandler cria o handler.
func NewAdminHandler(svc *AdminAuth, ratePerMinute int) *AdminHandler {
	return &AdminHandler{svc: svc, ratePerMinute: ratePerMinute}
}

func meta(r *http.Request) SessionMeta {
	return SessionMeta{UserAgent: r.UserAgent(), IP: httpapi.ClientIP(r.Context())}
}

// Routes registra as rotas. O registro inicial é montado pelo pacote admin.
func (h *AdminHandler) Routes(r chi.Router) {
	r.Route("/admin/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpapi.RateLimitByIPAndField(h.ratePerMinute, "email"))
			r.Post("/login", h.login)
			r.Post("/forgot-password", h.forgotPassword)
		})
		r.Group(func(r chi.Router) {
			r.Use(httpapi.RateLimitByIPAndField(h.ratePerMinute, ""))
			r.Post("/refresh", h.refresh)
			r.Post("/reset-password", h.resetPassword)
			r.Post("/accept-invite", h.acceptInvite)
		})
		r.Group(func(r chi.Router) {
			r.Use(h.svc.RequireAdmin)
			r.Post("/logout", h.logout)
			r.Get("/sessions", h.listSessions)
			r.Delete("/sessions", h.revokeAll)
			r.Delete("/sessions/{id}", h.revokeOne)
		})
	})
}

func (h *AdminHandler) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	pair, err := h.svc.Login(r.Context(), in.Email, in.Password, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, pair, nil)
}

func (h *AdminHandler) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	pair, err := h.svc.Refresh(r.Context(), in.RefreshToken, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, pair, nil)
}

func (h *AdminHandler) logout(w http.ResponseWriter, r *http.Request) {
	p, _ := AdminFrom(r.Context())
	if err := h.svc.Logout(r.Context(), p); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	p, _ := AdminFrom(r.Context())
	list, err := h.svc.ListSessions(r.Context(), p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *AdminHandler) revokeOne(w http.ResponseWriter, r *http.Request) {
	p, _ := AdminFrom(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpapi.Error(w, r, httpapi.ErrNotFound())
		return
	}
	if err := h.svc.RevokeSession(r.Context(), p, id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) revokeAll(w http.ResponseWriter, r *http.Request) {
	p, _ := AdminFrom(r.Context())
	if err := h.svc.RevokeAllSessions(r.Context(), p, r.URL.Query().Get("keepCurrent") == "true"); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.ForgotPassword(r.Context(), in.Email); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.ResetPassword(r.Context(), in.Token, in.Password); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) acceptInvite(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	pair, err := h.svc.AcceptInvite(r.Context(), in.Token, in.Password, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, pair, nil)
}
