package mailer

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// SMTP envia e-mails por um servidor SMTP. Na porta 465 usa TLS implícito; nas
// demais, STARTTLS quando o servidor o oferece. Só autentica se User estiver preenchido.
type SMTP struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string // remetente padrão (SMTP_FROM)
	Timeout  time.Duration
}

var errHeader = errors.New("cabeçalho de e-mail inválido")

// cleanHeader recusa quebras de linha, que permitiriam injetar cabeçalhos.
func cleanHeader(s string) (string, error) {
	if strings.ContainsAny(s, "\r\n") {
		return "", errHeader
	}
	return strings.TrimSpace(s), nil
}

func address(name, email string) (string, error) {
	email, err := cleanHeader(email)
	if err != nil {
		return "", err
	}
	name, err = cleanHeader(name)
	if err != nil {
		return "", err
	}
	a := mail.Address{Name: name, Address: email}
	if _, err := mail.ParseAddress(a.Address); err != nil {
		return "", fmt.Errorf("%w: endereço %q", errHeader, email)
	}
	return a.String(), nil
}

// Build monta a mensagem MIME e devolve o remetente do envelope.
func (s SMTP) Build(msg Message) (envelopeFrom string, raw []byte, err error) {
	from := msg.From
	if from == "" {
		from = s.From
	}
	fromHdr, err := address(msg.FromName, from)
	if err != nil {
		return "", nil, err
	}
	toHdr, err := address("", msg.To)
	if err != nil {
		return "", nil, err
	}
	subject, err := cleanHeader(msg.Subject)
	if err != nil {
		return "", nil, err
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\n", fromHdr, toHdr)
	if msg.ReplyTo != "" {
		rt, err := address("", msg.ReplyTo)
		if err != nil {
			return "", nil, err
		}
		fmt.Fprintf(&b, "Reply-To: %s\r\n", rt)
	}
	fmt.Fprintf(&b, "Subject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\n", mime.QEncoding.Encode("utf-8", subject), time.Now().Format(time.RFC1123Z))
	body, ctype := msg.Body, "text/plain"
	if msg.HTML != "" {
		body, ctype = msg.HTML, "text/html"
	}
	fmt.Fprintf(&b, "Content-Type: %s; charset=utf-8\r\nContent-Transfer-Encoding: quoted-printable\r\n\r\n", ctype)
	w := quotedprintable.NewWriter(&b)
	_, _ = w.Write([]byte(body))
	_ = w.Close()
	parsed, _ := mail.ParseAddress(from)
	return parsed.Address, b.Bytes(), nil
}

// Send entrega a mensagem.
func (s SMTP) Send(ctx context.Context, msg Message) error {
	envFrom, raw, err := s.Build(msg)
	if err != nil {
		return err
	}
	to, _ := mail.ParseAddress(msg.To)

	timeout := s.Timeout
	if timeout == 0 {
		timeout = 15 * time.Second
	}
	addr := net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
	d := net.Dialer{Timeout: timeout}
	var conn net.Conn
	if s.Port == 465 {
		td := tls.Dialer{NetDialer: &d, Config: &tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}}
		conn, err = td.DialContext(ctx, "tcp", addr)
	} else {
		conn, err = d.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return fmt.Errorf("conectando ao SMTP: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(2 * timeout))
	c, err := smtp.NewClient(conn, s.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("iniciando SMTP: %w", err)
	}
	defer func() { _ = c.Close() }()
	if s.Port != 465 {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(&tls.Config{ServerName: s.Host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("STARTTLS: %w", err)
			}
		}
	}
	if s.User != "" {
		if err := c.Auth(smtp.PlainAuth("", s.User, s.Password, s.Host)); err != nil {
			return fmt.Errorf("autenticando no SMTP: %w", err)
		}
	}
	if err := c.Mail(envFrom); err != nil {
		return err
	}
	if err := c.Rcpt(to.Address); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(raw); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
