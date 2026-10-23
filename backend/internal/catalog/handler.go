package catalog

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe os catálogos em /admin/catalog/....
type Handler struct {
	svc   *Service
	guard func(action string) func(http.Handler) http.Handler
}

// NewHandler cria o handler. guard autentica o administrador e exige a permissão.
func NewHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, guard: guard}
}

// Routes registra as rotas.
func (h *Handler) Routes(r chi.Router) {
	r.With(h.guard(ActionEntityTypesRead)).Get("/admin/catalog/entity-types", h.listEntityTypes)
	r.With(h.guard(ActionEntityTypesCreate)).Post("/admin/catalog/entity-types", h.createEntityType)
	r.With(h.guard(ActionEntityTypesRead)).Get("/admin/catalog/entity-types/{id}", h.getEntityType)
	r.With(h.guard(ActionEntityTypesUpdate)).Put("/admin/catalog/entity-types/{id}", h.updateEntityType)
	r.With(h.guard(ActionEntityTypesDelete)).Delete("/admin/catalog/entity-types/{id}", h.deleteEntityType)

	r.With(h.guard(ActionMunicipalitiesRead)).Get("/admin/catalog/municipalities", h.listMunicipalities)
	r.With(h.guard(ActionMunicipalitiesRead)).Get("/admin/catalog/municipalities/{ibgeCode}", h.getMunicipality)

	r.With(h.guard(ActionParliamentariansRead)).Get("/admin/catalog/parliamentarians", h.listParliamentarians)
	r.With(h.guard(ActionParliamentariansCreate)).Post("/admin/catalog/parliamentarians", h.createParliamentarian)
	r.With(h.guard(ActionParliamentariansRead)).Get("/admin/catalog/parliamentarians/{id}", h.getParliamentarian)
	r.With(h.guard(ActionParliamentariansUpdate)).Put("/admin/catalog/parliamentarians/{id}", h.updateParliamentarian)
	r.With(h.guard(ActionParliamentariansDelete)).Delete("/admin/catalog/parliamentarians/{id}", h.deleteParliamentarian)
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

// boolQuery lê um filtro booleano opcional (true/false); vazio vale false.
func boolQuery(r *http.Request, name string) (bool, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return false, nil
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false, fieldError(name, "Use true ou false.")
	}
	return b, nil
}

// ---- Tipos de entidade ----

func (h *Handler) listEntityTypes(w http.ResponseWriter, r *http.Request) {
	onlyActive, err := boolQuery(r, "active")
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	list, err := h.svc.ListEntityTypes(r.Context(), onlyActive)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *Handler) getEntityType(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.GetEntityType(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) createEntityType(w http.ResponseWriter, r *http.Request) {
	var in EntityTypeInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.CreateEntityType(r.Context(), actor(r), in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) updateEntityType(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in EntityTypeInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.UpdateEntityType(r.Context(), actor(r), id, in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) deleteEntityType(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.DeleteEntityType(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- Municípios ----

func (h *Handler) listMunicipalities(w http.ResponseWriter, r *http.Request) {
	page, err := httpapi.ParsePage(r, 50, 500)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	q := r.URL.Query()
	list, total, err := h.svc.ListMunicipalities(r.Context(), q.Get("uf"), q.Get("q"), page)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, page.Meta(total))
}

func (h *Handler) getMunicipality(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.GetMunicipality(r.Context(), chi.URLParam(r, "ibgeCode"))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

// ---- Parlamentares ----

func (h *Handler) listParliamentarians(w http.ResponseWriter, r *http.Request) {
	onlyActive, err := boolQuery(r, "active")
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	q := r.URL.Query()
	list, err := h.svc.ListParliamentarians(r.Context(), ParliamentarianFilter{Chamber: q.Get("chamber"), OnlyActive: onlyActive, Search: q.Get("q")})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *Handler) getParliamentarian(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.GetParliamentarian(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) createParliamentarian(w http.ResponseWriter, r *http.Request) {
	var in ParliamentarianInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.CreateParliamentarian(r.Context(), actor(r), in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) updateParliamentarian(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in ParliamentarianInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.UpdateParliamentarian(r.Context(), actor(r), id, in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) deleteParliamentarian(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.DeleteParliamentarian(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
