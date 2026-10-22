package document

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/admin"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// AdminHandler expõe /admin/content/documents.
type AdminHandler struct {
	svc    *Service
	engine *authz.Engine
	guard  func(action string) func(http.Handler) http.Handler
}

// NewAdminHandler cria o handler; guard autentica o administrador e exige a permissão.
func NewAdminHandler(svc *Service, engine *authz.Engine, guard func(action string) func(http.Handler) http.Handler) *AdminHandler {
	return &AdminHandler{svc: svc, engine: engine, guard: guard}
}

// Routes registra as rotas.
func (h *AdminHandler) Routes(r chi.Router) {
	g := h.guard
	r.With(g(ActionAdminRead)).Get("/admin/content/documents", h.list)
	r.With(g(ActionAdminCreate)).Post("/admin/content/documents", h.create)
	r.With(g(ActionAdminRead)).Get("/admin/content/documents/{id}", h.get)
	r.With(g(ActionAdminUpdate)).Put("/admin/content/documents/{id}", h.update)
	r.With(g(ActionAdminDelete)).Delete("/admin/content/documents/{id}", h.delete)
	r.With(g(ActionAdminPublish)).Post("/admin/content/documents/{id}/publish", h.setPublished(true))
	r.With(g(ActionAdminPublish)).Post("/admin/content/documents/{id}/unpublish", h.setPublished(false))

	// Arquivos: o navegador envia e baixa direto do armazenamento, por URLs pré-assinadas.
	r.With(g(ActionAdminFileRead)).Get("/admin/content/documents/{id}/files", h.listFiles)
	r.With(g(ActionAdminFileUpload)).Post("/admin/content/documents/{id}/files", h.createFile)
	r.With(g(ActionAdminFileUpload)).Post("/admin/content/documents/{id}/files/{fileId}/confirm", h.confirmFile)
	r.With(g(ActionAdminFileRead)).Get("/admin/content/documents/{id}/files/{fileId}/download", h.downloadFile)
	r.With(g(ActionAdminFileDelete)).Delete("/admin/content/documents/{id}/files/{fileId}", h.deleteFile)
}

func actorOf(r *http.Request) AdminActor {
	p, _ := auth.AdminFrom(r.Context())
	return AdminActor{ID: p.ID, Subject: admin.SubjectOf(p), DepartmentID: p.DepartmentID}
}

func idParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

// project aplica a permissão de campos de leitura a UM registro: cada registro pode ter
// campos liberados diferentes, conforme o filtro de cada permissão. Campos de sistema
// sempre saem. Sem permissão de leitura sobram só os campos de sistema.
func project(read authz.Decision, d Document) (map[string]any, error) {
	return Schema.Project(d, read.FieldsFor(d.Getter()))
}

// readDecision calcula a decisão de leitura do administrador (para filtrar as respostas
// de create/update/publish, que podem ser usadas por quem lê menos campos).
func (h *AdminHandler) readDecision(r *http.Request) (authz.Decision, error) {
	return h.engine.Can(r.Context(), actorOf(r).Subject, ActionAdminRead)
}

func (h *AdminHandler) render(w http.ResponseWriter, r *http.Request, status int, d Document) {
	read, err := h.readDecision(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	m, err := project(read, d)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, status, m, nil)
}

func (h *AdminHandler) list(w http.ResponseWriter, r *http.Request) {
	read, _ := admin.DecisionFrom(r.Context())
	page, err := httpapi.ParsePage(r, 25, 100)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	sort, err := httpapi.ParseSort(r, AdminSortColumns, httpapi.Sort{Column: "created_at", Desc: true})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	q := r.URL.Query()
	p := ListParams{Perm: read.Filter, Q: q.Get("q"), Status: q.Get("status"), Sort: sort, Page: page}
	if p.Status != "" && p.Status != StatusDraft && p.Status != StatusPublished {
		httpapi.Error(w, r, fieldErrors(httpapi.FieldError{Field: "status", Message: "Use rascunho ou publicado."}))
		return
	}
	if v := q.Get("department_id"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, r, fieldErrors(httpapi.FieldError{Field: "department_id", Message: "UUID inválido."}))
			return
		}
		p.DepartmentID = &id
	}
	docs, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	out := make([]map[string]any, len(docs))
	for i, d := range docs {
		if out[i], err = project(read, d); err != nil {
			httpapi.Error(w, r, err)
			return
		}
	}
	httpapi.JSON(w, http.StatusOK, out, page.Meta(total))
}

func (h *AdminHandler) get(w http.ResponseWriter, r *http.Request) {
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
	read, _ := admin.DecisionFrom(r.Context())
	if !read.AllowsRecord(d.Getter()) {
		httpapi.Error(w, r, httpapi.ErrForbidden())
		return
	}
	m, err := project(read, d)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, m, nil)
}

func schemaHas(k string) bool { return Schema.Has(k) }

func (h *AdminHandler) create(w http.ResponseWriter, r *http.Request) {
	dec, _ := admin.DecisionFrom(r.Context())
	p, err := decodePatch(r, schemaHas)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	d, err := h.svc.CreateByAdmin(r.Context(), actorOf(r), p, dec.Fields)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	h.render(w, r, http.StatusCreated, d)
}

func (h *AdminHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	p, err := decodePatch(r, schemaHas)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	d, err := h.svc.UpdateByAdmin(r.Context(), actorOf(r), id, p, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	h.render(w, r, http.StatusOK, d)
}

func (h *AdminHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	del, _ := admin.DecisionFrom(r.Context())
	if err := h.svc.DeleteByAdmin(r.Context(), id, del); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) setPublished(publish bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(r)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		pub, _ := admin.DecisionFrom(r.Context())
		d, err := h.svc.SetPublished(r.Context(), actorOf(r), id, publish, pub)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		h.render(w, r, http.StatusOK, d)
	}
}

func fileIDParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "fileId"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

// ids lê {id} e {fileId} da rota.
func ids(r *http.Request) (doc, file uuid.UUID, err error) {
	if doc, err = idParam(r); err != nil {
		return
	}
	file, err = fileIDParam(r)
	return
}

func (h *AdminHandler) listFiles(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	files, err := h.svc.ListFilesByAdmin(r.Context(), id, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, files, nil)
}

func (h *AdminHandler) createFile(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in FileInput
	if err := decodeStrict(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	f, up, err := h.svc.CreateFile(r.Context(), actorOf(r), id, in, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]any{"file": f, "upload": up}, nil)
}

func (h *AdminHandler) confirmFile(w http.ResponseWriter, r *http.Request) {
	docID, fileID, err := ids(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	f, err := h.svc.ConfirmFile(r.Context(), docID, fileID, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, f, nil)
}

func (h *AdminHandler) downloadFile(w http.ResponseWriter, r *http.Request) {
	docID, fileID, err := ids(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	u, exp, err := h.svc.DownloadByAdmin(r.Context(), docID, fileID, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, downloadLink{URL: u, ExpiresAt: exp}, nil)
}

func (h *AdminHandler) deleteFile(w http.ResponseWriter, r *http.Request) {
	docID, fileID, err := ids(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	if err := h.svc.DeleteFile(r.Context(), docID, fileID, dec); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
