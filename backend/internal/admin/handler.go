package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe o pré-cadastro e a consulta do administrador autenticado.
type Handler struct {
	svc    *Service
	auth   *auth.AdminAuth
	engine *authz.Engine
}

// NewHandler cria o handler.
func NewHandler(svc *Service, authSvc *auth.AdminAuth, engine *authz.Engine) *Handler {
	return &Handler{svc: svc, auth: authSvc, engine: engine}
}

// Routes registra as rotas.
func (h *Handler) Routes(r chi.Router) {
	r.With(h.auth.RequireAdmin).Get("/admin/auth/me", h.me)
	r.With(h.guard("admin::settings.users.create")).Post("/admin/users", h.register)
	h.rbacRoutes(r)
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

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	actor, _ := auth.AdminFrom(r.Context())
	var in struct {
		CPF          string      `json:"cpf" validate:"required"`
		RoleIDs      []uuid.UUID `json:"roleIds" validate:"required,min=1"`
		DepartmentID *uuid.UUID  `json:"departmentId"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Register(r.Context(), actor, RegisterInput{CPF: in.CPF, RoleIDs: in.RoleIDs, DepartmentID: in.DepartmentID}, httpapi.ClientIP(r.Context()))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}
