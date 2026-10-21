package authz

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStore struct {
	mu    sync.Mutex
	perms map[uuid.UUID][]Permission
	calls int
	err   error
}

func (f *fakeStore) RolePermissions(_ context.Context, id uuid.UUID) ([]Permission, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	return f.perms[id], f.err
}

func (f *fakeStore) set(id uuid.UUID, p ...Permission) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.perms[id] = p
}

type fakeMembers struct{ ids []uuid.UUID }

func (f fakeMembers) AdminIDsWithAnyRole(context.Context, []uuid.UUID) ([]uuid.UUID, error) {
	return f.ids, nil
}

type fixture struct {
	eng   *Engine
	store *fakeStore
}

func newFixture(t *testing.T, members RoleMembers) fixture {
	t.Helper()
	reg := NewRegistry()
	for _, v := range []string{"create", "read", "update", "delete", "publish"} {
		reg.MustRegister(Action{ID: "admin::content.document." + v, Domain: DomainAdmin, Category: CategoryCollectionTypes, Name: v, Resource: "document", SupportsFields: true, SupportsConditions: true})
	}
	conds := NewConditionRegistry()
	RegisterStandardConditions(conds, members)
	store := &fakeStore{perms: map[uuid.UUID][]Permission{}}
	return fixture{eng: NewEngine(reg, conds, store, LocalBroadcaster{}, slog.New(slog.NewJSONHandler(io.Discard, nil))), store: store}
}

func adminWith(roles ...Role) Subject {
	return Subject{Type: SubjectAdmin, ID: uuid.New(), Roles: roles}
}

func role(code string) Role { return Role{ID: uuid.New(), Code: code} }

const (
	read   = "admin::content.document.read"
	update = "admin::content.document.update"
	del    = "admin::content.document.delete"
)

func record(m map[string]any) Getter {
	return func(f string) (any, bool) { v, ok := m[f]; return v, ok }
}

func TestAcaoNaoRegistradaNega(t *testing.T) {
	f := newFixture(t, nil)
	r := role("editor")
	f.store.set(r.ID, Permission{Action: "admin::content.fantasma.read"})
	d, err := f.eng.Can(context.Background(), adminWith(r), "admin::content.fantasma.read")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
	assert.True(t, d.Filter.IsFalse())

	// Nem o Super Admin passa em ação não registrada.
	d, err = f.eng.Can(context.Background(), adminWith(role(SuperAdminRoleCode)), "admin::content.fantasma.read")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
}

func TestSuperAdminCuringa(t *testing.T) {
	f := newFixture(t, nil)
	d, err := f.eng.Can(context.Background(), adminWith(role(SuperAdminRoleCode)), update)
	require.NoError(t, err)
	assert.True(t, d.Allowed)
	assert.True(t, d.Filter.IsTrue())
	assert.True(t, d.Fields.IsAll())
	assert.Zero(t, f.store.calls, "Super Admin não consulta permissões")
}

func TestUniaoEntreRoles(t *testing.T) {
	f := newFixture(t, nil)
	r1, r2 := role("a"), role("b")
	f.store.set(r1.ID, Permission{Action: read, Fields: []string{"titulo"}})
	f.store.set(r2.ID, Permission{Action: read, Fields: []string{"conteudo"}}, Permission{Action: del})
	s := adminWith(r1, r2)

	d, err := f.eng.Can(context.Background(), s, read)
	require.NoError(t, err)
	assert.True(t, d.Allowed)
	assert.Equal(t, []string{"conteudo", "titulo"}, d.Fields.Names())

	d, _ = f.eng.Can(context.Background(), s, del)
	assert.True(t, d.Allowed, "a permissão vem da segunda role")
	d, _ = f.eng.Can(context.Background(), s, update)
	assert.False(t, d.Allowed)

	// Uma role sem restrição de campos prevalece (nil = todos).
	f.store.set(r2.ID, Permission{Action: read})
	f.eng.cache.invalidate()
	d, _ = f.eng.Can(context.Background(), s, read)
	assert.True(t, d.Fields.IsAll())
}

func TestCondicoesCombinadasComOR(t *testing.T) {
	dept := uuid.New()
	f := newFixture(t, nil)
	r := role("author")
	f.store.set(r.ID, Permission{Action: update, Conditions: []string{"admin::is-creator", "admin::same-department"}})
	s := adminWith(r)
	s.DepartmentID = &dept

	d, err := f.eng.Can(context.Background(), s, update)
	require.NoError(t, err)
	require.True(t, d.Allowed)
	assert.False(t, d.Filter.IsTrue())

	own := record(map[string]any{"created_by_id": s.ID, "department_id": uuid.New()})
	sameDept := record(map[string]any{"created_by_id": uuid.New(), "department_id": dept})
	neither := record(map[string]any{"created_by_id": uuid.New(), "department_id": uuid.New()})
	assert.True(t, d.AllowsRecord(own))
	assert.True(t, d.AllowsRecord(sameDept))
	assert.False(t, d.AllowsRecord(neither))
}

func TestCondicaoQueRetornaFiltro(t *testing.T) {
	f := newFixture(t, nil)
	r := role("author")
	f.store.set(r.ID, Permission{Action: update, Conditions: []string{"admin::is-creator"}})
	s := adminWith(r)
	d, err := f.eng.Can(context.Background(), s, update)
	require.NoError(t, err)

	sql, args, err := d.Filter.ToSQL(func(n string) (string, bool) { return n, n == "created_by_id" }, 1)
	require.NoError(t, err)
	assert.Equal(t, "created_by_id = $1", sql)
	assert.Equal(t, []any{s.ID}, args)
}

func TestCondicaoNegaEDesconhecida(t *testing.T) {
	f := newFixture(t, nil)
	r := role("x")
	s := adminWith(r) // sem departamento
	f.store.set(r.ID, Permission{Action: update, Conditions: []string{"admin::same-department"}})
	d, _ := f.eng.Can(context.Background(), s, update)
	assert.False(t, d.Allowed, "sem departamento a condição nega")

	f.store.set(r.ID, Permission{Action: update, Conditions: []string{"admin::nao-existe"}})
	f.eng.cache.invalidate()
	d, _ = f.eng.Can(context.Background(), s, update)
	assert.False(t, d.Allowed, "condição desconhecida nega")
}

func TestPermissaoIncondicionalSomeComOFiltro(t *testing.T) {
	f := newFixture(t, nil)
	r1, r2 := role("a"), role("b")
	f.store.set(r1.ID, Permission{Action: update, Conditions: []string{"admin::is-creator"}})
	f.store.set(r2.ID, Permission{Action: update})
	d, _ := f.eng.Can(context.Background(), adminWith(r1, r2), update)
	assert.True(t, d.Filter.IsTrue())
}

func TestCamposPorRegistro(t *testing.T) {
	// Revisor: edita só status/observacao dos documentos do departamento;
	// autor: edita tudo do que criou.
	dept := uuid.New()
	f := newFixture(t, nil)
	rev, aut := role("revisor"), role("author")
	f.store.set(rev.ID, Permission{Action: update, Fields: []string{"status", "observacao_interna"}, Conditions: []string{"admin::same-department"}})
	f.store.set(aut.ID, Permission{Action: update, Conditions: []string{"admin::is-creator"}})
	s := adminWith(rev, aut)
	s.DepartmentID = &dept
	d, _ := f.eng.Can(context.Background(), s, update)

	deptDoc := record(map[string]any{"created_by_id": uuid.New(), "department_id": dept})
	ownDoc := record(map[string]any{"created_by_id": s.ID, "department_id": uuid.New()})
	assert.Equal(t, []string{"observacao_interna", "status"}, d.FieldsFor(deptDoc).Names())
	assert.True(t, d.FieldsFor(ownDoc).IsAll())
	assert.False(t, d.FieldsFor(record(map[string]any{})).Allows("titulo"))
}

func TestTiposDeSujeitoEDominio(t *testing.T) {
	f := newFixture(t, nil)
	r := role("x")
	f.store.set(r.ID, Permission{Action: read})
	for _, typ := range []SubjectType{SubjectAnonymous, SubjectEndUser, SubjectAPIToken} {
		d, err := f.eng.Can(context.Background(), Subject{Type: typ, Roles: []Role{r}}, read)
		require.NoError(t, err)
		assert.False(t, d.Allowed, typ)
	}
	// Um "super-admin" que não é administrador não ganha curinga.
	assert.False(t, Subject{Type: SubjectEndUser, Roles: []Role{role(SuperAdminRoleCode)}}.IsSuperAdmin())
}

func TestFiltroPorRecurso(t *testing.T) {
	f := newFixture(t, nil)
	r := role("x")
	f.store.set(r.ID, Permission{Action: read, Subject: "document", Fields: []string{"titulo"}})
	d, _ := f.eng.Can(context.Background(), adminWith(r), read, "document")
	assert.True(t, d.Allowed)
	d, _ = f.eng.Can(context.Background(), adminWith(r), read, "outro")
	assert.False(t, d.Allowed)
	d, _ = f.eng.Can(context.Background(), adminWith(r), read)
	assert.True(t, d.Allowed)
}

func TestHasSameRoleAsCreator(t *testing.T) {
	colleague, other := uuid.New(), uuid.New()
	f := newFixture(t, fakeMembers{ids: []uuid.UUID{colleague, other}})
	r := role("x")
	f.store.set(r.ID, Permission{Action: read, Conditions: []string{"admin::has-same-role-as-creator"}})
	d, _ := f.eng.Can(context.Background(), adminWith(r), read)
	assert.True(t, d.AllowsRecord(record(map[string]any{"created_by_id": colleague})))
	assert.False(t, d.AllowsRecord(record(map[string]any{"created_by_id": uuid.New()})))

	empty := newFixture(t, fakeMembers{})
	empty.store.set(r.ID, Permission{Action: read, Conditions: []string{"admin::has-same-role-as-creator"}})
	d, _ = empty.eng.Can(context.Background(), adminWith(r), read)
	assert.False(t, d.Allowed, "sem colegas, a condição nega")
}

func TestCacheEInvalidacao(t *testing.T) {
	f := newFixture(t, nil)
	r := role("x")
	s := adminWith(r)
	f.store.set(r.ID, Permission{Action: read})

	for range 3 {
		d, _ := f.eng.Can(context.Background(), s, read)
		assert.True(t, d.Allowed)
	}
	assert.Equal(t, 1, f.store.calls, "permissões ficam em cache")

	f.store.set(r.ID)
	d, _ := f.eng.Can(context.Background(), s, read)
	assert.True(t, d.Allowed, "sem invalidação, o cache continua valendo")

	require.NoError(t, f.eng.Invalidate(context.Background(), r.ID))
	d, _ = f.eng.Can(context.Background(), s, read)
	assert.False(t, d.Allowed, "após invalidar, a mudança vale na hora")

	// Invalidar tudo.
	f.store.set(r.ID, Permission{Action: read})
	require.NoError(t, f.eng.Invalidate(context.Background()))
	d, _ = f.eng.Can(context.Background(), s, read)
	assert.True(t, d.Allowed)
}

func TestErroDoStorePropaga(t *testing.T) {
	f := newFixture(t, nil)
	f.store.err = errors.New("banco fora")
	_, err := f.eng.Can(context.Background(), adminWith(role("x")), read)
	assert.Error(t, err)
	_, err = f.eng.Effective(context.Background(), adminWith(role("x")))
	assert.Error(t, err)
}

type recordingBus struct {
	published [][]uuid.UUID
	handler   func([]uuid.UUID)
	ready     chan struct{}
}

func (b *recordingBus) Publish(_ context.Context, ids []uuid.UUID) error {
	b.published = append(b.published, ids)
	return nil
}

func (b *recordingBus) Subscribe(ctx context.Context, fn func([]uuid.UUID)) error {
	b.handler = fn
	close(b.ready)
	<-ctx.Done()
	return nil
}

func TestInvalidacaoEntreInstancias(t *testing.T) {
	f := newFixture(t, nil)
	bus := &recordingBus{ready: make(chan struct{})}
	f.eng.bus = bus
	r := role("x")
	s := adminWith(r)
	f.store.set(r.ID, Permission{Action: read})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = f.eng.Listen(ctx); close(done) }()
	<-bus.ready

	d, _ := f.eng.Can(ctx, s, read)
	require.True(t, d.Allowed)
	f.store.set(r.ID)
	bus.handler([]uuid.UUID{r.ID}) // outra instância avisou
	d, _ = f.eng.Can(ctx, s, read)
	assert.False(t, d.Allowed)

	require.NoError(t, f.eng.Invalidate(ctx, r.ID))
	assert.Equal(t, [][]uuid.UUID{{r.ID}}, bus.published, "a invalidação local é publicada")
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Listen não terminou")
	}
}

func TestPermissoesEfetivas(t *testing.T) {
	f := newFixture(t, nil)
	r1, r2 := role("a"), role("b")
	f.store.set(r1.ID,
		Permission{Action: read, Subject: "document", Fields: []string{"titulo"}, Conditions: []string{"admin::is-creator"}},
		Permission{Action: "admin::content.fantasma.read"})
	f.store.set(r2.ID, Permission{Action: read, Subject: "document", Fields: []string{"conteudo"}, Conditions: []string{"admin::same-department"}})

	got, err := f.eng.Effective(context.Background(), adminWith(r1, r2))
	require.NoError(t, err)
	require.Len(t, got, 1, "ação não registrada é ignorada")
	assert.Equal(t, []string{"conteudo", "titulo"}, got[0].Fields)
	assert.ElementsMatch(t, []string{"admin::is-creator", "admin::same-department"}, got[0].Conditions)

	f.store.set(r2.ID, Permission{Action: read, Subject: "document"})
	f.eng.cache.invalidate()
	got, _ = f.eng.Effective(context.Background(), adminWith(r1, r2))
	assert.Nil(t, got[0].Fields, "sem restrição de campos prevalece")
	assert.Empty(t, got[0].Conditions, "sem condição prevalece")

	su, _ := f.eng.Effective(context.Background(), adminWith(role(SuperAdminRoleCode)))
	assert.Len(t, su, 5)
	none, _ := f.eng.Effective(context.Background(), Subject{Type: SubjectAnonymous})
	assert.Empty(t, none)
}

func newAPIFixture(t *testing.T) (fixture, uuid.UUID, uuid.UUID) {
	t.Helper()
	f := newFixture(t, nil)
	f.eng.actions.MustRegister(
		Action{ID: "api::document.find", Domain: DomainAPI, Category: CategoryApplication, Name: "Listar", Method: "GET", Route: "/api/documents"},
		Action{ID: "api::document.create", Domain: DomainAPI, Category: CategoryApplication, Name: "Criar", Method: "POST", Route: "/api/documents"},
	)
	apiStore := &fakeStore{perms: map[uuid.UUID][]Permission{}}
	f.eng.SetEndUserStore(apiStore)
	public, authed := uuid.New(), uuid.New()
	apiStore.set(public, Permission{Action: "api::document.find"})
	apiStore.set(authed, Permission{Action: "api::document.find"}, Permission{Action: "api::document.create"})
	return f, public, authed
}

func TestDominioAPIUsuarioFinalEAnonimo(t *testing.T) {
	f, public, authed := newAPIFixture(t)
	ctx := context.Background()
	anon := Subject{Type: SubjectAnonymous, Roles: []Role{{ID: public, Code: "public"}}}
	user := Subject{Type: SubjectEndUser, ID: uuid.New(), Roles: []Role{{ID: authed, Code: "authenticated"}}}

	d, err := f.eng.Can(ctx, anon, "api::document.find")
	require.NoError(t, err)
	assert.True(t, d.Allowed)
	assert.True(t, d.Filter.IsTrue())
	assert.True(t, d.Fields.IsAll(), "usuário final não tem restrição de campos")
	d, _ = f.eng.Can(ctx, anon, "api::document.create")
	assert.False(t, d.Allowed, "o anônimo não cria")
	d, _ = f.eng.Can(ctx, user, "api::document.create")
	assert.True(t, d.Allowed)
	d, _ = f.eng.Can(ctx, user, "api::document.delete")
	assert.False(t, d.Allowed, "ação não registrada")

	// Cada sujeito só vale no seu domínio.
	d, _ = f.eng.Can(ctx, user, read)
	assert.False(t, d.Allowed, "usuário final não acessa ações de admin")
	d, _ = f.eng.Can(ctx, adminWith(role("x")), "api::document.find")
	assert.False(t, d.Allowed, "administrador não herda permissões de usuário final")
	d, _ = f.eng.Can(ctx, Subject{Type: SubjectAPIToken, Roles: []Role{{ID: authed}}}, "api::document.find")
	assert.False(t, d.Allowed, "token de API: Fase 6")
	d, _ = f.eng.Can(ctx, Subject{Type: SubjectEndUser}, "api::document.find")
	assert.False(t, d.Allowed, "sem role")
	d, _ = f.eng.Can(ctx, Subject{Type: SubjectEndUser, Roles: []Role{{ID: authed}, {ID: public}}}, "api::document.find")
	assert.False(t, d.Allowed, "usuário final tem uma única role")
}

func TestDominioAPISemStoreNega(t *testing.T) {
	f := newFixture(t, nil)
	f.eng.actions.MustRegister(Action{ID: "api::document.find", Domain: DomainAPI, Category: CategoryApplication, Name: "Listar", Method: "GET", Route: "/api/documents"})
	d, err := f.eng.Can(context.Background(), Subject{Type: SubjectAnonymous, Roles: []Role{role("public")}}, "api::document.find")
	require.NoError(t, err)
	assert.False(t, d.Allowed)
}

func TestDominioAPICacheEInvalidacao(t *testing.T) {
	f, public, _ := newAPIFixture(t)
	ctx := context.Background()
	anon := Subject{Type: SubjectAnonymous, Roles: []Role{{ID: public}}}
	d, _ := f.eng.Can(ctx, anon, "api::document.find")
	require.True(t, d.Allowed)
	f.eng.apiCache.store.(*fakeStore).set(public)
	d, _ = f.eng.Can(ctx, anon, "api::document.find")
	assert.True(t, d.Allowed, "cache")
	require.NoError(t, f.eng.Invalidate(ctx, public))
	d, _ = f.eng.Can(ctx, anon, "api::document.find")
	assert.False(t, d.Allowed)

	f.eng.apiCache.store.(*fakeStore).err = errors.New("banco fora")
	require.NoError(t, f.eng.Invalidate(ctx))
	_, err := f.eng.Can(ctx, anon, "api::document.find")
	assert.Error(t, err)
}

func TestTokensDeAPI(t *testing.T) {
	f, _, _ := newAPIFixture(t)
	f.eng.actions.MustRegister(Action{ID: "api::document.update", Domain: DomainAPI, Category: CategoryApplication, Name: "Atualizar", Method: "PUT", Route: "/api/documents/{id}"},
		Action{ID: "api::document.head", Domain: DomainAPI, Category: CategoryApplication, Name: "Cabeçalho", Method: "HEAD", Route: "/api/documents/{id}"})
	ctx := context.Background()
	tok := func(typ string, actions ...string) Subject {
		return Subject{Type: SubjectAPIToken, ID: uuid.New(), APIToken: &TokenGrant{Type: typ, Actions: actions}}
	}
	tests := []struct {
		nome   string
		s      Subject
		action string
		want   bool
	}{
		{"full-access: leitura", tok(TokenFullAccess), "api::document.find", true},
		{"full-access: escrita", tok(TokenFullAccess), "api::document.create", true},
		{"full-access: ação não registrada", tok(TokenFullAccess), "api::document.delete", false},
		{"full-access: ação de admin", tok(TokenFullAccess), read, false},
		{"read-only: GET", tok(TokenReadOnly), "api::document.find", true},
		{"read-only: HEAD", tok(TokenReadOnly), "api::document.head", true},
		{"read-only: POST", tok(TokenReadOnly), "api::document.create", false},
		{"read-only: PUT", tok(TokenReadOnly), "api::document.update", false},
		{"custom: na lista", tok(TokenCustom, "api::document.create"), "api::document.create", true},
		{"custom: fora da lista", tok(TokenCustom, "api::document.create"), "api::document.find", false},
		{"custom: lista vazia", tok(TokenCustom), "api::document.find", false},
		{"custom: ação removida do código", tok(TokenCustom, "api::document.delete"), "api::document.delete", false},
		{"tipo desconhecido", tok("outro"), "api::document.find", false},
		{"sem grant", Subject{Type: SubjectAPIToken}, "api::document.find", false},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			d, err := f.eng.Can(ctx, tt.s, tt.action)
			require.NoError(t, err)
			assert.Equal(t, tt.want, d.Allowed)
			if tt.want {
				assert.True(t, d.Filter.IsTrue())
				assert.True(t, d.Fields.IsAll())
			}
		})
	}
	// O token não tem role: não herda permissões de roles nem funciona em engine sem store de usuário final.
	d, _ := f.eng.Can(ctx, Subject{Type: SubjectAPIToken, Roles: []Role{{ID: uuid.New()}}}, "api::document.find")
	assert.False(t, d.Allowed)
}
