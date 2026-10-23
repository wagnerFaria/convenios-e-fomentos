package rules

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe a administração dos parâmetros em /admin/rule-parameters.
type Handler struct {
	svc   *Service
	guard func(action string) func(http.Handler) http.Handler
}

// NewHandler cria o handler. guard autentica o administrador e exige a permissão.
func NewHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, guard: guard}
}

// Routes registra as rotas. "keys" e "effective" vêm antes de "{id}" por clareza; o roteador distingue pelo literal.
func (h *Handler) Routes(r chi.Router) {
	read := h.guard(ActionRead)
	r.With(read).Get("/admin/rule-parameters", h.list)
	r.With(read).Get("/admin/rule-parameters/keys", h.keys)
	r.With(read).Get("/admin/rule-parameters/effective", h.effective)
	r.With(h.guard(ActionCreate)).Post("/admin/rule-parameters", h.create)
	r.With(read).Get("/admin/rule-parameters/{id}", h.get)
	r.With(h.guard(ActionUpdate)).Put("/admin/rule-parameters/{id}", h.update)
	r.With(h.guard(ActionDelete)).Delete("/admin/rule-parameters/{id}", h.delete)
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

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key != "" {
		if _, ok := SpecFor(key); !ok {
			httpapi.Error(w, r, fieldError("key", "Chave de parâmetro desconhecida."))
			return
		}
	}
	list, err := h.svc.List(r.Context(), key)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *Handler) keys(w http.ResponseWriter, _ *http.Request) {
	httpapi.JSON(w, http.StatusOK, h.svc.Keys(), nil)
}

func (h *Handler) effective(w http.ResponseWriter, r *http.Request) {
	date := h.svc.Today()
	if v := r.URL.Query().Get("date"); v != "" {
		d, err := time.Parse(time.DateOnly, v)
		if err != nil {
			httpapi.Error(w, r, fieldError("date", "Informe uma data no formato AAAA-MM-DD."))
			return
		}
		date = d
	}
	list, err := h.svc.Effective(r.Context(), date)
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
	var in Input
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Create(r.Context(), actor(r), in)
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
	var in Input
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Update(r.Context(), actor(r), id, in)
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
