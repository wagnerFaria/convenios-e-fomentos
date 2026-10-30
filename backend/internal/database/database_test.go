package database_test

import (
	"context"
	"testing"

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

	require.NoError(t, m.Down(ctx))
	require.NoError(t, m.Up(ctx))
}
