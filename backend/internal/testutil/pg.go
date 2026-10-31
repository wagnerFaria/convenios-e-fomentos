// Pacote testutil oferece apoio aos testes de integração com PostgreSQL real.
package testutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

var (
	once    sync.Once
	baseURL string
	initErr error
)

// NewPool devolve um pool para um banco novo e migrado. O contêiner PostgreSQL é
// compartilhado pelo processo de teste (encerrado pelo Ryuk); cada teste recebe
// seu próprio banco, o que dá isolamento sem custo de subir outro contêiner.
func NewPool(t testing.TB) *pgxpool.Pool {
	t.Helper()
	if testing.Short() {
		t.Skip("teste de integração (ignorado com -short)")
	}
	ctx := context.Background()
	once.Do(func() {
		ctr, err := postgres.Run(ctx, "postgres:16-alpine",
			postgres.WithDatabase("postgres"), postgres.WithUsername("app"), postgres.WithPassword("app"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)))
		if err != nil {
			initErr = err
			return
		}
		baseURL, initErr = ctr.ConnectionString(ctx, "sslmode=disable")
	})
	if initErr != nil {
		t.Fatalf("subindo PostgreSQL: %v", initErr)
	}

	b := make([]byte, 6)
	_, _ = rand.Read(b)
	name := "t_" + hex.EncodeToString(b)
	root, err := pgxpool.New(ctx, baseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer root.Close()
	if _, err := root.Exec(ctx, "CREATE DATABASE "+name); err != nil {
		t.Fatal(err)
	}

	u, _ := url.Parse(baseURL)
	u.Path = "/" + name
	pool, err := database.Connect(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	m := database.NewMigrator(pool)
	defer func() { _ = m.Close() }()
	if err := m.Up(ctx); err != nil {
		t.Fatalf("migrando: %v", err)
	}
	return pool
}

// CaptureMailer guarda os e-mails enviados, para os testes lerem os tokens.
type CaptureMailer struct {
	mu       sync.Mutex
	Msgs     []mailer.Message
	failures int
}

// FailTimes faz os n próximos envios falharem (SMTP fora do ar), para os testes de nova tentativa.
func (c *CaptureMailer) FailTimes(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = n
}

// LastToken extrai o valor de "token=" do último e-mail enviado.
func (c *CaptureMailer) LastToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Msgs) == 0 {
		return ""
	}
	body := c.Msgs[len(c.Msgs)-1].Body
	i := strings.Index(body, "token=")
	if i < 0 {
		return ""
	}
	rest := body[i+len("token="):]
	if j := strings.IndexAny(rest, " \n\r"); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// Count devolve o número de e-mails enviados.
func (c *CaptureMailer) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.Msgs)
}

// Send implementa mailer.Mailer.
func (c *CaptureMailer) Send(_ context.Context, msg mailer.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failures > 0 {
		c.failures--
		return errors.New("SMTP fora do ar (simulado)")
	}
	c.Msgs = append(c.Msgs, msg)
	return nil
}

// LastParam extrai o valor de "nome=" (query string) do último e-mail enviado,
// procurando no texto e no HTML. O valor termina em aspas, espaço, '<' ou '&'.
func (c *CaptureMailer) LastParam(name string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Msgs) == 0 {
		return ""
	}
	m := c.Msgs[len(c.Msgs)-1]
	text := m.Body + "\n" + m.HTML
	i := strings.Index(text, name+"=")
	if i < 0 {
		return ""
	}
	rest := text[i+len(name)+1:]
	if j := strings.IndexAny(rest, "\" <&\n\r'"); j >= 0 {
		rest = rest[:j]
	}
	return rest
}

// Last devolve o último e-mail enviado.
func (c *CaptureMailer) Last() mailer.Message {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.Msgs) == 0 {
		return mailer.Message{}
	}
	return c.Msgs[len(c.Msgs)-1]
}
