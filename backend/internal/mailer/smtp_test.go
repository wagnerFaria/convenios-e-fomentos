package mailer

import (
	"bufio"
	"context"
	"io"
	"mime"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildMensagem(t *testing.T) {
	s := SMTP{From: "no-reply@exemplo.com"}
	from, raw, err := s.Build(Message{To: "Ana <ana@exemplo.com>", FromName: "Equipe Ação", ReplyTo: "ajuda@exemplo.com", Subject: "Confirmação de e-mail", HTML: "<p>Olá, João!</p>"})
	require.NoError(t, err)
	assert.Equal(t, "no-reply@exemplo.com", from)
	m, err := mail.ReadMessage(strings.NewReader(string(raw)))
	require.NoError(t, err)
	subj, _ := new(mime.WordDecoder).DecodeHeader(m.Header.Get("Subject"))
	assert.Equal(t, "Confirmação de e-mail", subj)
	assert.Contains(t, m.Header.Get("Content-Type"), "text/html")
	assert.Equal(t, "<ajuda@exemplo.com>", m.Header.Get("Reply-To"))
	body, _ := io.ReadAll(quotedprintable.NewReader(m.Body))
	assert.Equal(t, "<p>Olá, João!</p>", string(body))
	fromHdr, _ := new(mime.WordDecoder).DecodeHeader(m.Header.Get("From"))
	assert.Contains(t, fromHdr, "Equipe Ação")

	_, raw, err = s.Build(Message{To: "a@b.com", From: "outro@exemplo.com", Subject: "x", Body: "texto"})
	require.NoError(t, err)
	assert.Contains(t, string(raw), "text/plain")
}

func TestBuildRecusaInjecaoDeCabecalho(t *testing.T) {
	s := SMTP{From: "no-reply@exemplo.com"}
	bad := []Message{
		{To: "a@b.com", Subject: "x\r\nBcc: vitima@exemplo.com"},
		{To: "a@b.com\r\nBcc: v@e.com", Subject: "x"},
		{To: "a@b.com", Subject: "x", ReplyTo: "r@e.com\nX: y"},
		{To: "a@b.com", Subject: "x", FromName: "Fulano\r\nBcc: v@e.com"},
		{To: "nao-e-email", Subject: "x"},
		{To: "a@b.com", Subject: "x", From: "invalido"},
	}
	for i, m := range bad {
		_, _, err := s.Build(m)
		assert.Error(t, err, i)
	}
}

// fakeSMTP aceita uma conexão, conversa o mínimo do protocolo e devolve a mensagem recebida.
func fakeSMTP(t *testing.T) (port int, got <-chan string) {
	t.Helper()
	ln, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	ch := make(chan string, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer func() { _ = conn.Close() }()
		r := bufio.NewReader(conn)
		say := func(s string) { _, _ = conn.Write([]byte(s + "\r\n")) }
		say("220 fake ESMTP")
		var data strings.Builder
		inData := false
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if inData {
				if line == "." {
					inData = false
					ch <- data.String()
					say("250 ok")
					continue
				}
				data.WriteString(line + "\n")
				continue
			}
			switch cmd := strings.ToUpper(line); {
			case strings.HasPrefix(cmd, "EHLO"), strings.HasPrefix(cmd, "HELO"):
				say("250 fake")
			case strings.HasPrefix(cmd, "MAIL FROM"), strings.HasPrefix(cmd, "RCPT TO"):
				say("250 ok")
				data.WriteString(line + "\n")
			case cmd == "DATA":
				say("354 go")
				inData = true
			case cmd == "QUIT":
				say("221 bye")
				return
			default:
				say("250 ok")
			}
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port, ch
}

func TestSMTPEnvio(t *testing.T) {
	port, got := fakeSMTP(t)
	s := SMTP{Host: "127.0.0.1", Port: port, From: "no-reply@exemplo.com"}
	require.NoError(t, s.Send(context.Background(), Message{To: "ana@exemplo.com", Subject: "Olá", HTML: "<b>oi</b>"}))
	raw := <-got
	assert.Contains(t, raw, "MAIL FROM:<no-reply@exemplo.com>")
	assert.Contains(t, raw, "RCPT TO:<ana@exemplo.com>")
	assert.Contains(t, raw, "text/html")
}

func TestSMTPFalhas(t *testing.T) {
	s := SMTP{Host: "127.0.0.1", Port: 1, From: "a@b.com", Timeout: 500_000_000}
	assert.Error(t, s.Send(context.Background(), Message{To: "x@y.com", Subject: "s", Body: "b"}), "sem servidor")
	assert.Error(t, s.Send(context.Background(), Message{To: "invalido", Subject: "s"}), "mensagem inválida nem conecta")
}
