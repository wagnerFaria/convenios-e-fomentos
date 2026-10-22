package department

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe a administração de departamentos em /admin/departments.
type Handler struct {
	svc      *Service
	guard    func(action string) func(http.Handler) http.Handler
	guardAny func(actions ...string) func(http.Handler) http.Handler
}

// NewHandler cria o handler. guard autentica o administrador e exige a permissão; guardAny exige
// uma entre várias.
func NewHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler, guardAny func(actions ...string) func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, guard: guard, guardAny: guardAny}
}

// Routes registra as rotas. A leitura também é aberta a quem administra administradores, porque
// a lista alimenta o seletor de departamento deles.
func (h *Handler) Routes(r chi.Router) {
	read := h.guardAny("admin::settings.departments.read", "admin::settings.users.read")
	r.With(read).Get("/admin/departments", h.list)
	r.With(h.guard("admin::settings.departments.create")).Post("/admin/departments", h.create)
	r.With(read).Get("/admin/departments/{id}", h.get)
	r.With(h.guard("admin::settings.departments.update")).Put("/admin/departments/{id}", h.update)
	r.With(h.guard("admin::settings.departments.delete")).Delete("/admin/departments/{id}", h.delete)
}

func actor(r *http.Request) Actor {
	p, _ := auth.AdminFrom(r.Context())
	return Actor{ID: p.ID, IP: httpapi.ClientIP(r.Context())}
}

func idParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

type input struct {
	Name string `json:"name" validate:"required,max=100"`
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
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

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var in input
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Create(r.Context(), actor(r), in.Name)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in input
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Update(r.Context(), actor(r), id, in.Name)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.Delete(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
