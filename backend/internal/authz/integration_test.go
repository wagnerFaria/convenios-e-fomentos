package authz

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/testutil"
)

var discardLog = slog.New(slog.NewJSONHandler(io.Discard, nil))

func addPerm(t *testing.T, q *db.Queries, roleID uuid.UUID, action string, subject *string, fields, conds string) {
	t.Helper()
	var f []byte
	if fields != "" {
		f = []byte(fields)
	}
	_, err := q.CreateAdminPermission(context.Background(), db.CreateAdminPermissionParams{
		RoleID: roleID, Action: action, Subject: subject, Fields: f, Conditions: []byte(conds), Properties: []byte(`{"locales":["pt-BR"]}`)})
	require.NoError(t, err)
}

func TestDBStoreESincronizacao(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	q := db.New(pool)
	role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "editor", Name: "Editor"})
	require.NoError(t, err)
	doc := "document"
	addPerm(t, q, role.ID, read, &doc, `["titulo","conteudo"]`, `["admin::is-creator","admin::same-department"]`)
	addPerm(t, q, role.ID, update, nil, "", `[]`)
	addPerm(t, q, role.ID, "admin::content.removida.read", nil, "", `[]`)

	store := DBStore{Q: q}
	perms, err := store.RolePermissions(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, perms, 3)
	byAction := map[string]Permission{}
	for _, p := range perms {
		byAction[p.Action] = p
	}
	assert.Equal(t, "document", byAction[read].Subject)
	assert.Equal(t, []string{"titulo", "conteudo"}, byAction[read].Fields)
	assert.Equal(t, []string{"admin::is-creator", "admin::same-department"}, byAction[read].Conditions)
	assert.Nil(t, byAction[update].Fields, "fields nulo = todos")
	assert.Empty(t, byAction[update].Subject)
	assert.Equal(t, []any{"pt-BR"}, byAction[read].Properties["locales"])

	// A mesma ação/recurso não pode se repetir na role (inclusive com subject nulo).
	_, err = q.CreateAdminPermission(ctx, db.CreateAdminPermissionParams{RoleID: role.ID, Action: update, Conditions: []byte(`[]`), Properties: []byte(`{}`)})
	assert.Error(t, err)
	_, err = q.CreateAdminPermission(ctx, db.CreateAdminPermissionParams{RoleID: role.ID, Action: "admin::x.y", Fields: []byte(`{}`), Conditions: []byte(`[]`), Properties: []byte(`{}`)})
	assert.Error(t, err, "fields precisa ser array")

	// Sincronização: a ação removida do código perde as permissões.
	reg := NewRegistry()
	for _, v := range []string{"read", "update"} {
		reg.MustRegister(Action{ID: "admin::content.document." + v, Domain: DomainAdmin, Category: CategoryCollectionTypes, Name: v})
	}
	eng := NewEngine(reg, NewConditionRegistry(), store, LocalBroadcaster{}, discardLog)
	s := Subject{Type: SubjectAdmin, ID: uuid.New(), Roles: []Role{{ID: role.ID, Code: "editor"}}}
	_, err = eng.Effective(ctx, s) // aquece o cache
	require.NoError(t, err)
	require.NoError(t, SyncActions(ctx, q, reg, eng, discardLog))
	perms, _ = store.RolePermissions(ctx, role.ID)
	assert.Len(t, perms, 2)
	d, err := eng.Can(ctx, s, update)
	require.NoError(t, err)
	assert.True(t, d.Allowed)

	// Registro vazio aborta, sem apagar nada.
	assert.Error(t, SyncActions(ctx, q, NewRegistry(), eng, discardLog))
	perms, _ = store.RolePermissions(ctx, role.ID)
	assert.Len(t, perms, 2)
}

func TestMembrosDeRoles(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx := context.Background()
	q := db.New(pool)
	r1, _ := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "r1", Name: "R1"})
	r2, _ := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "r2", Name: "R2"})
	r3, _ := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "r3", Name: "R3"})
	mk := func(email string, roles ...db.AdminRole) uuid.UUID {
		a, err := q.CreateAdmin(ctx, db.CreateAdminParams{FirstName: "a", LastName: "b", Email: email, IsActive: true})
		require.NoError(t, err)
		for _, r := range roles {
			require.NoError(t, q.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: a.ID, RoleID: r.ID}))
		}
		return a.ID
	}
	a := mk("a@x.com", r1, r2)
	b := mk("b@x.com", r2)
	mk("c@x.com", r3)

	ids, err := DBStore{Q: q}.AdminIDsWithAnyRole(ctx, []uuid.UUID{r1.ID, r2.ID})
	require.NoError(t, err)
	assert.ElementsMatch(t, []uuid.UUID{a, b}, ids)
	ids, _ = DBStore{Q: q}.AdminIDsWithAnyRole(ctx, nil)
	assert.Empty(t, ids)

	// A condição usa o mesmo caminho.
	conds := NewConditionRegistry()
	RegisterStandardConditions(conds, DBStore{Q: q})
	c, _ := conds.Get("admin::has-same-role-as-creator")
	res, err := c.Handler(ctx, Subject{Type: SubjectAdmin, Roles: []Role{{ID: r2.ID}}})
	require.NoError(t, err)
	assert.True(t, res.filter.Match(record(map[string]any{"created_by_id": a})))
	assert.False(t, res.filter.Match(record(map[string]any{"created_by_id": uuid.New()})))
}

func TestInvalidacaoViaListenNotify(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := db.New(pool)
	role, _ := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "editor", Name: "Editor"})
	addPerm(t, q, role.ID, read, nil, "", `[]`)

	reg := NewRegistry()
	reg.MustRegister(Action{ID: read, Domain: DomainAdmin, Category: CategoryCollectionTypes, Name: "Ler"})
	newInstance := func() *Engine {
		return NewEngine(reg, NewConditionRegistry(), DBStore{Q: q}, PGBroadcaster{Pool: pool, Log: discardLog}, discardLog)
	}
	a, b := newInstance(), newInstance()
	go func() { _ = a.Listen(ctx) }()
	go func() { _ = b.Listen(ctx) }()

	// Espera as duas escutas (LISTEN) estarem ativas; ao conectar, cada uma invalida o cache uma vez,
	// então só depois disso o cache pode ser aquecido sem ser limpo por esse efeito.
	require.Eventually(t, func() bool {
		var n int
		require.NoError(t, pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE query LIKE 'LISTEN authz_invalidate%'").Scan(&n))
		return n == 2
	}, 5*time.Second, 50*time.Millisecond)
	time.Sleep(200 * time.Millisecond)

	s := Subject{Type: SubjectAdmin, ID: uuid.New(), Roles: []Role{{ID: role.ID, Code: "editor"}}}
	can := func(e *Engine) bool { d, err := e.Can(ctx, s, read); require.NoError(t, err); return d.Allowed }
	require.True(t, can(a))
	require.True(t, can(b))

	_, err := pool.Exec(ctx, "DELETE FROM admin_permissions")
	require.NoError(t, err)
	assert.True(t, can(b), "sem invalidação o cache da outra instância segue velho")

	// Dá tempo de a escuta (LISTEN) ser estabelecida antes de publicar.
	require.Eventually(t, func() bool {
		require.NoError(t, a.Invalidate(ctx, role.ID))
		return !can(b)
	}, 5*time.Second, 100*time.Millisecond, "a instância B deve receber a invalidação da A")
	assert.False(t, can(a), "a instância que alterou invalida imediatamente")
}

func TestPGBroadcasterPayloadGrande(t *testing.T) {
	pool := testutil.NewPool(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := PGBroadcaster{Pool: pool, Log: discardLog}
	got := make(chan []uuid.UUID, 10)
	go func() { _ = b.Subscribe(ctx, func(ids []uuid.UUID) { got <- ids }) }()
	<-got // chamada inicial (invalida tudo ao conectar)

	big := make([]uuid.UUID, maxNotifyIDs+1)
	for i := range big {
		big[i] = uuid.New()
	}
	require.NoError(t, b.Publish(ctx, big))
	select {
	case ids := <-got:
		assert.Empty(t, ids, "muitos IDs viram invalidação total")
	case <-time.After(5 * time.Second):
		t.Fatal("notificação não chegou")
	}
	one := []uuid.UUID{uuid.New()}
	require.NoError(t, b.Publish(ctx, one))
	assert.Equal(t, one, <-got)
}

func TestLocalBroadcaster(t *testing.T) {
	var b LocalBroadcaster
	require.NoError(t, b.Publish(context.Background(), nil))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error)
	go func() { done <- b.Subscribe(ctx, nil) }()
	cancel()
	assert.NoError(t, <-done)
}
