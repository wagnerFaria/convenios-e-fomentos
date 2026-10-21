package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/app"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/config"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/testutil"
)

const senha = "senha-super-forte-1"

type env struct {
	t     *testing.T
	h     http.Handler
	pool  *pgxpool.Pool
	mail  *testutil.CaptureMailer
	cfg   *config.Config
	authz *authz.Engine
	app   *app.App
}

type resp struct {
	Status int
	Body   map[string]any
}

func (r resp) data() map[string]any {
	d, _ := r.Body["data"].(map[string]any)
	return d
}

func (r resp) errCode() string {
	e, _ := r.Body["error"].(map[string]any)
	c, _ := e["code"].(string)
	return c
}

func newEnv(t *testing.T, mutate ...func(*config.Config)) *env {
	t.Helper()
	pool := testutil.NewPool(t)
	cfg := &config.Config{
		AppEnv: "test", HTTPPort: 8080,
		JWTAdminSecret: "admin-secret-de-teste-com-32-caracteres!", JWTAPISecret: "api-secret-de-teste-com-32-caracteres!!!",
		APITokenSalt:   "salt-de-teste-com-32-caracteres-ok-123",
		AccessTokenTTL: 15 * time.Minute, RefreshTokenMaxTTL: time.Hour, RefreshTokenIdleTTL: 30 * time.Minute,
		InviteTokenTTL: time.Hour, ResetTokenTTL: 10 * time.Minute, EmailConfirmationTTL: time.Hour,
		RateLimitAuthPerMin: 1000, MaxBodyBytes: 1 << 20, PasswordMinLength: 8,
		Argon2MemoryKiB: 64, Argon2Iterations: 1, Argon2Parallel: 1,
		PublicURL: "http://app.test", CORSAllowedOrigins: []string{"http://localhost:3000"},
	}
	for _, m := range mutate {
		m(cfg)
	}
	mail := &testutil.CaptureMailer{}
	a, err := app.New(cfg, slog.New(slog.NewJSONHandler(io.Discard, nil)), pool, mail)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	require.NoError(t, a.Start(ctx))
	return &env{t: t, h: a.Handler, pool: pool, mail: mail, cfg: cfg, authz: a.Authz, app: a}
}

func (e *env) do(method, path string, body any, token string) resp {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		rd = bytes.NewReader(b)
	}
	req := httptest.NewRequestWithContext(context.Background(), method, path, rd)
	req.RemoteAddr = "10.0.0.1:1234"
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, req)
	out := resp{Status: rec.Code}
	if rec.Body.Len() > 0 {
		require.NoError(e.t, json.Unmarshal(rec.Body.Bytes(), &out.Body), rec.Body.String())
	}
	return out
}

func (e *env) registerFirst() {
	e.t.Helper()
	r := e.do("POST", "/admin/auth/register-first", map[string]any{"firstName": "Ana", "lastName": "Admin", "email": "ana@exemplo.com", "password": senha}, "")
	require.Equal(e.t, 201, r.Status, r.Body)
}

func (e *env) login(email, password string) resp {
	return e.do("POST", "/admin/auth/login", map[string]any{"email": email, "password": password}, "")
}

func (e *env) loginOK(email, password string) (access, refresh string) {
	e.t.Helper()
	r := e.login(email, password)
	require.Equal(e.t, 200, r.Status, r.Body)
	return r.data()["accessToken"].(string), r.data()["refreshToken"].(string)
}

func TestRegistroInicialELogin(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()

	r := e.do("POST", "/admin/auth/register-first", map[string]any{"firstName": "B", "lastName": "B", "email": "b@exemplo.com", "password": senha}, "")
	assert.Equal(t, 403, r.Status)
	assert.Equal(t, "setup_completed", r.errCode())

	access, _ := e.loginOK("ANA@exemplo.com", senha) // e-mail sem diferenciar maiúsculas
	me := e.do("GET", "/admin/auth/me", nil, access)
	require.Equal(t, 200, me.Status)
	assert.Equal(t, "ana@exemplo.com", me.data()["email"])
	roles := me.data()["roles"].([]any)
	require.Len(t, roles, 1)
	assert.Equal(t, "super-admin", roles[0].(map[string]any)["code"])

	// Credenciais inválidas: mesma resposta para senha errada e e-mail inexistente.
	wrong := e.login("ana@exemplo.com", "errada-errada")
	ghost := e.login("ninguem@exemplo.com", "errada-errada")
	assert.Equal(t, 401, wrong.Status)
	assert.Equal(t, wrong.Body, ghost.Body)

	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, "").Status)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, "lixo").Status)
	assert.Equal(t, 422, e.login("nao-e-email", "x").Status)
}

func TestJWTDeOutraAudienciaERejeitado(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	a, err := db.New(e.pool).GetAdminByEmail(context.Background(), "ana@exemplo.com")
	require.NoError(t, err)

	apiTok, err := auth.NewJWTIssuer(e.cfg.JWTAPISecret, auth.AudienceAPI, time.Minute).Issue(a.ID, uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, apiTok).Status)

	// Mesmo com o segredo de admin, a audiência "api" é rejeitada.
	wrongAud, _ := auth.NewJWTIssuer(e.cfg.JWTAdminSecret, auth.AudienceAPI, time.Minute).Issue(a.ID, uuid.New())
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, wrongAud).Status)

	// Token válido, porém com sessão inexistente.
	noSession, _ := auth.NewJWTIssuer(e.cfg.JWTAdminSecret, auth.AudienceAdmin, time.Minute).Issue(a.ID, uuid.New())
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, noSession).Status)
}

func TestRefreshComRotacaoEDeteccaoDeReuso(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	access1, refresh1 := e.loginOK("ana@exemplo.com", senha)

	r := e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh1}, "")
	require.Equal(t, 200, r.Status, r.Body)
	access2, refresh2 := r.data()["accessToken"].(string), r.data()["refreshToken"].(string)
	assert.NotEqual(t, refresh1, refresh2)
	// A sessão (família) é a mesma: o access token antigo segue válido até expirar.
	assert.Equal(t, 200, e.do("GET", "/admin/auth/me", nil, access1).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/auth/me", nil, access2).Status)

	// Reuso do refresh token antigo: nega e revoga toda a família.
	reuse := e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh1}, "")
	assert.Equal(t, 401, reuse.Status)
	assert.Equal(t, "invalid_refresh_token", reuse.errCode())
	assert.Equal(t, 401, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh2}, "").Status, "o token novo também deve ter sido revogado")
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, access2).Status, "o access token da família revogada deixa de valer")

	// Token desconhecido.
	assert.Equal(t, 401, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": "inexistente"}, "").Status)

	// Reuso de uma sessão não afeta as outras sessões do usuário.
	_, other := e.loginOK("ana@exemplo.com", senha)
	assert.Equal(t, 200, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": other}, "").Status)
}

func TestRefreshExpiradoPorInatividade(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	_, refresh := e.loginOK("ana@exemplo.com", senha)
	_, err := e.pool.Exec(context.Background(), "UPDATE admin_sessions SET expires_at = now() - interval '1 second'")
	require.NoError(t, err)
	assert.Equal(t, 401, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh}, "").Status)
}

func TestRefreshRespeitaSessaoMaxima(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	_, refresh := e.loginOK("ana@exemplo.com", senha)
	// A família começou antes do tempo máximo de sessão.
	_, err := e.pool.Exec(context.Background(), "UPDATE admin_sessions SET family_started_at = now() - interval '2 hours'")
	require.NoError(t, err)
	assert.Equal(t, 401, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh}, "").Status)
}

func TestSessoes(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	a1, _ := e.loginOK("ana@exemplo.com", senha)
	a2, _ := e.loginOK("ana@exemplo.com", senha)

	list := e.do("GET", "/admin/auth/sessions", nil, a1)
	require.Equal(t, 200, list.Status)
	items := list.Body["data"].([]any)
	require.Len(t, items, 2)
	var otherID string
	for _, it := range items {
		m := it.(map[string]any)
		if m["current"] == false {
			otherID = m["id"].(string)
		}
	}
	require.NotEmpty(t, otherID)

	assert.Equal(t, 204, e.do("DELETE", "/admin/auth/sessions/"+otherID, nil, a1).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/auth/sessions/"+otherID, nil, a1).Status)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, a2).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/auth/me", nil, a1).Status)

	// Revogar todas, preservando a atual.
	e.loginOK("ana@exemplo.com", senha)
	assert.Equal(t, 204, e.do("DELETE", "/admin/auth/sessions?keepCurrent=true", nil, a1).Status)
	assert.Len(t, e.do("GET", "/admin/auth/sessions", nil, a1).Body["data"], 1)

	// Logout invalida a sessão atual.
	assert.Equal(t, 204, e.do("POST", "/admin/auth/logout", nil, a1).Status)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, a1).Status)
}

func TestSessaoDeOutroAdminNaoPodeSerRevogada(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	_, err := db.New(e.pool).UpsertAdminRole(context.Background(), db.UpsertAdminRoleParams{Code: "editor", Name: "Editor"})
	require.NoError(t, err)
	a1, _ := e.loginOK("ana@exemplo.com", senha)
	sess := e.do("GET", "/admin/auth/sessions", nil, a1).Body["data"].([]any)[0].(map[string]any)["id"].(string)

	other := inviteAndAccept(t, e, a1, "beto@exemplo.com", "editor")
	assert.Equal(t, 404, e.do("DELETE", "/admin/auth/sessions/"+sess, nil, other).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/auth/me", nil, a1).Status)
}

// inviteAndAccept convida um administrador com a role (por code) e devolve seu access token.
func inviteAndAccept(t *testing.T, e *env, inviterToken, email, roleCode string) string {
	t.Helper()
	var roleID uuid.UUID
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT id FROM admin_roles WHERE code=$1", roleCode).Scan(&roleID))
	inv := e.do("POST", "/admin/users/invite", map[string]any{"firstName": "Novo", "lastName": "Admin", "email": email, "roleIds": []string{roleID.String()}}, inviterToken)
	require.Equal(t, 201, inv.Status, inv.Body)
	url := inv.data()["inviteUrl"].(string)
	token := url[len("http://app.test/admin/accept-invite?token="):]
	acc := e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": token, "password": senha}, "")
	require.Equal(t, 200, acc.Status, acc.Body)
	return acc.data()["accessToken"].(string)
}

func TestConvite(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	_, err := db.New(e.pool).UpsertAdminRole(context.Background(), db.UpsertAdminRoleParams{Code: "author", Name: "Author"})
	require.NoError(t, err)
	boss, _ := e.loginOK("ana@exemplo.com", senha)

	var roleID uuid.UUID
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT id FROM admin_roles WHERE code='author'").Scan(&roleID))
	body := map[string]any{"firstName": "Beto", "lastName": "Autor", "email": "beto@exemplo.com", "roleIds": []string{roleID.String()}}

	assert.Equal(t, 401, e.do("POST", "/admin/users/invite", body, "").Status)

	inv := e.do("POST", "/admin/users/invite", body, boss)
	require.Equal(t, 201, inv.Status, inv.Body)
	assert.Equal(t, false, inv.data()["admin"].(map[string]any)["isActive"])
	url := inv.data()["inviteUrl"].(string)
	token := url[len("http://app.test/admin/accept-invite?token="):]

	// A conta fica inativa até o aceite.
	assert.Equal(t, 401, e.login("beto@exemplo.com", senha).Status)

	// Duplicado e role inexistente.
	assert.Equal(t, 409, e.do("POST", "/admin/users/invite", body, boss).Status)
	bad := map[string]any{"firstName": "C", "lastName": "C", "email": "c@exemplo.com", "roleIds": []string{uuid.NewString()}}
	assert.Equal(t, 422, e.do("POST", "/admin/users/invite", bad, boss).Status)
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM admins WHERE email='c@exemplo.com'").Scan(&n))
	assert.Zero(t, n, "convite com role inválida não pode deixar o admin criado")

	// Aceite: senha fraca, token errado, sucesso, reuso.
	assert.Equal(t, 422, e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": token, "password": "curta"}, "").Status)
	assert.Equal(t, 400, e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": "errado", "password": senha}, "").Status)
	acc := e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": token, "password": senha}, "")
	require.Equal(t, 200, acc.Status, acc.Body)
	assert.Equal(t, 400, e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": token, "password": senha}, "").Status)

	me := e.do("GET", "/admin/auth/me", nil, acc.data()["accessToken"].(string))
	assert.Equal(t, "author", me.data()["roles"].([]any)[0].(map[string]any)["code"])
	assert.Equal(t, 200, e.login("beto@exemplo.com", senha).Status)

	// Sem permissão: quem não é Super Admin não convida (substituído pelo motor na Fase 3).
	assert.Equal(t, 403, e.do("POST", "/admin/users/invite", map[string]any{"firstName": "D", "lastName": "D", "email": "d@exemplo.com", "roleIds": []string{roleID.String()}}, acc.data()["accessToken"].(string)).Status)

	// Convite expirado.
	inv2 := e.do("POST", "/admin/users/invite", map[string]any{"firstName": "E", "lastName": "E", "email": "e@exemplo.com", "roleIds": []string{roleID.String()}}, boss)
	tok2 := inv2.data()["inviteUrl"].(string)[len("http://app.test/admin/accept-invite?token="):]
	_, err = e.pool.Exec(context.Background(), "UPDATE one_time_tokens SET expires_at = now() - interval '1 minute'")
	require.NoError(t, err)
	assert.Equal(t, 400, e.do("POST", "/admin/auth/accept-invite", map[string]any{"token": tok2, "password": senha}, "").Status)
}

func TestRedefinicaoDeSenha(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	access, _ := e.loginOK("ana@exemplo.com", senha)

	// E-mail inexistente: mesma resposta e nenhum e-mail enviado.
	assert.Equal(t, 204, e.do("POST", "/admin/auth/forgot-password", map[string]any{"email": "ghost@exemplo.com"}, "").Status)
	assert.Zero(t, e.mail.Count())

	assert.Equal(t, 204, e.do("POST", "/admin/auth/forgot-password", map[string]any{"email": "ana@exemplo.com"}, "").Status)
	require.Equal(t, 1, e.mail.Count())
	token := e.mail.LastToken()
	require.NotEmpty(t, token)

	assert.Equal(t, 422, e.do("POST", "/admin/auth/reset-password", map[string]any{"token": token, "password": "curta"}, "").Status)
	assert.Equal(t, 400, e.do("POST", "/admin/auth/reset-password", map[string]any{"token": "x", "password": "nova-senha-123"}, "").Status)
	assert.Equal(t, 204, e.do("POST", "/admin/auth/reset-password", map[string]any{"token": token, "password": "nova-senha-123"}, "").Status)
	assert.Equal(t, 400, e.do("POST", "/admin/auth/reset-password", map[string]any{"token": token, "password": "outra-senha-123"}, "").Status, "uso único")

	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, access).Status, "sessões antigas são revogadas")
	assert.Equal(t, 401, e.login("ana@exemplo.com", senha).Status)
	assert.Equal(t, 200, e.login("ana@exemplo.com", "nova-senha-123").Status)

	// Um novo pedido invalida o token anterior.
	e.do("POST", "/admin/auth/forgot-password", map[string]any{"email": "ana@exemplo.com"}, "")
	first := e.mail.LastToken()
	e.do("POST", "/admin/auth/forgot-password", map[string]any{"email": "ana@exemplo.com"}, "")
	assert.Equal(t, 400, e.do("POST", "/admin/auth/reset-password", map[string]any{"token": first, "password": "mais-uma-senha-1"}, "").Status)
}

func TestAdminBloqueadoOuInativoPerdeAcessoImediatamente(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	access, refresh := e.loginOK("ana@exemplo.com", senha)

	_, err := e.pool.Exec(context.Background(), "UPDATE admins SET blocked = true")
	require.NoError(t, err)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, access).Status)
	assert.Equal(t, 401, e.do("POST", "/admin/auth/refresh", map[string]any{"refreshToken": refresh}, "").Status)
	assert.Equal(t, 401, e.login("ana@exemplo.com", senha).Status)
	assert.Zero(t, e.mailAfterForgot("ana@exemplo.com"), "bloqueado não recebe e-mail de redefinição")
}

func (e *env) mailAfterForgot(email string) int {
	e.do("POST", "/admin/auth/forgot-password", map[string]any{"email": email}, "")
	return e.mail.Count()
}

func TestLimiteDeRequisicoes(t *testing.T) {
	e := newEnv(t, func(c *config.Config) { c.RateLimitAuthPerMin = 3 })
	e.registerFirst()
	for range 3 {
		assert.Equal(t, 401, e.login("ana@exemplo.com", "errada-errada").Status)
	}
	r := e.login("ana@exemplo.com", "errada-errada")
	assert.Equal(t, 429, r.Status)
	assert.Equal(t, "rate_limited", r.errCode())
	// Outro identificador do mesmo IP tem contagem própria.
	assert.Equal(t, 401, e.login("outro@exemplo.com", "errada-errada").Status)
}

func TestAuditoriaDoConviteERegistroInicial(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM audit_logs WHERE action='admin.create-super-admin'").Scan(&n))
	assert.Equal(t, 1, n)
}

func TestRegistroInicialConcorrente(t *testing.T) {
	e := newEnv(t)
	results := make(chan int, 5)
	for i := range 5 {
		go func() {
			r := e.do("POST", "/admin/auth/register-first", map[string]any{"firstName": "A", "lastName": "B", "email": "x" + string(rune('a'+i)) + "@exemplo.com", "password": senha}, "")
			results <- r.Status
		}()
	}
	created := 0
	for range 5 {
		if <-results == 201 {
			created++
		}
	}
	assert.Equal(t, 1, created, "apenas um registro inicial pode vencer a corrida")
}

// grant dá à role (por code) a permissão e invalida o cache do motor.
func (e *env) grant(roleCode, action string, fields, conds string) {
	e.t.Helper()
	ctx := context.Background()
	var roleID uuid.UUID
	require.NoError(e.t, e.pool.QueryRow(ctx, "SELECT id FROM admin_roles WHERE code=$1", roleCode).Scan(&roleID))
	var f []byte
	if fields != "" {
		f = []byte(fields)
	}
	_, err := db.New(e.pool).CreateAdminPermission(ctx, db.CreateAdminPermissionParams{RoleID: roleID, Action: action, Fields: f, Conditions: []byte(conds), Properties: []byte(`{}`)})
	require.NoError(e.t, err)
	require.NoError(e.t, e.authz.Invalidate(ctx, roleID))
}

func TestPermissaoParaConvidarViaMotor(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	boss, _ := e.loginOK("ana@exemplo.com", senha)
	_, err := db.New(e.pool).UpsertAdminRole(context.Background(), db.UpsertAdminRoleParams{Code: "rh", Name: "RH"})
	require.NoError(t, err)
	rh := inviteAndAccept(t, e, boss, "rh@exemplo.com", "rh")

	var roleID uuid.UUID
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT id FROM admin_roles WHERE code='rh'").Scan(&roleID))
	body := map[string]any{"firstName": "N", "lastName": "N", "email": "novo@exemplo.com", "roleIds": []string{roleID.String()}}

	assert.Equal(t, 403, e.do("POST", "/admin/users/invite", body, rh).Status)
	assert.Equal(t, 401, e.do("POST", "/admin/users/invite", body, "").Status)

	e.grant("rh", "admin::settings.users.create", "", "[]")
	assert.Equal(t, 201, e.do("POST", "/admin/users/invite", body, rh).Status, "a mudança vale logo após a invalidação")
}

func TestMeComPermissoesEfetivas(t *testing.T) {
	e := newEnv(t)
	e.registerFirst()
	boss, _ := e.loginOK("ana@exemplo.com", senha)

	me := e.do("GET", "/admin/auth/me", nil, boss)
	assert.Equal(t, true, me.data()["isSuperAdmin"])
	assert.NotEmpty(t, me.data()["permissions"])

	_, err := db.New(e.pool).UpsertAdminRole(context.Background(), db.UpsertAdminRoleParams{Code: "leitor", Name: "Leitor"})
	require.NoError(t, err)
	e.grant("leitor", "admin::settings.users.read", `["email"]`, `["admin::is-creator"]`)
	e.grant("leitor", "admin::acao.fantasma.read", "", "[]") // ignorada: não registrada
	leitor := inviteAndAccept(t, e, boss, "leitor@exemplo.com", "leitor")

	me = e.do("GET", "/admin/auth/me", nil, leitor)
	assert.Equal(t, false, me.data()["isSuperAdmin"])
	perms := me.data()["permissions"].([]any)
	require.Len(t, perms, 1)
	p := perms[0].(map[string]any)
	assert.Equal(t, "admin::settings.users.read", p["action"])
	assert.Equal(t, []any{"email"}, p["fields"])
	assert.Equal(t, []any{"admin::is-creator"}, p["conditions"])
}

func TestStartRemovePermissoesDeAcoesInexistentes(t *testing.T) {
	e := newEnv(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q := db.New(e.pool)
	role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: "x", Name: "X"})
	require.NoError(t, err)
	e.grant("x", "admin::settings.users.read", "", "[]")
	e.grant("x", "admin::acao.antiga.read", "", "[]")

	a, err := app.New(e.cfg, slog.New(slog.NewJSONHandler(io.Discard, nil)), e.pool, e.mail)
	require.NoError(t, err)
	require.NoError(t, a.Start(ctx))

	perms, err := q.ListAdminPermissionsByRole(ctx, role.ID)
	require.NoError(t, err)
	require.Len(t, perms, 1)
	assert.Equal(t, "admin::settings.users.read", perms[0].Action)
}
