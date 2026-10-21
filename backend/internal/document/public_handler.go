package document

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/enduser"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/policy"
)

// PublicHandler expõe /api/documents. Cada rota combina: identificação do sujeito
// (usuário final, token de API ou anônimo), permissão da ação e, quando preciso, policy.
type PublicHandler struct {
	svc *Service
	id  *enduser.Identifier
}

// NewPublicHandler cria o handler.
func NewPublicHandler(svc *Service, id *enduser.Identifier) *PublicHandler {
	return &PublicHandler{svc: svc, id: id}
}

// Routes registra as rotas públicas.
func (h *PublicHandler) Routes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(h.id.Identify)
		r.With(h.id.RequireAction(ActionAPIFind)).Get("/api/documents", h.find)
		r.With(h.id.RequireAction(ActionAPIFindOne)).Get("/api/documents/{id}", h.findOne)
		r.With(h.id.RequireAction(ActionAPICreate)).Post("/api/documents", h.create)
		// A policy is-owner roda depois da permissão da ação.
		r.With(h.id.RequireAction(ActionAPIUpdate), policy.Require(h.isOwner())).Put("/api/documents/{id}", h.update)
		r.With(h.id.RequireAction(ActionAPIDelete), policy.Require(h.isOwner())).Delete("/api/documents/{id}", h.delete)
	})
}

// isOwner: só o dono altera. Rascunho de outra pessoa responde 404, para não revelar que existe.
func (h *PublicHandler) isOwner() policy.Policy {
	return policy.IsOwner(func(ctx context.Context, r *http.Request) (uuid.UUID, error) {
		id, err := idParam(r)
		if err != nil {
			return uuid.Nil, err
		}
		d, err := h.svc.Get(ctx, id)
		if err != nil {
			return uuid.Nil, err
		}
		owner := uuid.Nil
		if d.OwnerUserID != nil {
			owner = *d.OwnerUserID
		}
		if subj, ok := authz.SubjectFrom(ctx); ok && d.Status != StatusPublished && owner != subj.ID {
			return uuid.Nil, httpapi.ErrNotFound()
		}
		return owner, nil
	})
}

func (h *PublicHandler) find(w http.ResponseWriter, r *http.Request) {
	page, err := httpapi.ParsePage(r, 25, 100)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	sort, err := httpapi.ParseSort(r, PublicSortColumns, httpapi.Sort{Column: "published_at", Desc: true})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	docs, total, err := h.svc.List(r.Context(), ListParams{Perm: authz.FilterTrue(), OnlyPublished: true, Q: r.URL.Query().Get("q"), Sort: sort, Page: page})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	out := make([]PublicDocument, len(docs))
	for i, d := range docs {
		out[i] = publicOf(d)
	}
	httpapi.JSON(w, http.StatusOK, out, page.Meta(total))
}

func (h *PublicHandler) findOne(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	d, err := h.svc.Get(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if d.Status != StatusPublished {
		httpapi.Error(w, r, httpapi.ErrNotFound())
		return
	}
	httpapi.JSON(w, http.StatusOK, publicOf(d), nil)
}

// publicKeys: a API pública só grava título e conteúdo. Qualquer outro campo (inclusive
// observacao_interna, status e os de sistema) é recusado.
func publicKeys(k string) bool { return k == "titulo" || k == "conteudo" }

func (h *PublicHandler) create(w http.ResponseWriter, r *http.Request) {
	p, err := decodePatch(r, publicKeys)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var owner *uuid.UUID
	if u, ok := enduser.PrincipalFrom(r.Context()); ok {
		owner = &u.ID
	}
	d, err := h.svc.CreateByUser(r.Context(), owner, p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, publicOf(d), nil)
}

func (h *PublicHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	p, err := decodePatch(r, publicKeys)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	d, err := h.svc.UpdateByUser(r.Context(), id, p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, publicOf(d), nil)
}

func (h *PublicHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.DeleteByUser(r.Context(), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
