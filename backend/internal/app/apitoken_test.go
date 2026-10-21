package app_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/policy"
)

func (e *env) createToken(boss, name, typ, lifespan string, perms ...string) resp {
	body := map[string]any{"name": name, "type": typ, "lifespan": lifespan}
	if perms != nil {
		body["permissions"] = perms
	}
	return e.do("POST", "/admin/api-tokens", body, boss)
}

// tokenFixture registra ações e rotas públicas de teste protegidas por ação.
func (e *env) tokenFixture() {
	e.t.Helper()
	e.registerAPIActions()
	ok := func(w http.ResponseWriter, _ *http.Request) { httpapi.JSON(w, 200, map[string]string{"ok": "1"}, nil) }
	owner := policy.IsOwner(func(context.Context, *http.Request) (uuid.UUID, error) { return uuid.New(), nil })
	id := e.app.Identifier
	e.app.Router.Group(func(r chi.Router) {
		r.Use(id.Identify)
		r.With(id.RequireAction("api::document.find")).Get("/api/t/documents", ok)
		r.With(id.RequireAction("api::document.create")).Post("/api/t/documents", ok)
		r.With(id.RequireAction("api::document.update"), policy.Require(owner)).Put("/api/t/documents/{id}", ok)
	})
}

func (e *env) tokenRow(id string) (hash, hint string, expires, lastUsed *time.Time) {
	e.t.Helper()
	require.NoError(e.t, e.pool.QueryRow(context.Background(), "SELECT token_hash, hint, expires_at, last_used_at FROM api_tokens WHERE id=$1", id).Scan(&hash, &hint, &expires, &lastUsed))
	return
}

func TestCriarTokenMostraOValorUmaVezESoGuardaOHMAC(t *testing.T) {
	e, boss := setup(t)
	r := e.createToken(boss, "Integração ERP", "full-access", "30d")
	require.Equal(t, 201, r.Status, r.Body)
	raw := r.data()["token"].(string)
	id := r.data()["id"].(string)
	assert.True(t, strings.HasPrefix(raw, "gat_"))
	assert.Len(t, raw, 4+43)
	assert.Equal(t, "active", r.data()["status"])
	assert.Equal(t, "30d", r.data()["lifespan"])
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "/admin/api-tokens", strings.NewReader(`{"name":"cache","type":"read-only","lifespan":"7d"}`))
	req.Header.Set("Authorization", "Bearer "+boss)
	req.RemoteAddr = "10.0.0.1:1"
	rec := newRecorder()
	e.h.ServeHTTP(rec, req)
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"), "o valor do token não pode ser cacheado")
	assert.Equal(t, []any{}, r.data()["permissions"])
	assert.Equal(t, "gat_…"+raw[len(raw)-4:], r.data()["hint"])
	exp, err := time.Parse(time.RFC3339, r.data()["expiresAt"].(string))
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().Add(30*24*time.Hour), exp, time.Minute)

	// No banco só existe o HMAC-SHA256 (chave = API_TOKEN_SALT), nunca o token.
	hash, hint, _, _ := e.tokenRow(id)
	m := hmac.New(sha256.New, []byte(e.cfg.APITokenSalt))
	m.Write([]byte(raw))
	assert.Equal(t, hex.EncodeToString(m.Sum(nil)), hash)
	assert.NotContains(t, hash, raw)
	assert.Equal(t, raw[len(raw)-4:], hint)
	var leaked int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM api_tokens WHERE position($1 in name||description||hint||token_hash||permissions::text)>0", raw).Scan(&leaked))
	assert.Zero(t, leaked)

	// Depois da criação o valor nunca mais aparece.
	for _, resp := range []resp{e.do("GET", "/admin/api-tokens/"+id, nil, boss), e.do("GET", "/admin/api-tokens", nil, boss), e.do("GET", "/admin/api-tokens/"+id, nil, boss)} {
		body := toJSON(resp.Body)
		assert.NotContains(t, body, raw)
		assert.NotContains(t, body, hash)
		assert.NotContains(t, body, `"token"`)
	}

	// Cada token tem valor único.
	r2 := e.createToken(boss, "Outro", "read-only", "unlimited")
	assert.NotEqual(t, raw, r2.data()["token"])
	assert.Nil(t, r2.data()["expiresAt"])
	assert.Equal(t, "unlimited", r2.data()["lifespan"])
}

func TestDuracoesEValidacaoNaCriacao(t *testing.T) {
	e, boss := setup(t)
	e.registerAPIActions()
	for lifespan, days := range map[string]int{"7d": 7, "30d": 30, "90d": 90} {
		r := e.createToken(boss, "t-"+lifespan, "read-only", lifespan)
		require.Equal(t, 201, r.Status, r.Body)
		exp, _ := time.Parse(time.RFC3339, r.data()["expiresAt"].(string))
		assert.WithinDuration(t, time.Now().Add(time.Duration(days)*24*time.Hour), exp, time.Minute, lifespan)
	}
	cases := map[string]resp{
		"duração inválida":         e.createToken(boss, "a", "read-only", "15d"),
		"duração ausente":          e.do("POST", "/admin/api-tokens", map[string]any{"name": "b", "type": "read-only"}, boss),
		"tipo inválido":            e.createToken(boss, "c", "super", "7d"),
		"sem nome":                 e.createToken(boss, "", "read-only", "7d"),
		"custom sem ações":         e.createToken(boss, "d", "custom", "7d"),
		"custom com ação inválida": e.createToken(boss, "e", "custom", "7d", "api::nao.existe"),
		"custom com ação de admin": e.createToken(boss, "f", "custom", "7d", "admin::settings.users.read"),
		"read-only com ações":      e.createToken(boss, "g", "read-only", "7d", "api::document.find"),
		"full-access com ações":    e.createToken(boss, "h", "full-access", "7d", "api::document.find"),
	}
	for name, r := range cases {
		assert.Equal(t, 422, r.Status, name)
	}
	assert.Equal(t, 201, e.createToken(boss, "custom ok", "custom", "7d", "api::document.find", "api::document.find").Status)
	dup := e.createToken(boss, "CUSTOM OK", "read-only", "7d")
	assert.Equal(t, 409, dup.Status, "nome único (sem diferenciar maiúsculas)")
	assert.Equal(t, "name_taken", dup.errCode())
}

func TestTiposDeTokenNaAPIPublica(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	ro := e.createToken(boss, "leitura", "read-only", "7d").data()["token"].(string)
	full := e.createToken(boss, "total", "full-access", "7d").data()["token"].(string)
	custom := e.createToken(boss, "custom", "custom", "7d", "api::document.create").data()["token"].(string)

	// Sem permissão de role nenhuma: o token vale por si só.
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, "").Status, "anônimo continua sem acesso")

	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, ro).Status)
	assert.Equal(t, 403, e.do("POST", "/api/t/documents", map[string]any{}, ro).Status, "read-only tentando escrever")
	assert.Equal(t, 403, e.do("PUT", "/api/t/documents/"+uuid.NewString(), map[string]any{}, ro).Status)

	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, full).Status)
	assert.Equal(t, 200, e.do("POST", "/api/t/documents", map[string]any{}, full).Status)

	assert.Equal(t, 200, e.do("POST", "/api/t/documents", map[string]any{}, custom).Status)
	assert.Equal(t, 403, e.do("GET", "/api/t/documents", nil, custom).Status, "fora da lista")

	// Policy is-owner: token não tem usuário, logo nunca é dono.
	assert.Equal(t, 403, e.do("PUT", "/api/t/documents/"+uuid.NewString(), map[string]any{}, full).Status)

	// Token de API não acessa a API administrativa nem /api/users/me.
	assert.Equal(t, 401, e.do("GET", "/admin/roles", nil, full).Status)
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, full).Status)
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, full).Status)
}

func TestTokensInvalidosNaoViramAnonimos(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	good := e.createToken(boss, "ok", "full-access", "7d").data()["token"].(string)
	for name, tok := range map[string]string{
		"prefixo sozinho":     "gat_",
		"lixo":                "gat_abc",
		"tamanho certo":       "gat_" + strings.Repeat("a", 43),
		"um caractere a mais": good + "x",
		"um a menos":          good[:len(good)-1],
		"sem prefixo":         strings.TrimPrefix(good, "gat_"),
		"prefixo trocado":     "xxx_" + strings.TrimPrefix(good, "gat_"),
		"maiúsculas":          strings.ToUpper(good),
	} {
		assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, tok).Status, name)
	}
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, good).Status)
	// O token também é secreto para o HMAC: com outro salt o mesmo valor seria inválido.
	other := e.app.APITokens.Hash(good)
	m := hmac.New(sha256.New, []byte("outro-salt"))
	m.Write([]byte(good))
	assert.NotEqual(t, other, hex.EncodeToString(m.Sum(nil)))
}

func TestExpiracaoERevogacao(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	r := e.createToken(boss, "temporário", "full-access", "7d")
	tok, id := r.data()["token"].(string), r.data()["id"].(string)
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, tok).Status)

	_, err := e.pool.Exec(context.Background(), "UPDATE api_tokens SET expires_at = now() - interval '1 second' WHERE id=$1", id)
	require.NoError(t, err)
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, tok).Status, "expirado")
	assert.Equal(t, "expired", e.do("GET", "/admin/api-tokens/"+id, nil, boss).data()["status"])

	// Revogar vale na hora.
	r = e.createToken(boss, "revogável", "full-access", "unlimited")
	tok2, id2 := r.data()["token"].(string), r.data()["id"].(string)
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, tok2).Status)
	assert.Equal(t, 204, e.do("DELETE", "/admin/api-tokens/"+id2, nil, boss).Status)
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, tok2).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/api-tokens/"+id2, nil, boss).Status, "já revogado")
	assert.Equal(t, 404, e.do("DELETE", "/admin/api-tokens/"+uuid.NewString(), nil, boss).Status)

	// Some da lista, mas segue consultável e pode ser listado explicitamente.
	for _, it := range e.do("GET", "/admin/api-tokens", nil, boss).list() {
		assert.NotEqual(t, id2, it.(map[string]any)["id"])
	}
	found := false
	for _, it := range e.do("GET", "/admin/api-tokens?includeRevoked=true", nil, boss).list() {
		if it.(map[string]any)["id"] == id2 {
			found = true
			assert.Equal(t, "revoked", it.(map[string]any)["status"])
		}
	}
	assert.True(t, found)
	d := e.do("GET", "/admin/api-tokens/"+id2, nil, boss)
	assert.Equal(t, "revoked", d.data()["status"])
	assert.NotNil(t, d.data()["revokedAt"])

	// Revogado não é editado nem regenerado, e o nome fica livre de novo.
	body := map[string]any{"name": "x", "type": "read-only"}
	assert.Equal(t, "token_revoked", e.do("PUT", "/admin/api-tokens/"+id2, body, boss).errCode())
	assert.Equal(t, "token_revoked", e.do("POST", "/admin/api-tokens/"+id2+"/regenerate", nil, boss).errCode())
	assert.Equal(t, 201, e.createToken(boss, "revogável", "read-only", "7d").Status)
}

func TestRegenerarTokenInvalidaOAnterior(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	r := e.createToken(boss, "regen", "full-access", "30d")
	old, id := r.data()["token"].(string), r.data()["id"].(string)
	oldHash, _, _, _ := e.tokenRow(id)
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, old).Status)

	// Simula um token já perto de expirar: regenerar reinicia a validade (30 dias).
	_, err := e.pool.Exec(context.Background(), "UPDATE api_tokens SET expires_at = now() + interval '1 day' WHERE id=$1", id)
	require.NoError(t, err)

	n := e.do("POST", "/admin/api-tokens/"+id+"/regenerate", nil, boss)
	require.Equal(t, 200, n.Status, n.Body)
	fresh := n.data()["token"].(string)
	assert.NotEqual(t, old, fresh)
	assert.True(t, strings.HasPrefix(fresh, "gat_"))
	assert.Equal(t, id, n.data()["id"], "é o mesmo token (mesmas permissões e nome)")
	exp, _ := time.Parse(time.RFC3339, n.data()["expiresAt"].(string))
	assert.WithinDuration(t, time.Now().Add(30*24*time.Hour), exp, time.Minute)
	newHash, _, _, lastUsed := e.tokenRow(id)
	assert.NotEqual(t, oldHash, newHash)
	assert.Nil(t, lastUsed, "last_used_at reinicia")

	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, old).Status, "o valor anterior morre na hora")
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, fresh).Status)
	assert.Equal(t, 404, e.do("POST", "/admin/api-tokens/"+uuid.NewString()+"/regenerate", nil, boss).Status)
	assert.NotContains(t, toJSON(e.do("GET", "/admin/api-tokens/"+id, nil, boss).Body), fresh)
}

func TestEditarTokenValeNaProximaRequisicao(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	r := e.createToken(boss, "editável", "read-only", "7d")
	tok, id := r.data()["token"].(string), r.data()["id"].(string)
	other := e.createToken(boss, "outro", "read-only", "7d").data()["id"].(string)
	assert.Equal(t, 403, e.do("POST", "/api/t/documents", map[string]any{}, tok).Status)

	u := e.do("PUT", "/admin/api-tokens/"+id, map[string]any{"name": "editado", "description": "agora cria", "type": "custom", "permissions": []string{"api::document.create"}}, boss)
	require.Equal(t, 200, u.Status, u.Body)
	assert.Equal(t, "custom", u.data()["type"])
	assert.Equal(t, []any{"api::document.create"}, u.data()["permissions"])
	assert.NotContains(t, u.Body["data"], "token")
	assert.Equal(t, 200, e.do("POST", "/api/t/documents", map[string]any{}, tok).Status)
	assert.Equal(t, 403, e.do("GET", "/api/t/documents", nil, tok).Status)

	// Voltar para um tipo sem lista limpa as ações.
	u = e.do("PUT", "/admin/api-tokens/"+id, map[string]any{"name": "editado", "type": "read-only"}, boss)
	assert.Equal(t, []any{}, u.data()["permissions"])
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, tok).Status)

	assert.Equal(t, 409, e.do("PUT", "/admin/api-tokens/"+other, map[string]any{"name": "Editado", "type": "read-only"}, boss).Status)
	assert.Equal(t, 422, e.do("PUT", "/admin/api-tokens/"+id, map[string]any{"name": "x", "type": "custom"}, boss).Status)
	assert.Equal(t, 422, e.do("PUT", "/admin/api-tokens/"+id, map[string]any{"name": "x", "type": "custom", "permissions": []string{"api::nada"}}, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/admin/api-tokens/"+uuid.NewString(), map[string]any{"name": "x", "type": "read-only"}, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/api-tokens/xyz", nil, boss).Status)
}

func TestUltimoUsoEAtualizadoDeFormaLimitada(t *testing.T) {
	e, boss := setup(t)
	e.tokenFixture()
	r := e.createToken(boss, "uso", "full-access", "7d")
	tok, id := r.data()["token"].(string), r.data()["id"].(string)
	assert.Nil(t, r.data()["lastUsedAt"])

	for range 5 {
		require.Equal(t, 200, e.do("GET", "/api/t/documents", nil, tok).Status)
	}
	e.app.APITokens.Wait()
	_, _, _, first := e.tokenRow(id)
	require.NotNil(t, first)
	assert.WithinDuration(t, time.Now(), *first, 10*time.Second)

	// Dentro de um minuto não há nova escrita (nem na memória, nem no banco).
	time.Sleep(50 * time.Millisecond)
	for range 3 {
		e.do("GET", "/api/t/documents", nil, tok)
	}
	e.app.APITokens.Wait()
	_, _, _, again := e.tokenRow(id)
	assert.Equal(t, *first, *again)
	var writes int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM audit_logs WHERE entity='api_token' AND action LIKE '%use%'").Scan(&writes))
	assert.Zero(t, writes, "o uso não gera auditoria")

	// O banco também limita: com a marca recente, um novo touch é ignorado; com marca antiga, atualiza.
	_, err := e.pool.Exec(context.Background(), "UPDATE api_tokens SET last_used_at = now() - interval '2 minutes' WHERE id=$1", id)
	require.NoError(t, err)
	e.app.APITokens.ResetThrottleForTest()
	e.do("GET", "/api/t/documents", nil, tok)
	e.app.APITokens.Wait()
	_, _, _, later := e.tokenRow(id)
	assert.WithinDuration(t, time.Now(), *later, 10*time.Second, "marca antiga é renovada")

	// Segunda instância (memória zerada) dentro do minuto: o banco ignora a escrita.
	stamp := *later
	e.app.APITokens.ResetThrottleForTest()
	e.do("GET", "/api/t/documents", nil, tok)
	e.app.APITokens.Wait()
	_, _, _, same := e.tokenRow(id)
	assert.Equal(t, stamp, *same)

	// O valor aparece na API admin.
	d := e.do("GET", "/admin/api-tokens/"+id, nil, boss)
	assert.NotNil(t, d.data()["lastUsedAt"])
}

func TestAuditoriaDeTokensNaoVazaSegredos(t *testing.T) {
	e, boss := setup(t)
	e.registerAPIActions()
	r := e.createToken(boss, "auditado", "custom", "30d", "api::document.find")
	raw, id := r.data()["token"].(string), r.data()["id"].(string)
	hash, _, _, _ := e.tokenRow(id)
	e.do("PUT", "/admin/api-tokens/"+id, map[string]any{"name": "auditado 2", "type": "read-only"}, boss)
	regen := e.do("POST", "/admin/api-tokens/"+id+"/regenerate", nil, boss).data()["token"].(string)
	e.do("DELETE", "/admin/api-tokens/"+id, nil, boss)

	logs := e.do("GET", "/admin/audit-logs?entity=api_token&entityId="+id+"&pageSize=100", nil, boss)
	actions := map[string]bool{}
	for _, l := range logs.list() {
		m := l.(map[string]any)
		actions[m["action"].(string)] = true
		body := strings.ToLower(toJSON(m))
		for _, secret := range []string{raw, regen, hash, strings.ToLower(raw)} {
			assert.NotContains(t, body, secret)
		}
		assert.Equal(t, e.adminID("ana@exemplo.com"), m["actorId"])
	}
	for _, a := range []string{"api-token.create", "api-token.update", "api-token.regenerate", "api-token.revoke"} {
		assert.True(t, actions[a], a)
	}
	upd := e.do("GET", "/admin/audit-logs?action=api-token.update", nil, boss).list()[0].(map[string]any)["payload"].(map[string]any)
	assert.Equal(t, "custom", upd["before"].(map[string]any)["type"])
	assert.Equal(t, "read-only", upd["after"].(map[string]any)["type"])
	var createdBy uuid.UUID
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT created_by FROM api_tokens WHERE id=$1", id).Scan(&createdBy))
	assert.Equal(t, e.adminID("ana@exemplo.com"), createdBy.String())
}

func TestRotasDeTokensExigemPermissao(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("vazio")
	nobody, _ := e.newAdmin(boss, "nobody@exemplo.com", "vazio")
	id := uuid.NewString()
	routes := []struct{ method, path string }{
		{"GET", "/admin/api-tokens"}, {"POST", "/admin/api-tokens"}, {"GET", "/admin/api-tokens/" + id},
		{"PUT", "/admin/api-tokens/" + id}, {"POST", "/admin/api-tokens/" + id + "/regenerate"}, {"DELETE", "/admin/api-tokens/" + id},
	}
	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			assert.Equal(t, 401, e.do(rt.method, rt.path, map[string]any{}, "").Status)
			assert.Equal(t, 403, e.do(rt.method, rt.path, map[string]any{}, nobody).Status)
		})
	}
	// Permissões independentes: regenerar não vem com editar.
	e.grant("vazio", "admin::settings.api-tokens.read", "", "[]")
	e.grant("vazio", "admin::settings.api-tokens.update", "", "[]")
	tok := e.createToken(boss, "p", "read-only", "7d").data()["id"].(string)
	assert.Equal(t, 200, e.do("GET", "/admin/api-tokens/"+tok, nil, nobody).Status)
	assert.Equal(t, 403, e.do("POST", "/admin/api-tokens/"+tok+"/regenerate", nil, nobody).Status)
	assert.Equal(t, 403, e.do("DELETE", "/admin/api-tokens/"+tok, nil, nobody).Status)
	assert.Equal(t, 200, e.do("PUT", "/admin/api-tokens/"+tok, map[string]any{"name": "p2", "type": "read-only"}, nobody).Status)
}
