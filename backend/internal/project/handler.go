package project

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/admin"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Limits são os limites de requisições por minuto, por usuário, das rotas mais caras. Zero desliga o limite.
type Limits struct {
	// Upload vale para o pedido e a confirmação de envio de arquivo (a confirmação lê o arquivo inteiro).
	Upload int
	// Heavy vale para envio, reenvio, duplicação e comprovante.
	Heavy int
}

// DefaultHeavyLimit é o limite padrão das rotas pesadas: uma pessoa não envia nem duplica mais de 20 vezes por minuto.
const DefaultHeavyLimit = 20

// Handler expõe /admin/projects e /admin/proponent-profile.
type Handler struct {
	svc    *Service
	guard  func(action string) func(http.Handler) http.Handler
	limits Limits
}

// NewHandler cria o handler; guard autentica o administrador e exige a permissão.
func NewHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler, limits Limits) *Handler {
	return &Handler{svc: svc, guard: guard, limits: limits}
}

// perUser limita as requisições por minuto de cada administrador (depois da autenticação). n <= 0 desliga.
func perUser(n int) func(http.Handler) http.Handler {
	if n <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}
	return httprate.LimitBy(n, time.Minute,
		func(r *http.Request) (string, error) {
			p, _ := auth.AdminFrom(r.Context())
			if p == nil {
				return httpapi.ClientIP(r.Context()), nil
			}
			return p.ID.String(), nil
		},
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) { httpapi.Error(w, r, httpapi.ErrTooManyRequests()) }),
	)
}

// Routes registra as rotas.
func (h *Handler) Routes(r chi.Router) {
	g := h.guard
	r.With(g(ActionProfileRead)).Get("/admin/proponent-profile", h.getProfile)
	r.With(g(ActionProfileUpdate)).Put("/admin/proponent-profile", h.putProfile)

	r.With(g(ActionRead)).Get("/admin/projects", h.list)
	r.With(g(ActionCreate)).Post("/admin/projects", h.create)
	r.With(g(ActionRead)).Get("/admin/projects/{id}", h.get)
	r.With(g(ActionDelete)).Delete("/admin/projects/{id}", h.delete)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/proponent", h.putProponent)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/institution", h.putInstitution)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/proposal", h.putProposal)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/funding", h.putFunding)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/document-options", h.putDocumentOptions)

	heavy, upload := perUser(h.limits.Heavy), perUser(h.limits.Upload)
	r.With(g(ActionSubmit), heavy).Post("/admin/projects/{id}/submit", h.submit)
	r.With(g(ActionSubmit), heavy).Post("/admin/projects/{id}/resubmit", h.resubmit)
	r.With(g(ActionWithdraw)).Post("/admin/projects/{id}/withdraw", h.withdraw)
	r.With(g(ActionRequestComplement)).Post("/admin/projects/{id}/complement-requests", h.requestComplement)
	r.With(g(ActionArchive)).Post("/admin/projects/{id}/archive", h.archive)
	r.With(g(ActionDeny)).Post("/admin/projects/{id}/deny", h.deny)
	r.With(g(ActionCreate), heavy).Post("/admin/projects/{id}/duplicate", h.duplicate)
	r.With(g(ActionRead), heavy).Get("/admin/projects/{id}/receipt", h.receipt)

	r.With(g(ActionShare)).Get("/admin/projects/{id}/members", h.members)
	r.With(g(ActionShare), heavy).Post("/admin/projects/{id}/members", h.invite)
	r.With(g(ActionShare)).Put("/admin/projects/{id}/members/{memberId}", h.updateMember)
	r.With(g(ActionShare)).Delete("/admin/projects/{id}/members/{memberId}", h.revokeMember)

	r.With(g(ActionRead)).Get("/admin/projects/{id}/documents", h.checklist)
	r.With(g(ActionUpdate)).Put("/admin/projects/{id}/documents/{code}/justification", h.putJustification)
	r.With(g(ActionUpdate), upload).Post("/admin/projects/{id}/documents/{code}/files", h.requestUpload)
	r.With(g(ActionUpdate), upload).Post("/admin/projects/{id}/documents/files/{fileId}/confirm", h.confirmUpload)
	r.With(g(ActionRead)).Get("/admin/projects/{id}/documents/files/{fileId}/download", h.download)
	r.With(g(ActionUpdate)).Delete("/admin/projects/{id}/documents/files/{fileId}", h.deleteFile)
}

func actorOf(r *http.Request) Actor {
	p, _ := auth.AdminFrom(r.Context())
	return Actor{ID: p.ID, CPF: p.CPF, Name: strings.TrimSpace(p.FirstName + " " + p.LastName), IP: httpapi.ClientIP(r.Context())}
}

func idParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

func (h *Handler) getProfile(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.GetProfile(r.Context(), actorOf(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) putProfile(w http.ResponseWriter, r *http.Request) {
	var in Contact
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.SaveProfile(r.Context(), actorOf(r), in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	read, _ := admin.DecisionFrom(r.Context())
	page, err := httpapi.ParsePage(r, 25, 100)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	sort, err := httpapi.ParseSort(r, SortColumns, httpapi.Sort{Column: "updated_at", Desc: true})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	q := r.URL.Query()
	p := ListParams{Perm: read.Filter, Actor: actorOf(r), Q: q.Get("q"), Status: q.Get("status"), Sort: sort, Page: page}
	if p.Status != "" && p.Status != StatusDraft && p.Status != StatusSubmitted && p.Status != StatusArchived {
		httpapi.Error(w, r, httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: "status", Message: "Valores aceitos: draft, submitted, archived."}}}))
		return
	}
	list, total, err := h.svc.List(r.Context(), p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, page.Meta(total))
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.Create(r.Context(), actorOf(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	read, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.Get(r.Context(), actorOf(r), id, read)
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
	del, _ := admin.DecisionFrom(r.Context())
	if err := h.svc.Delete(r.Context(), actorOf(r), id, del); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Cada seção é gravada inteira, junto com a versão que o cliente leu.
type proponentBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
	ProponentSection
}

type institutionBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
	InstitutionSection
}

type proposalBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
	ProposalSection
}

func (h *Handler) putProponent(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in proponentBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SaveProponent(r.Context(), actorOf(r), id, in.Version, in.ProponentSection, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) putInstitution(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in institutionBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SaveInstitution(r.Context(), actorOf(r), id, in.Version, in.InstitutionSection, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) putProposal(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in proposalBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SaveProposal(r.Context(), actorOf(r), id, in.Version, in.ProposalSection, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

type fundingBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
	FundingSection
}

func (h *Handler) putFunding(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in fundingBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SaveFunding(r.Context(), actorOf(r), id, in.Version, in.FundingSection, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

type documentOptionsBody struct {
	Version int32 `json:"version" validate:"required,min=1"`
	DocumentOptions
}

func (h *Handler) putDocumentOptions(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in documentOptionsBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SaveDocumentOptions(r.Context(), actorOf(r), id, in.Version, in.DocumentOptions, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) checklist(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	read, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.Checklist(r.Context(), actorOf(r), id, read)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

type justificationBody struct {
	AmendmentID string `json:"amendmentId"`
	Text        string `json:"text"`
}

func (h *Handler) putJustification(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in justificationBody
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.SetJustification(r.Context(), actorOf(r), id, chi.URLParam(r, "code"), in.AmendmentID, in.Text, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) requestUpload(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in FileInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	file, up, err := h.svc.RequestUpload(r.Context(), actorOf(r), id, chi.URLParam(r, "code"), in, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, map[string]any{"file": file, "upload": up}, nil)
}

func fileIDParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "fileId"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

func (h *Handler) confirmUpload(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	fileID, err2 := fileIDParam(r)
	if err != nil || err2 != nil {
		httpapi.Error(w, r, httpapi.ErrNotFound())
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.ConfirmUpload(r.Context(), id, fileID, upd)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) download(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	fileID, err2 := fileIDParam(r)
	if err != nil || err2 != nil {
		httpapi.Error(w, r, httpapi.ErrNotFound())
		return
	}
	read, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.Download(r.Context(), actorOf(r), id, fileID, read)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) deleteFile(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	fileID, err2 := fileIDParam(r)
	if err != nil || err2 != nil {
		httpapi.Error(w, r, httpapi.ErrNotFound())
		return
	}
	upd, _ := admin.DecisionFrom(r.Context())
	if err := h.svc.DeleteFile(r.Context(), id, fileID, upd); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) submit(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in SubmitInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	sub, _ := admin.DecisionFrom(r.Context())
	v, err := h.svc.Submit(r.Context(), actorOf(r), id, in, sub)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

// receipt devolve o comprovante de recebimento em PDF (não é o envelope JSON padrão).
func (h *Handler) receipt(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	read, _ := admin.DecisionFrom(r.Context())
	pdf, name, err := h.svc.Receipt(r.Context(), actorOf(r), id, read)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", `attachment; filename="`+name+`"`)
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdf) //nolint:gosec // PDF gerado pelo servidor, servido como application/pdf (não é HTML)
}

// decode lê o corpo e devolve a decisão de permissão da rota; num único passo para os handlers do ciclo de vida.
func decode[T any](w http.ResponseWriter, r *http.Request) (T, uuid.UUID, authz.Decision, bool) {
	var in T
	id, err := idParam(r)
	if err == nil {
		err = httpapi.Decode(r, &in)
	}
	if err != nil {
		httpapi.Error(w, r, err)
		return in, uuid.Nil, authz.Decision{}, false
	}
	dec, _ := admin.DecisionFrom(r.Context())
	return in, id, dec, true
}

func respond(w http.ResponseWriter, r *http.Request, v *Project, err error) {
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) resubmit(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[SubmitInput](w, r)
	if !ok {
		return
	}
	v, err := h.svc.Resubmit(r.Context(), actorOf(r), id, in, dec)
	respond(w, r, v, err)
}

type noteBody struct {
	Note string `json:"note"`
}

func (h *Handler) withdraw(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[noteBody](w, r)
	if !ok {
		return
	}
	v, err := h.svc.Withdraw(r.Context(), actorOf(r), id, in.Note, dec)
	respond(w, r, v, err)
}

func (h *Handler) deny(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[noteBody](w, r)
	if !ok {
		return
	}
	v, err := h.svc.Deny(r.Context(), actorOf(r), id, in.Note, dec)
	respond(w, r, v, err)
}

func (h *Handler) archive(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[ArchiveInput](w, r)
	if !ok {
		return
	}
	v, err := h.svc.Archive(r.Context(), actorOf(r), id, in, dec)
	respond(w, r, v, err)
}

func (h *Handler) requestComplement(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[ComplementInput](w, r)
	if !ok {
		return
	}
	v, err := h.svc.RequestComplement(r.Context(), actorOf(r), id, in, dec)
	respond(w, r, v, err)
}

func (h *Handler) duplicate(w http.ResponseWriter, r *http.Request) {
	in, id, _, ok := decode[DuplicateInput](w, r)
	if !ok {
		return
	}
	v, err := h.svc.Duplicate(r.Context(), actorOf(r), id, in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *Handler) members(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	dec, _ := admin.DecisionFrom(r.Context())
	list, err := h.svc.Members(r.Context(), id, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *Handler) invite(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[MemberInput](w, r)
	if !ok {
		return
	}
	m, err := h.svc.Invite(r.Context(), actorOf(r), id, in, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, m, nil)
}

func memberParam(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "memberId"))
	if err != nil {
		return uuid.Nil, httpapi.ErrNotFound()
	}
	return id, nil
}

func (h *Handler) updateMember(w http.ResponseWriter, r *http.Request) {
	in, id, dec, ok := decode[MemberInput](w, r)
	if !ok {
		return
	}
	memberID, err := memberParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	m, err := h.svc.UpdateMember(r.Context(), actorOf(r), id, memberID, in, dec)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, m, nil)
}

func (h *Handler) revokeMember(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err == nil {
		var memberID uuid.UUID
		if memberID, err = memberParam(r); err == nil {
			dec, _ := admin.DecisionFrom(r.Context())
			err = h.svc.RevokeMember(r.Context(), actorOf(r), id, memberID, dec)
		}
	}
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
