package apitoken

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe a administração de tokens de API em /admin/api-tokens.
type Handler struct {
	svc   *Service
	guard func(action string) func(http.Handler) http.Handler
}

// NewHandler cria o handler; guard autentica o administrador e exige a permissão.
func NewHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, guard: guard}
}

// Routes registra as rotas.
func (h *Handler) Routes(r chi.Router) {
	g := h.guard
	r.With(g("admin::settings.api-tokens.read")).Get("/admin/api-tokens", h.list)
	r.With(g("admin::settings.api-tokens.create")).Post("/admin/api-tokens", h.create)
	r.With(g("admin::settings.api-tokens.read")).Get("/admin/api-tokens/{id}", h.get)
	r.With(g("admin::settings.api-tokens.update")).Put("/admin/api-tokens/{id}", h.update)
	r.With(g("admin::settings.api-tokens.regenerate")).Post("/admin/api-tokens/{id}/regenerate", h.regenerate)
	r.With(g("admin::settings.api-tokens.delete")).Delete("/admin/api-tokens/{id}", h.revoke)
}

func actor(r *http.Request) AdminActor {
	p, _ := auth.AdminFrom(r.Context())
	return AdminActor{ID: p.ID, IP: httpapi.ClientIP(r.Context())}
}

func idParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context(), r.URL.Query().Get("includeRevoked") == "true")
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

type tokenInput struct {
	Name        string   `json:"name" validate:"required,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Type        string   `json:"type" validate:"required"`
	Permissions []string `json:"permissions"`
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in struct {
		tokenInput
		Lifespan string `json:"lifespan" validate:"required,oneof=7d 30d 90d unlimited"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Create(r.Context(), actor(r), Input(in.tokenInput), in.Lifespan)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	// O valor do token aparece só nesta resposta.
	w.Header().Set("Cache-Control", "no-store")
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in tokenInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Update(r.Context(), actor(r), id, Input(in))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) regenerate(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Regenerate(r.Context(), actor(r), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) revoke(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.Revoke(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
