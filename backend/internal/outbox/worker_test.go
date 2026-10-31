package outbox_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
	"github.com/wagner/convenios-e-fomentos/backend/internal/outbox"
	"github.com/wagner/convenios-e-fomentos/backend/internal/testutil"
)

var quiet = slog.New(slog.NewJSONHandler(io.Discard, nil))

func testConfig() outbox.Config {
	c := outbox.DefaultConfig()
	c.Interval, c.Lease, c.SendTimeout = 10*time.Millisecond, time.Minute, 2*time.Second
	c.Backoff = []time.Duration{time.Hour}
	return c
}

func enqueue(t *testing.T, pool *pgxpool.Pool, to, key string) {
	t.Helper()
	_, err := outbox.Enqueue(context.Background(), db.New(pool), outbox.Message{Kind: "teste", To: to, Subject: "Assunto", Body: "Corpo", DedupKey: key})
	require.NoError(t, err)
}

type row struct {
	Status, LastError string
	Attempts          int
}

func rowOf(t *testing.T, pool *pgxpool.Pool, to string) row {
	t.Helper()
	var r row
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT status, last_error, attempts FROM outbound_emails WHERE to_email = $1", to).Scan(&r.Status, &r.LastError, &r.Attempts))
	return r
}

// failing falha nas n primeiras chamadas e depois funciona.
type failing struct {
	mu    sync.Mutex
	fails int
	sent  []string
}

func (f *failing) Send(_ context.Context, m mailer.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fails > 0 {
		f.fails--
		return errors.New("SMTP fora do ar")
	}
	f.sent = append(f.sent, m.To)
	return nil
}

func TestEnqueueNaoDuplicaEIgnoraDestinatarioVazio(t *testing.T) {
	pool := testutil.NewPool(t)
	q := db.New(pool)
	m := outbox.Message{Kind: "k", To: "a@x.com", Subject: "s", Body: "b", DedupKey: "mesma"}
	ok, err := outbox.Enqueue(context.Background(), q, m)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = outbox.Enqueue(context.Background(), q, m)
	require.NoError(t, err)
	assert.False(t, ok, "a mesma chave não entra duas vezes")

	ok, err = outbox.Enqueue(context.Background(), q, outbox.Message{Kind: "k", To: "", Subject: "s", Body: "b"})
	require.NoError(t, err)
	assert.False(t, ok, "sem destinatário não grava")

	// Sem chave, cada chamada grava (mensagens repetidas de propósito, como convites).
	for range 2 {
		ok, err = outbox.Enqueue(context.Background(), q, outbox.Message{Kind: "k", To: "b@x.com", Subject: "s", Body: "b"})
		require.NoError(t, err)
		assert.True(t, ok)
	}
	var n int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT count(*) FROM outbound_emails").Scan(&n))
	assert.Equal(t, 3, n)
}

func TestEnqueueDentroDeTransacaoDesfeitaNaoDeixaRastro(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	tx, err := pool.Begin(ctx)
	require.NoError(t, err)
	_, err = outbox.Enqueue(ctx, db.New(pool).WithTx(tx), outbox.Message{Kind: "k", To: "a@x.com", Subject: "s", Body: "b"})
	require.NoError(t, err)
	require.NoError(t, tx.Rollback(ctx))
	var n int
	require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM outbound_emails").Scan(&n))
	assert.Equal(t, 0, n, "se a mudança de estado desfaz, o e-mail também não existe")
}

func TestDrainEnviaEMarcaComoEnviado(t *testing.T) {
	pool := testutil.NewPool(t)
	mail := &testutil.CaptureMailer{}
	w := outbox.NewWorker(pool, mail, quiet, testConfig())
	enqueue(t, pool, "a@x.com", "1")
	enqueue(t, pool, "b@x.com", "2")

	n, err := w.Drain(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, 2, mail.Count())
	assert.Equal(t, row{Status: "sent", Attempts: 1}, rowOf(t, pool, "a@x.com"))

	n, err = w.Drain(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, n, "o que já foi enviado não é reenviado")
}

func TestFalhaAgendaNovaTentativaComRecuo(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	mail := &failing{fails: 1}
	w := outbox.NewWorker(pool, mail, quiet, testConfig())
	enqueue(t, pool, "a@x.com", "1")

	n, err := w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, row{Status: "pending", Attempts: 1, LastError: "SMTP fora do ar"}, rowOf(t, pool, "a@x.com"))

	// Dentro do recuo a mensagem não é tentada de novo.
	n, err = w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Empty(t, mail.sent)

	// Passado o recuo, sai.
	_, err = pool.Exec(ctx, "UPDATE outbound_emails SET next_attempt_at = now() - interval '1 second'")
	require.NoError(t, err)
	_, err = w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, []string{"a@x.com"}, mail.sent)
	assert.Equal(t, row{Status: "sent", Attempts: 2}, rowOf(t, pool, "a@x.com"), "o erro anterior é limpo ao enviar")
}

func TestDesisteDepoisDoLimiteDeTentativas(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	cfg := testConfig()
	cfg.MaxAttempts = 3
	w := outbox.NewWorker(pool, &failing{fails: 100}, quiet, cfg)
	enqueue(t, pool, "a@x.com", "1")

	for range 3 {
		_, err := w.Drain(ctx)
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "UPDATE outbound_emails SET next_attempt_at = now() - interval '1 second' WHERE status = 'pending'")
		require.NoError(t, err)
	}
	r := rowOf(t, pool, "a@x.com")
	assert.Equal(t, "failed", r.Status)
	assert.Equal(t, 3, r.Attempts)
	assert.Equal(t, "SMTP fora do ar", r.LastError)

	n, err := w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "o que desistiu não volta à fila")
}

// Se a instância cai depois de reservar a mensagem, a concessão vence e outra instância a envia.
func TestConcessaoVencidaVoltaParaAFila(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	mail := &testutil.CaptureMailer{}
	w := outbox.NewWorker(pool, mail, quiet, testConfig())
	enqueue(t, pool, "a@x.com", "1")

	// "Instância que caiu": reservou e nunca marcou.
	claimed, err := db.New(pool).ClaimDueEmails(ctx, db.ClaimDueEmailsParams{LeaseSeconds: 3600, Batch: 10})
	require.NoError(t, err)
	require.Len(t, claimed, 1)

	n, err := w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "enquanto a concessão vale, ninguém mais pega")

	_, err = pool.Exec(ctx, "UPDATE outbound_emails SET locked_until = now() - interval '1 second'")
	require.NoError(t, err)
	n, err = w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.Equal(t, 1, mail.Count())
	assert.Equal(t, row{Status: "sent", Attempts: 2}, rowOf(t, pool, "a@x.com"), "a reserva perdida conta como tentativa")
}

func TestQuedasSeguidasNaoInsistemParaSempre(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	cfg := testConfig()
	cfg.MaxAttempts = 2
	mail := &testutil.CaptureMailer{}
	w := outbox.NewWorker(pool, mail, quiet, cfg)
	enqueue(t, pool, "a@x.com", "1")

	for range 3 { // três reservas perdidas
		_, err := db.New(pool).ClaimDueEmails(ctx, db.ClaimDueEmailsParams{LeaseSeconds: 3600, Batch: 10})
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "UPDATE outbound_emails SET locked_until = now() - interval '1 second'")
		require.NoError(t, err)
	}
	_, err := w.Drain(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, mail.Count(), "não envia")
	assert.Equal(t, "failed", rowOf(t, pool, "a@x.com").Status)
}

// Duas instâncias esvaziando a mesma fila: cada mensagem sai exatamente uma vez.
func TestDuasInstanciasNaoEnviamAMesmaMensagem(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	const total = 60
	for i := range total {
		enqueue(t, pool, fmt.Sprintf("u%02d@x.com", i), fmt.Sprint(i))
	}
	mail := &testutil.CaptureMailer{}
	cfg := testConfig()
	cfg.Batch = 7
	w1, w2 := outbox.NewWorker(pool, mail, quiet, cfg), outbox.NewWorker(pool, mail, quiet, cfg)

	var wg sync.WaitGroup
	for _, w := range []*outbox.Worker{w1, w2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := w.Drain(ctx)
			assert.NoError(t, err)
		}()
	}
	wg.Wait()

	seen := map[string]int{}
	for _, m := range mail.Msgs {
		seen[m.To]++
	}
	assert.Len(t, seen, total)
	for to, n := range seen {
		assert.Equal(t, 1, n, "%s enviado %d vezes", to, n)
	}
}

// slow só termina quando liberado, e ignora o cancelamento do contexto (como um envio já entregue ao servidor).
type slow struct {
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (s *slow) Send(_ context.Context, _ mailer.Message) error {
	s.calls.Add(1)
	s.started <- struct{}{}
	<-s.release
	return nil
}

// O envio acontece fora de quem gravou: gravar não espera o SMTP, e um SMTP lento não impede nada além do próprio envio.
func TestRunEnviaEmSegundoPlano(t *testing.T) {
	pool := testutil.NewPool(t)
	s := &slow{started: make(chan struct{}, 1), release: make(chan struct{})}
	w := outbox.NewWorker(pool, s, quiet, testConfig())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()

	begin := time.Now()
	enqueue(t, pool, "a@x.com", "1") // volta na hora, mesmo com o SMTP travado
	assert.Less(t, time.Since(begin), time.Second)

	select {
	case <-s.started:
	case <-time.After(5 * time.Second):
		t.Fatal("o worker não começou a enviar")
	}
	assert.Equal(t, "pending", rowOf(t, pool, "a@x.com").Status, "ainda não terminou")
	close(s.release)

	require.Eventually(t, func() bool { return rowOf(t, pool, "a@x.com").Status == "sent" }, 5*time.Second, 20*time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("o worker não parou com o contexto")
	}
}

// No desligamento, o envio em andamento termina e o resultado é gravado: a mensagem não volta à fila para sair duas vezes.
func TestDesligamentoGravaOEnvioEmAndamento(t *testing.T) {
	pool := testutil.NewPool(t)
	s := &slow{started: make(chan struct{}, 1), release: make(chan struct{})}
	w := outbox.NewWorker(pool, s, quiet, testConfig())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { w.Run(ctx); close(done) }()
	enqueue(t, pool, "a@x.com", "1")
	<-s.started

	cancel() // desligando com um envio no meio
	select {
	case <-done:
		t.Fatal("o worker não deveria terminar antes do envio em andamento")
	case <-time.After(200 * time.Millisecond):
	}
	close(s.release)
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("o worker não terminou depois do envio")
	}
	assert.Equal(t, "sent", rowOf(t, pool, "a@x.com").Status)
	assert.EqualValues(t, 1, s.calls.Load(), "enviado uma única vez")
}
