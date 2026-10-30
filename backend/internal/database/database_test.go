package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

func TestMigracoesEConsultas(t *testing.T) {
	if testing.Short() {
		t.Skip("teste de integração")
	}
	ctx := context.Background()
	ctr, err := postgres.Run(ctx, "postgres:16-alpine",
		postgres.WithDatabase("app"), postgres.WithUsername("app"), postgres.WithPassword("app"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2)))
	require.NoError(t, err)
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(ctr) })

	url, err := ctr.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	pool, err := database.Connect(ctx, url)
	require.NoError(t, err)
	t.Cleanup(pool.Close)

	m := database.NewMigrator(pool)
	t.Cleanup(func() { _ = m.Close() })
	require.NoError(t, m.Up(ctx))
	require.NoError(t, m.Up(ctx), "Up deve ser idempotente")

	q := db.New(pool)
	d1, err := q.UpsertDepartment(ctx, db.UpsertDepartmentParams{Name: "Financeiro"})
	require.NoError(t, err)
	d2, err := q.UpsertDepartment(ctx, db.UpsertDepartmentParams{Name: "Financeiro"})
	require.NoError(t, err)
	require.Equal(t, d1.ID, d2.ID)

	// A 00024 altera permissões direto no banco. O cache do motor não tem prazo de validade, então a migração avisa as instâncias em
	// execução pelo canal authz_invalidate (lista vazia = recarregar todos os perfis), tanto ao subir quanto ao descer.
	conn, err := pool.Acquire(ctx)
	require.NoError(t, err)
	t.Cleanup(conn.Release)
	_, err = conn.Exec(ctx, "LISTEN authz_invalidate")
	require.NoError(t, err)
	expectInvalidation := func(when string) {
		t.Helper()
		wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		n, err := conn.Conn().WaitForNotification(wctx)
		require.NoError(t, err, "a migração deve avisar o cache ao %s", when)
		require.Equal(t, "[]", n.Payload)
	}
	require.NoError(t, m.Down(ctx))
	expectInvalidation("descer")
	require.NoError(t, m.Up(ctx))
	expectInvalidation("subir")
}
