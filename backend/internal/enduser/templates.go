package enduser

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	htmltemplate "html/template"
	"net/mail"
	texttemplate "text/template"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// Chaves dos templates de e-mail.
const (
	TemplateEmailConfirmation = "email-confirmation"
	TemplateResetPassword     = "reset-password"
)

// TemplateData são as variáveis disponíveis nos templates (sintaxe de html/template):
// {{.User.Username}}, {{.User.Email}}, {{.Token}}, {{.URL}} e {{.ServerURL}}.
type TemplateData struct {
	User      TemplateUser
	Token     string
	URL       string
	ServerURL string
}

// TemplateUser é o usuário exposto aos templates.
type TemplateUser struct{ Username, Email string }

// TemplateView é um template de e-mail exposto pela API.
type TemplateView struct {
	Key         string    `json:"key"`
	SenderName  string    `json:"senderName"`
	SenderEmail string    `json:"senderEmail"`
	ReplyTo     string    `json:"replyTo"`
	Subject     string    `json:"subject"`
	BodyHTML    string    `json:"bodyHtml"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

const maxTemplateBody = 100 * 1024

func defaultTemplates(senderEmail string) []db.EmailTemplate {
	return []db.EmailTemplate{
		{
			Key: TemplateEmailConfirmation, SenderName: "Convênios e Fomentos", SenderEmail: senderEmail,
			Subject: "Confirme seu e-mail",
			BodyHtml: `<p>Olá, {{.User.Username}}!</p>
<p>Para confirmar seu e-mail, acesse o link abaixo:</p>
<p><a href="{{.URL}}">Confirmar e-mail</a></p>
<p>Se você não criou esta conta, ignore esta mensagem.</p>`,
		},
		{
			Key: TemplateResetPassword, SenderName: "Convênios e Fomentos", SenderEmail: senderEmail,
			Subject: "Redefinição de senha",
			BodyHtml: `<p>Olá, {{.User.Username}}!</p>
<p>Recebemos um pedido para redefinir sua senha. Acesse o link abaixo para escolher uma nova:</p>
<p><a href="{{.URL}}">Redefinir senha</a></p>
<p>O link expira em breve e só pode ser usado uma vez. Se não foi você, ignore esta mensagem.</p>`,
		},
	}
}

func templateView(t db.EmailTemplate) TemplateView {
	return TemplateView{Key: t.Key, SenderName: t.SenderName, SenderEmail: t.SenderEmail, ReplyTo: t.ReplyTo, Subject: t.Subject, BodyHTML: t.BodyHtml, UpdatedAt: t.UpdatedAt}
}

// renderTemplate processa assunto (texto) e corpo (HTML, com escape automático).
func renderTemplate(t db.EmailTemplate, data TemplateData) (subject, body string, err error) {
	st, err := texttemplate.New("subject").Option("missingkey=error").Parse(t.Subject)
	if err != nil {
		return "", "", fmt.Errorf("assunto: %w", err)
	}
	var sb bytes.Buffer
	if err := st.Execute(&sb, data); err != nil {
		return "", "", fmt.Errorf("assunto: %w", err)
	}
	bt, err := htmltemplate.New("body").Option("missingkey=error").Parse(t.BodyHtml)
	if err != nil {
		return "", "", fmt.Errorf("corpo: %w", err)
	}
	var bb bytes.Buffer
	if err := bt.Execute(&bb, data); err != nil {
		return "", "", fmt.Errorf("corpo: %w", err)
	}
	return sb.String(), bb.String(), nil
}

// ListTemplates lista os templates de e-mail.
func (s *Service) ListTemplates(ctx context.Context) ([]TemplateView, error) {
	rows, err := s.q.ListEmailTemplates(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]TemplateView, len(rows))
	for i, r := range rows {
		out[i] = templateView(r)
	}
	return out, nil
}

// GetTemplate devolve um template.
func (s *Service) GetTemplate(ctx context.Context, key string) (*TemplateView, error) {
	t, err := s.q.GetEmailTemplate(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrNotFound()
	}
	if err != nil {
		return nil, err
	}
	v := templateView(t)
	return &v, nil
}

// TemplateInput são os campos editáveis.
type TemplateInput struct {
	SenderName, SenderEmail, ReplyTo, Subject, BodyHTML string
}

// UpdateTemplate valida (e-mails, sintaxe e execução com dados de exemplo) e grava o template.
func (s *Service) UpdateTemplate(ctx context.Context, actor AdminActor, key string, in TemplateInput) (*TemplateView, error) {
	before, err := s.q.GetEmailTemplate(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, httpapi.ErrNotFound()
	}
	if err != nil {
		return nil, err
	}
	var issues []httpapi.FieldError
	if _, err := mail.ParseAddress(in.SenderEmail); err != nil {
		issues = append(issues, httpapi.FieldError{Field: "senderEmail", Message: "E-mail inválido."})
	}
	if in.ReplyTo != "" {
		if _, err := mail.ParseAddress(in.ReplyTo); err != nil {
			issues = append(issues, httpapi.FieldError{Field: "replyTo", Message: "E-mail inválido."})
		}
	}
	if len(in.BodyHTML) > maxTemplateBody {
		issues = append(issues, httpapi.FieldError{Field: "bodyHtml", Message: "O corpo excede 100 KB."})
	}
	if len(issues) == 0 {
		sample := TemplateData{User: TemplateUser{Username: "usuario", Email: "usuario@exemplo.com"}, Token: "token", URL: "https://exemplo.com/link", ServerURL: "https://exemplo.com"}
		candidate := db.EmailTemplate{Key: key, Subject: in.Subject, BodyHtml: in.BodyHTML}
		if subject, _, err := renderTemplate(candidate, sample); err != nil {
			issues = append(issues, httpapi.FieldError{Field: "bodyHtml", Message: "Template inválido: " + err.Error() + ". Variáveis: .User.Username, .User.Email, .Token, .URL, .ServerURL."})
		} else if subject == "" {
			issues = append(issues, httpapi.FieldError{Field: "subject", Message: "O assunto não pode ficar vazio."})
		}
	}
	if len(issues) > 0 {
		return nil, httpapi.ErrValidation(map[string]any{"fields": issues})
	}
	after, err := s.q.UpsertEmailTemplate(ctx, db.UpsertEmailTemplateParams{Key: key, SenderName: in.SenderName, SenderEmail: in.SenderEmail, ReplyTo: in.ReplyTo, Subject: in.Subject, BodyHtml: in.BodyHTML})
	if err != nil {
		return nil, err
	}
	if err := s.audit.Record(ctx, s.q, actor.entryKey("email-template.update", "email_template", key, templateView(before), templateView(after))); err != nil {
		return nil, err
	}
	v := templateView(after)
	return &v, nil
}
