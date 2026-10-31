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

// E-mails já resolvidos têm corpo com dados pessoais: saem depois da carência. Os pendentes nunca saem.
func TestCleanupApagaEmailsResolvidosAntigos(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	for i, c := range []struct{ status, age string }{
		{"sent", "30 days"}, {"failed", "30 days"}, {"sent", "2 days"}, {"pending", "30 days"},
	} {
		_, err := pool.Exec(ctx, `INSERT INTO outbound_emails (kind, to_email, subject, body, dedup_key, status, created_at)
			VALUES ('k', 'a@x.com', 's', 'corpo com dados pessoais', $1, $2, now() - $3::interval)`, string(rune('a'+i)), c.status, c.age)
		require.NoError(t, err)
	}
	res, err := maintenance.Cleanup(ctx, db.New(pool), 7)
	require.NoError(t, err)
	assert.EqualValues(t, 2, res.ResolvedEmails, "enviado e falho antigos")

	var left []string
	rows, err := pool.Query(ctx, "SELECT dedup_key FROM outbound_emails ORDER BY dedup_key")
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var k string
		require.NoError(t, rows.Scan(&k))
		left = append(left, k)
	}
	assert.Equal(t, []string{"c", "d"}, left, "ficam o enviado recente e o pendente")
}
