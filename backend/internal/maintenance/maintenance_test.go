package maintenance_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/maintenance"
	"github.com/wagner/convenios-e-fomentos/backend/internal/testutil"
)

func TestCleanupRemoveSoOQueVenceuHaMaisDeNDias(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	q := db.New(pool)
	exec := func(sql string, args ...any) {
		_, err := pool.Exec(ctx, sql, args...)
		require.NoError(t, err)
	}
	role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "r", Name: "R"})
	require.NoError(t, err)
	_ = role
	admin, err := q.CreateAdmin(ctx, db.CreateAdminParams{FirstName: "a", LastName: "b", Email: "a@x.com", IsActive: true})
	require.NoError(t, err)
	eRole, err := q.UpsertEndUserRole(ctx, db.UpsertEndUserRoleParams{Type: "authenticated", Name: "A"})
	require.NoError(t, err)
	user, err := q.CreateEndUser(ctx, db.CreateEndUserParams{Username: "u", Email: "u@x.com", RoleID: eRole.ID, Profile: []byte("{}")})
	require.NoError(t, err)

	// Sessões: 1 ativa, 1 expirada há 2 dias (dentro da carência), 1 expirada há 30 dias, 1 revogada e expirada há 30 dias.
	for i, c := range []struct{ expires, revoked string }{
		{"now() + interval '1 day'", "null"}, {"now() - interval '2 days'", "null"}, {"now() - interval '30 days'", "null"}, {"now() - interval '30 days'", "now() - interval '31 days'"},
	} {
		exec("INSERT INTO end_user_sessions (end_user_id, family_id, refresh_token_hash, expires_at, revoked_at) VALUES ($1, gen_random_uuid(), $2, "+c.expires+", "+c.revoked+")", user.ID, "u"+string(rune('0'+i)))
	}
	// Tokens de uso único: pendente válido, expirado recente, expirado antigo, usado antigo.
	for i, c := range []struct{ expires, used string }{
		{"now() + interval '1 hour'", "null"}, {"now() - interval '1 day'", "null"}, {"now() - interval '20 days'", "null"}, {"now() + interval '1 hour'", "now() - interval '20 days'"},
	} {
		exec("INSERT INTO one_time_tokens (subject_type, subject_id, purpose, token_hash, expires_at, used_at) VALUES ('admin', $1, 'reset', $2, "+c.expires+", "+c.used+")", admin.ID, "t"+string(rune('0'+i)))
	}

	res, err := maintenance.Cleanup(ctx, q, 7)
	require.NoError(t, err)
	assert.Equal(t, maintenance.Result{EndUserSessions: 2, OneTimeTokens: 2}, res)

	count := func(table string) int {
		var n int
		require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM "+table).Scan(&n))
		return n
	}
	assert.Equal(t, 2, count("end_user_sessions"))
	assert.Equal(t, 2, count("one_time_tokens"))

	// Idempotente.
	res, err = maintenance.Cleanup(ctx, q, 7)
	require.NoError(t, err)
	assert.Equal(t, maintenance.Result{}, res)
	// Com carência zero, a sessão expirada há 2 dias e o token expirado há 1 dia também saem.
	res, err = maintenance.Cleanup(ctx, q, 0)
	require.NoError(t, err)
	assert.Equal(t, maintenance.Result{EndUserSessions: 1, OneTimeTokens: 1}, res)

	_, err = maintenance.Cleanup(ctx, q, -1)
	assert.Error(t, err)
}
