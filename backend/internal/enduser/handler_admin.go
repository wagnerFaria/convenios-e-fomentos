package enduser

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// AdminHandler expõe a administração de usuários finais, roles, configurações e templates.
type AdminHandler struct {
	svc   *Service
	guard func(action string) func(http.Handler) http.Handler
}

// NewAdminHandler cria o handler; guard autentica o administrador e exige a permissão.
func NewAdminHandler(svc *Service, guard func(action string) func(http.Handler) http.Handler) *AdminHandler {
	return &AdminHandler{svc: svc, guard: guard}
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

// Routes registra as rotas /admin/end-user-roles, /admin/end-users, /admin/end-user-settings e /admin/email-templates.
func (h *AdminHandler) Routes(r chi.Router) {
	g := h.guard
	r.With(g("admin::plugin.enduser-roles.read")).Get("/admin/end-user-roles", h.listRoles)
	r.With(g("admin::plugin.enduser-roles.create")).Post("/admin/end-user-roles", h.createRole)
	r.With(g("admin::plugin.enduser-roles.read")).Get("/admin/end-user-roles/{id}", h.getRole)
	r.With(g("admin::plugin.enduser-roles.update")).Put("/admin/end-user-roles/{id}", h.updateRole)
	r.With(g("admin::plugin.enduser-roles.delete")).Delete("/admin/end-user-roles/{id}", h.deleteRole)

	r.With(g("admin::plugin.enduser-users.read")).Get("/admin/end-users", h.listUsers)
	r.With(g("admin::plugin.enduser-users.create")).Post("/admin/end-users", h.createUser)
	r.With(g("admin::plugin.enduser-users.read")).Get("/admin/end-users/{id}", h.getUser)
	r.With(g("admin::plugin.enduser-users.update")).Put("/admin/end-users/{id}", h.updateUser)
	r.With(g("admin::plugin.enduser-users.delete")).Delete("/admin/end-users/{id}", h.deleteUser)
	r.With(g("admin::plugin.enduser-users.update")).Post("/admin/end-users/{id}/block", h.setBlocked(true))
	r.With(g("admin::plugin.enduser-users.update")).Post("/admin/end-users/{id}/unblock", h.setBlocked(false))
	r.With(g("admin::plugin.enduser-users.update")).Put("/admin/end-users/{id}/role", h.setRole)
	r.With(g("admin::plugin.enduser-users.update")).Post("/admin/end-users/{id}/confirm", h.confirm)

	r.With(g("admin::plugin.enduser-settings.read")).Get("/admin/end-user-settings", h.getSettings)
	r.With(g("admin::plugin.enduser-settings.update")).Put("/admin/end-user-settings", h.putSettings)

	r.With(g("admin::settings.email-templates.read")).Get("/admin/email-templates", h.listTemplates)
	r.With(g("admin::settings.email-templates.read")).Get("/admin/email-templates/{key}", h.getTemplate)
	r.With(g("admin::settings.email-templates.update")).Put("/admin/email-templates/{key}", h.putTemplate)
}

// ---- roles ----

type roleInput struct {
	Name        string    `json:"name" validate:"required,max=100"`
	Description string    `json:"description" validate:"max=500"`
	Permissions *[]string `json:"permissions"` // ações api::; ausente na edição = não altera
}

func (h *AdminHandler) listRoles(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListRoles(r.Context())
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *AdminHandler) getRole(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.GetRole(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) createRole(w http.ResponseWriter, r *http.Request) {
	var in roleInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var actions []string
	if in.Permissions != nil {
		actions = *in.Permissions
	}
	v, err := h.svc.CreateRole(r.Context(), actor(r), in.Name, in.Description, actions)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *AdminHandler) updateRole(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in roleInput
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.UpdateRole(r.Context(), actor(r), id, in.Name, in.Description, in.Permissions)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) deleteRole(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.DeleteRole(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---- usuários ----

func (h *AdminHandler) listUsers(w http.ResponseWriter, r *http.Request) {
	page, err := httpapi.ParsePage(r, 25, 100)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	sort, err := httpapi.ParseSort(r, UserSortColumns, httpapi.Sort{Column: "u.created_at", Desc: true})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	q := r.URL.Query()
	f := ListUsersFilter{Q: q.Get("q"), Sort: sort, Page: page}
	if v := q.Get("roleId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			httpapi.Error(w, r, fieldIssue("roleId", "UUID inválido."))
			return
		}
		f.RoleID = &id
	}
	for _, p := range []struct {
		name string
		dst  **bool
	}{{"confirmed", &f.Confirmed}, {"blocked", &f.Blocked}} {
		switch v := q.Get(p.name); v {
		case "":
		case "true", "false":
			b := v == "true"
			*p.dst = &b
		default:
			httpapi.Error(w, r, fieldIssue(p.name, "Use true ou false."))
			return
		}
	}
	list, total, err := h.svc.ListUsers(r.Context(), f)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, page.Meta(total))
}

func (h *AdminHandler) getUser(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.GetUser(r.Context(), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) createUser(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Username  string     `json:"username" validate:"required,min=3,max=50,username"`
		Email     string     `json:"email" validate:"required,email,max=255"`
		Password  string     `json:"password" validate:"required"`
		RoleID    *uuid.UUID `json:"roleId"`
		Confirmed bool       `json:"confirmed"`
		Blocked   bool       `json:"blocked"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.CreateUser(r.Context(), actor(r), CreateUserInput{Username: in.Username, Email: in.Email, Password: in.Password, RoleID: in.RoleID, Confirmed: in.Confirmed, Blocked: in.Blocked})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, v, nil)
}

func (h *AdminHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in struct {
		Username string  `json:"username" validate:"required,min=3,max=50,username"`
		Email    string  `json:"email" validate:"required,email,max=255"`
		Password *string `json:"password"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.UpdateUser(r.Context(), actor(r), id, UpdateUserInput{Username: in.Username, Email: in.Email, Password: in.Password})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) deleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.DeleteUser(r.Context(), actor(r), id); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AdminHandler) setBlocked(blocked bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := idParam(r)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		v, err := h.svc.SetBlocked(r.Context(), actor(r), id, blocked)
		if err != nil {
			httpapi.Error(w, r, err)
			return
		}
		httpapi.JSON(w, http.StatusOK, v, nil)
	}
}

func (h *AdminHandler) setRole(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	var in struct {
		RoleID uuid.UUID `json:"roleId" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.SetRole(r.Context(), actor(r), id, in.RoleID)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) confirm(w http.ResponseWriter, r *http.Request) {
	id, err := idParam(r)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.Confirm(r.Context(), actor(r), id)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

// ---- configurações e templates ----

func (h *AdminHandler) getSettings(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetSettings(r.Context())
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, st, nil)
}

func (h *AdminHandler) putSettings(w http.ResponseWriter, r *http.Request) {
	var in Settings
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	st, err := h.svc.UpdateSettings(r.Context(), actor(r), in)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, st, nil)
}

func (h *AdminHandler) listTemplates(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListTemplates(r.Context())
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, list, nil)
}

func (h *AdminHandler) getTemplate(w http.ResponseWriter, r *http.Request) {
	v, err := h.svc.GetTemplate(r.Context(), chi.URLParam(r, "key"))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *AdminHandler) putTemplate(w http.ResponseWriter, r *http.Request) {
	var in struct {
		SenderName  string `json:"senderName" validate:"max=100"`
		SenderEmail string `json:"senderEmail" validate:"required,email"`
		ReplyTo     string `json:"replyTo"`
		Subject     string `json:"subject" validate:"required,max=300"`
		BodyHTML    string `json:"bodyHtml" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	v, err := h.svc.UpdateTemplate(r.Context(), actor(r), chi.URLParam(r, "key"), TemplateInput{SenderName: in.SenderName, SenderEmail: in.SenderEmail, ReplyTo: in.ReplyTo, Subject: in.Subject, BodyHTML: in.BodyHTML})
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}
