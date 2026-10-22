package audit

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe a consulta ao registro de auditoria.
type Handler struct{ q *db.Queries }

// NewHandler cria o handler.
func NewHandler(q *db.Queries) *Handler { return &Handler{q: q} }

// Routes registra GET /admin/audit-logs protegido pelo middleware informado
// (autenticação + permissão admin::settings.audit-logs.read).
func (h *Handler) Routes(r chi.Router, guard func(action string) func(http.Handler) http.Handler) {
	r.With(guard("admin::settings.audit-logs.read")).Get("/admin/audit-logs", h.list)
}

// LogView é um registro de auditoria exposto pela API.
type LogView struct {
	ID        uuid.UUID       `json:"id"`
	ActorType string          `json:"actorType"`
	ActorID   *string         `json:"actorId"`
	Action    string          `json:"action"`
	Entity    string          `json:"entity"`
	EntityID  *string         `json:"entityId"`
	Payload   json.RawMessage `json:"payload"`
	IP        *string         `json:"ip"`
	CreatedAt time.Time       `json:"createdAt"`
}

func optional(r *http.Request, name string) *string {
	if v := r.URL.Query().Get(name); v != "" {
		return &v
	}
	return nil
}

func optionalTime(r *http.Request, name string) (*time.Time, error) {
	v := r.URL.Query().Get(name)
	if v == "" {
		return nil, nil
	}
	t, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return nil, httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: name, Message: "Use o formato RFC 3339, ex.: 2026-01-31T00:00:00Z."}}})
	}
	return &t, nil
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	page, err := httpapi.ParsePage(r, 25, 100)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	from, err := optionalTime(r, "from")
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	to, err := optionalTime(r, "to")
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	filter := db.ListAuditLogsParams{
		ActorType: optional(r, "actorType"), ActorID: optional(r, "actorId"), Action: optional(r, "action"),
		Entity: optional(r, "entity"), EntityID: optional(r, "entityId"), FromTime: from, ToTime: to,
		PageLimit: int32(page.PageSize), PageOffset: int32(page.Offset()), //nolint:gosec // limitado por ParsePage
	}
	rows, err := h.q.ListAuditLogs(r.Context(), filter)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	total, err := h.q.CountAuditLogs(r.Context(), db.CountAuditLogsParams{
		ActorType: filter.ActorType, ActorID: filter.ActorID, Action: filter.Action, Entity: filter.Entity,
		EntityID: filter.EntityID, FromTime: from, ToTime: to})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	out := make([]LogView, 0, len(rows))
	for _, l := range rows {
		out = append(out, LogView{ID: l.ID, ActorType: l.ActorType, ActorID: l.ActorID, Action: l.Action, Entity: l.Entity,
			EntityID: l.EntityID, Payload: l.Payload, IP: l.Ip, CreatedAt: l.CreatedAt})
	}
	httpapi.JSON(w, http.StatusOK, out, page.Meta(int(total)))
}
