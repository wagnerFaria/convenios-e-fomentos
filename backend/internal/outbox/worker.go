package outbox

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

// Config ajusta o worker. Os valores padrão (DefaultConfig) servem à produção; os testes encurtam os tempos.
type Config struct {
	// Interval é a pausa entre uma varredura e outra quando a fila está vazia.
	Interval time.Duration
	// Workers é quantos e-mails saem ao mesmo tempo, para um SMTP lento não segurar os demais.
	Workers int
	// Batch é quantas mensagens cada varredura reserva.
	Batch int
	// Lease é por quanto tempo uma mensagem reservada fica fora da fila. Se o processo cair, ela volta depois desse prazo.
	Lease time.Duration
	// SendTimeout limita cada envio.
	SendTimeout time.Duration
	// MaxAttempts é o número de tentativas antes de desistir (a mensagem fica "failed", com o último erro).
	MaxAttempts int32
	// Backoff é a espera depois da 1ª, 2ª, ... falha; a última repete.
	Backoff []time.Duration
}

// DefaultConfig são os valores de produção: tentativas em 1 min, 5 min, 30 min, 2 h e 12 h (e a de 12 h repete), até 8 no total.
func DefaultConfig() Config {
	return Config{
		Interval: 5 * time.Second, Workers: 4, Batch: 20, Lease: 2 * time.Minute, SendTimeout: 45 * time.Second, MaxAttempts: 8,
		Backoff: []time.Duration{time.Minute, 5 * time.Minute, 30 * time.Minute, 2 * time.Hour, 12 * time.Hour},
	}
}

// Worker envia a caixa de saída. Várias instâncias da API podem rodar o mesmo Worker ao mesmo tempo: cada mensagem é reservada
// com SKIP LOCKED e uma concessão, então ninguém pega a mesma.
//
// A entrega é "pelo menos uma vez": se a instância cair entre enviar e marcar como enviada, a concessão vence e a mensagem sai de
// novo. É raro, e é preferível a perder o e-mail.
type Worker struct {
	q    *db.Queries
	mail mailer.Mailer
	log  *slog.Logger
	cfg  Config
	now  func() time.Time
}

// NewWorker cria o worker.
func NewWorker(pool *pgxpool.Pool, mail mailer.Mailer, log *slog.Logger, cfg Config) *Worker {
	return &Worker{q: db.New(pool), mail: mail, log: log, cfg: cfg, now: time.Now}
}

// Run varre a fila até o contexto terminar. Com a fila cheia, repete logo; vazia, espera Interval. Ao receber o fim do contexto,
// termina os envios em andamento (registrando o resultado deles) e não reserva mais nada.
func (w *Worker) Run(ctx context.Context) {
	t := time.NewTimer(0)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
		claimed, err := w.process(ctx)
		if err != nil && ctx.Err() == nil {
			w.log.ErrorContext(ctx, "varrendo a caixa de saída de e-mails", "error", err)
		}
		wait := w.cfg.Interval
		if claimed >= w.cfg.Batch {
			wait = 0 // ainda deve haver mensagens esperando
		}
		t.Reset(wait)
	}
}

// Drain envia agora tudo o que está vencido, sem esperar o intervalo, e devolve quantas mensagens reservou. Serve aos testes e a
// quem precisa esvaziar a fila (por exemplo, antes de desligar). Não espera as que ainda estão aguardando uma nova tentativa.
func (w *Worker) Drain(ctx context.Context) (int, error) {
	total := 0
	for {
		n, err := w.process(ctx)
		total += n
		if err != nil || n < w.cfg.Batch {
			return total, err
		}
	}
}

// process reserva um lote e envia. Devolve quantas mensagens reservou.
func (w *Worker) process(ctx context.Context) (int, error) {
	rows, err := w.q.ClaimDueEmails(ctx, db.ClaimDueEmailsParams{LeaseSeconds: clampInt32(w.cfg.Lease.Seconds()), Batch: clampInt32(float64(w.cfg.Batch))})
	if err != nil || len(rows) == 0 {
		return 0, err
	}
	sem := make(chan struct{}, max(w.cfg.Workers, 1))
	var wg sync.WaitGroup
	for _, r := range rows {
		sem <- struct{}{}
		wg.Add(1)
		go func(r db.ClaimDueEmailsRow) {
			defer func() { <-sem; wg.Done() }()
			w.deliver(ctx, r)
		}(r)
	}
	wg.Wait()
	return len(rows), nil
}

func (w *Worker) deliver(ctx context.Context, r db.ClaimDueEmailsRow) {
	// O resultado é gravado mesmo que o contexto da varredura já tenha terminado (desligamento): senão a mensagem enviada
	// voltaria à fila e sairia de novo.
	rec, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
	defer cancel()

	if r.Attempts > w.cfg.MaxAttempts {
		// Foi reservada e perdida (a instância caiu) mais vezes que o permitido: desiste em vez de insistir para sempre.
		w.fail(rec, r, "excedeu o número de tentativas (a instância caiu durante o envio)")
		return
	}
	sendCtx, stop := context.WithTimeout(ctx, w.cfg.SendTimeout)
	err := w.mail.Send(sendCtx, mailer.Message{To: r.ToEmail, Subject: r.Subject, Body: r.Body})
	stop()
	if err == nil {
		if err := w.q.MarkEmailSent(rec, r.ID); err != nil {
			w.log.ErrorContext(rec, "marcando e-mail como enviado", "id", r.ID, "kind", r.Kind, "error", err)
		}
		return
	}
	if ctx.Err() != nil {
		// Desligamento no meio do envio: não conta como falha do e-mail; a concessão vence e outra instância tenta.
		return
	}
	msg := truncate(err.Error(), 500)
	if r.Attempts >= w.cfg.MaxAttempts {
		w.fail(rec, r, msg)
		return
	}
	next := w.now().Add(w.backoff(int(r.Attempts)))
	w.log.WarnContext(rec, "falha ao enviar e-mail; nova tentativa agendada", "id", r.ID, "kind", r.Kind, "tentativa", r.Attempts, "proxima", next, "error", msg)
	if err := w.q.RescheduleEmail(rec, db.RescheduleEmailParams{ID: r.ID, NextAttemptAt: next, LastError: msg}); err != nil {
		w.log.ErrorContext(rec, "agendando nova tentativa de e-mail", "id", r.ID, "error", err)
	}
}

func (w *Worker) fail(ctx context.Context, r db.ClaimDueEmailsRow, reason string) {
	w.log.ErrorContext(ctx, "e-mail desistido", "id", r.ID, "kind", r.Kind, "tentativas", r.Attempts, "error", reason)
	if err := w.q.FailEmail(ctx, db.FailEmailParams{ID: r.ID, LastError: reason}); err != nil {
		w.log.ErrorContext(ctx, "marcando e-mail como falho", "id", r.ID, "error", err)
	}
}

// backoff é a espera depois da falha de número attempt (1 = primeira).
func (w *Worker) backoff(attempt int) time.Duration {
	if len(w.cfg.Backoff) == 0 {
		return time.Minute
	}
	return w.cfg.Backoff[min(attempt-1, len(w.cfg.Backoff)-1)]
}

// clampInt32 converte para int32 sem estourar (valores absurdos de configuração viram o limite, não um número negativo).
func clampInt32(v float64) int32 {
	return int32(min(max(v, 1), 1<<30)) //nolint:gosec // limitado acima
}

func truncate(s string, n int) string {
	if r := []rune(s); len(r) > n {
		return string(r[:n])
	}
	return s
}
