package authz

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Broadcaster propaga invalidações de cache entre instâncias. Uma lista vazia
// de IDs significa "todas as roles".
type Broadcaster interface {
	Publish(ctx context.Context, roleIDs []uuid.UUID) error
	Subscribe(ctx context.Context, fn func(roleIDs []uuid.UUID)) error
}

// LocalBroadcaster é a implementação para instância única: não propaga nada.
type LocalBroadcaster struct{}

// Publish não faz nada.
func (LocalBroadcaster) Publish(context.Context, []uuid.UUID) error { return nil }

// Subscribe bloqueia até o contexto terminar.
func (LocalBroadcaster) Subscribe(ctx context.Context, _ func([]uuid.UUID)) error {
	<-ctx.Done()
	return nil
}

const (
	notifyChannel = "authz_invalidate"
	// O payload do NOTIFY é limitado a 8000 bytes; acima disso invalida tudo.
	maxNotifyIDs = 100
)

// PGBroadcaster usa LISTEN/NOTIFY do PostgreSQL para sincronizar várias instâncias.
type PGBroadcaster struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

// Publish envia NOTIFY com os IDs (ou "todos").
func (b PGBroadcaster) Publish(ctx context.Context, roleIDs []uuid.UUID) error {
	if len(roleIDs) > maxNotifyIDs {
		roleIDs = nil
	}
	payload, err := json.Marshal(roleIDs)
	if err != nil {
		return err
	}
	_, err = b.Pool.Exec(ctx, "SELECT pg_notify($1, $2)", notifyChannel, string(payload))
	return err
}

// Subscribe escuta o canal até o contexto terminar, reconectando com espera
// crescente. Após cada reconexão invalida tudo, pois notificações podem ter se perdido.
func (b PGBroadcaster) Subscribe(ctx context.Context, fn func([]uuid.UUID)) error {
	backoff := time.Second
	for ctx.Err() == nil {
		err := b.listenOnce(ctx, fn, &backoff)
		if ctx.Err() != nil {
			return nil
		}
		b.Log.WarnContext(ctx, "escuta de invalidação de permissões interrompida; reconectando", "error", err, "em", backoff)
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, 30*time.Second)
	}
	return nil
}

func (b PGBroadcaster) listenOnce(ctx context.Context, fn func([]uuid.UUID), backoff *time.Duration) error {
	conn, err := b.Pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()
	if _, err := conn.Exec(ctx, "LISTEN "+notifyChannel); err != nil {
		return err
	}
	*backoff = time.Second
	fn(nil) // pode ter havido alterações enquanto não escutávamos
	for {
		n, err := conn.Conn().WaitForNotification(ctx)
		if err != nil {
			return err
		}
		var ids []uuid.UUID
		if err := json.Unmarshal([]byte(n.Payload), &ids); err != nil {
			b.Log.ErrorContext(ctx, "payload de invalidação inválido; invalidando tudo", "error", err)
			ids = nil
		}
		fn(ids)
	}
}
