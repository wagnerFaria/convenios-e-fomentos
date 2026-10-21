package admin

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe o registro inicial e o convite de administradores.
type Handler struct {
	svc           *Service
	auth          *auth.AdminAuth
	engine        *authz.Engine
	ratePerMinute int
	initialReg    bool
}

// NewHandler cria o handler.
func NewHandler(svc *Service, authSvc *auth.AdminAuth, engine *authz.Engine, ratePerMinute int, initialRegistration bool) *Handler {
	return &Handler{svc: svc, auth: authSvc, engine: engine, ratePerMinute: ratePerMinute, initialReg: initialRegistration}
}

// Routes registra as rotas.
func (h *Handler) Routes(r chi.Router) {
	r.With(httpapi.RateLimitByIPAndField(h.ratePerMinute, "email")).Post("/admin/auth/register-first", h.registerFirst)
	r.With(h.auth.RequireAdmin).Get("/admin/auth/me", h.me)
	r.With(h.guard("admin::settings.users.create")).Post("/admin/users/invite", h.invite)
	h.rbacRoutes(r)
}

type adminView struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	IsActive  bool      `json:"isActive"`
}

func (h *Handler) registerFirst(w http.ResponseWriter, r *http.Request) {
	if !h.initialReg {
		httpapi.Error(w, r, &httpapi.APIError{Status: 403, Code: "setup_disabled", Message: "O registro inicial por HTTP está desabilitado. Use a CLI: cli admin:create."})
		return
	}
	var in struct {
		FirstName string `json:"firstName" validate:"required,max=100"`
		LastName  string `json:"lastName" validate:"required,max=100"`
		Email     string `json:"email" validate:"required,email"`
		Password  string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	a, err := h.svc.CreateSuperAdmin(r.Context(), CreateSuperAdminInput{FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, Password: in.Password}, true, httpapi.ClientIP(r.Context()))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, adminView{ID: a.ID, Email: a.Email, FirstName: a.FirstName, LastName: a.LastName, IsActive: a.IsActive}, nil)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.AdminFrom(r.Context())
	view, err := h.svc.Me(r.Context(), actor)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, view, nil)
}

func (h *Handler) invite(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.AdminFrom(r.Context())
	var in struct {
		FirstName    string      `json:"firstName" validate:"required,max=100"`
		LastName     string      `json:"lastName" validate:"required,max=100"`
		Email        string      `json:"email" validate:"required,email"`
		RoleIDs      []uuid.UUID `json:"roleIds" validate:"required,min=1"`
		DepartmentID *uuid.UUID  `json:"departmentId"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	res, err := h.svc.Invite(r.Context(), actor, InviteInput{FirstName: in.FirstName, LastName: in.LastName, Email: in.Email, RoleIDs: in.RoleIDs, DepartmentID: in.DepartmentID}, httpapi.ClientIP(r.Context()))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]any{
		"admin":     adminView{ID: res.Admin.ID, Email: res.Admin.Email, FirstName: res.Admin.FirstName, LastName: res.Admin.LastName, IsActive: res.Admin.IsActive},
		"inviteUrl": res.InviteURL,
		"expiresAt": res.ExpiresAt.Format(time.RFC3339),
	}, nil)
}
