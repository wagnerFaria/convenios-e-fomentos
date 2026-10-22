package enduser

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Handler expõe as rotas públicas /api/auth e /api/users/me.
type Handler struct {
	svc           *Service
	id            *Identifier
	ratePerMinute int
}

// NewHandler cria o handler público.
func NewHandler(svc *Service, id *Identifier, ratePerMinute int) *Handler {
	return &Handler{svc: svc, id: id, ratePerMinute: ratePerMinute}
}

func meta(r *http.Request) auth.SessionMeta {
	return auth.SessionMeta{UserAgent: r.UserAgent(), IP: httpapi.ClientIP(r.Context())}
}

// Routes registra as rotas. As de autenticação são públicas e limitadas por taxa.
func (h *Handler) Routes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(httpapi.RateLimitByIPAndField(h.ratePerMinute, "identifier"))
			r.Post("/local", h.login)
		})
		r.Group(func(r chi.Router) {
			r.Use(httpapi.RateLimitByIPAndField(h.ratePerMinute, "email"))
			r.Post("/local/register", h.register)
			r.Post("/forgot-password", h.forgotPassword)
			r.Post("/send-email-confirmation", h.sendEmailConfirmation)
		})
		r.Group(func(r chi.Router) {
			r.Use(httpapi.RateLimitByIPAndField(h.ratePerMinute, ""))
			r.Post("/refresh", h.refresh)
			r.Post("/reset-password", h.resetPassword)
			r.Get("/email-confirmation", h.confirmEmail)
		})
		r.Group(func(r chi.Router) {
			r.Use(h.id.Identify, RequireUser)
			r.Post("/logout", h.logout)
			r.Post("/change-password", h.changePassword)
		})
	})
	r.With(h.id.Identify, RequireUser).Get("/users/me", h.me)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Identifier string `json:"identifier" validate:"required,max=255"`
		Password   string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	res, err := h.svc.Login(r.Context(), in.Identifier, in.Password, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, res, nil)
}

// decodeRegister lê o corpo do cadastro. Além de username, email e password,
// aceita apenas os campos extras habilitados na configuração; qualquer outro → 422.
func decodeRegister(r *http.Request, allowed []string) (RegisterInput, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return RegisterInput{}, &httpapi.APIError{Status: 413, Code: "payload_too_large", Message: "O corpo da requisição excede o limite permitido."}
		}
		return RegisterInput{}, httpapi.ErrBadRequest("Corpo da requisição inválido.")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return RegisterInput{}, httpapi.ErrBadRequest("Corpo da requisição inválido: esperado um objeto JSON.")
	}
	var base struct {
		Username string `json:"username" validate:"required,min=3,max=50,username"`
		Email    string `json:"email" validate:"required,email,max=255"`
		Password string `json:"password" validate:"required"`
	}
	known := map[string]bool{"username": true, "email": true, "password": true}
	baseJSON := map[string]json.RawMessage{}
	extra := map[string]any{}
	var issues []httpapi.FieldError
	for k, v := range fields {
		switch {
		case known[k]:
			baseJSON[k] = v
		case slices.Contains(allowed, k):
			var val any
			if err := json.Unmarshal(v, &val); err != nil {
				issues = append(issues, httpapi.FieldError{Field: k, Message: "Valor inválido."})
				continue
			}
			switch t := val.(type) {
			case string:
				if len(t) > maxExtraValueBytes {
					issues = append(issues, httpapi.FieldError{Field: k, Message: "Valor longo demais."})
					continue
				}
			case float64, bool, nil:
			default:
				issues = append(issues, httpapi.FieldError{Field: k, Message: "Apenas texto, número ou booleano."})
				continue
			}
			extra[k] = val
		default:
			issues = append(issues, httpapi.FieldError{Field: k, Message: "Campo não permitido no cadastro."})
		}
	}
	b, _ := json.Marshal(baseJSON)
	if err := json.NewDecoder(bytes.NewReader(b)).Decode(&base); err != nil {
		return RegisterInput{}, httpapi.ErrBadRequest("Corpo da requisição inválido: " + err.Error())
	}
	if err := httpapi.Validate(&base); err != nil {
		var apiErr *httpapi.APIError
		if errors.As(err, &apiErr) {
			if d, ok := apiErr.Details.(map[string]any); ok {
				if fs, ok := d["fields"].([]httpapi.FieldError); ok {
					issues = append(fs, issues...)
				}
			}
		}
	}
	if len(issues) > 0 {
		slices.SortFunc(issues, func(a, b httpapi.FieldError) int { return strings.Compare(a.Field, b.Field) })
		return RegisterInput{}, httpapi.ErrValidation(map[string]any{"fields": issues})
	}
	return RegisterInput{Username: base.Username, Email: base.Email, Password: base.Password, Extra: extra}, nil
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.GetSettings(r.Context())
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if !st.RegistrationEnabled {
		httpapi.Error(w, r, errRegistrationDisabled)
		return
	}
	in, err := decodeRegister(r, st.RegisterAllowedFields)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	res, err := h.svc.Register(r.Context(), in, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusCreated, res, nil)
}

func (h *Handler) refresh(w http.ResponseWriter, r *http.Request) {
	var in struct {
		RefreshToken string `json:"refreshToken" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	pair, err := h.svc.Refresh(r.Context(), in.RefreshToken, meta(r))
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, pair, nil)
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFrom(r.Context())
	if err := h.svc.Logout(r.Context(), p); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	p, _ := PrincipalFrom(r.Context())
	v, err := h.svc.Me(r.Context(), p)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	httpapi.JSON(w, http.StatusOK, v, nil)
}

func (h *Handler) confirmEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("confirmation")
	if token == "" {
		httpapi.Error(w, r, httpapi.ErrBadRequest("Informe o parâmetro confirmation."))
		return
	}
	redirect, err := h.svc.ConfirmEmail(r.Context(), token)
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if redirect != "" {
		http.Redirect(w, r, redirect, http.StatusFound)
		return
	}
	httpapi.JSON(w, http.StatusOK, map[string]bool{"confirmed": true}, nil)
}

func emailOnly(r *http.Request) (string, error) {
	var in struct {
		Email string `json:"email" validate:"required,email"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		return "", err
	}
	return in.Email, nil
}

func (h *Handler) sendEmailConfirmation(w http.ResponseWriter, r *http.Request) {
	email, err := emailOnly(r)
	if err == nil {
		err = h.svc.SendEmailConfirmation(r.Context(), email)
	}
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) forgotPassword(w http.ResponseWriter, r *http.Request) {
	email, err := emailOnly(r)
	if err == nil {
		err = h.svc.ForgotPassword(r.Context(), email)
	}
	if err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Token    string `json:"token" validate:"required"`
		Password string `json:"password" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	if err := h.svc.ResetPassword(r.Context(), in.Token, in.Password); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	var in struct {
		CurrentPassword string `json:"currentPassword" validate:"required"`
		NewPassword     string `json:"newPassword" validate:"required"`
	}
	if err := httpapi.Decode(r, &in); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	p, _ := PrincipalFrom(r.Context())
	if err := h.svc.ChangePassword(r.Context(), p, in.CurrentPassword, in.NewPassword); err != nil {
		httpapi.Error(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
