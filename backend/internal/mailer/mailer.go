// Pacote mailer define o envio de e-mails e as implementações disponíveis.
package mailer

import (
	"context"
	"log/slog"
)

// Message é um e-mail. Se HTML estiver preenchido, ele é enviado como corpo
// (e Body vira a alternativa em texto simples quando existir).
type Message struct {
	To       string
	From     string // vazio = remetente padrão do transporte
	FromName string
	ReplyTo  string
	Subject  string
	Body     string // texto simples
	HTML     string
}

// Mailer envia e-mails.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
}

// LogMailer apenas registra o e-mail no log. Uso restrito a desenvolvimento:
// o corpo pode conter tokens, por isso nunca deve ser usado em produção.
type LogMailer struct{ Log *slog.Logger }

// Send registra a mensagem no log.
func (m LogMailer) Send(ctx context.Context, msg Message) error {
	m.Log.InfoContext(ctx, "e-mail (modo desenvolvimento)", "to", msg.To, "subject", msg.Subject, "body", msg.Body, "html", msg.HTML)
	return nil
}

// NopMailer descarta o e-mail, registrando apenas o destinatário (sem o corpo).
// Usado em produção enquanto nenhum SMTP estiver configurado.
type NopMailer struct{ Log *slog.Logger }

// Send registra um aviso sem o conteúdo da mensagem.
func (m NopMailer) Send(ctx context.Context, msg Message) error {
	m.Log.WarnContext(ctx, "e-mail descartado: nenhum transporte configurado", "to", msg.To, "subject", msg.Subject)
	return nil
}
