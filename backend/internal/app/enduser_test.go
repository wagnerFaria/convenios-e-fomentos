package app_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/config"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/policy"
)

const userPass = "senha-de-usuario-1"

func (e *env) settingsPatch(boss string, patch map[string]any) resp {
	e.t.Helper()
	cur := e.do("GET", "/api/admin/end-user-settings", nil, boss)
	require.Equal(e.t, 200, cur.Status, cur.Body)
	body := map[string]any{}
	for k, v := range cur.data() {
		body[k] = v
	}
	for k, v := range patch {
		body[k] = v
	}
	return e.do("PUT", "/api/admin/end-user-settings", body, boss)
}

func (e *env) register(username, email string) resp {
	return e.do("POST", "/api/auth/local/register", map[string]any{"username": username, "email": email, "password": userPass}, "")
}

func (e *env) userLogin(identifier, password string) resp {
	return e.do("POST", "/api/auth/local", map[string]any{"identifier": identifier, "password": password}, "")
}

// userToken registra o usuário e devolve (accessToken, refreshToken, id).
func (e *env) userToken(username, email string) (string, string, string) {
	e.t.Helper()
	r := e.register(username, email)
	require.Equal(e.t, 201, r.Status, r.Body)
	tk := r.data()["tokens"].(map[string]any)
	return tk["accessToken"].(string), tk["refreshToken"].(string), r.data()["user"].(map[string]any)["id"].(string)
}

func (e *env) endRoleID(typ string) string {
	e.t.Helper()
	var id uuid.UUID
	require.NoError(e.t, e.pool.QueryRow(context.Background(), "SELECT id FROM end_user_roles WHERE type=$1", typ).Scan(&id))
	return id.String()
}

// registerAPIActions é mantido para os testes antigos: as ações api::document.* agora
// são registradas pelo próprio recurso de exemplo (pacote document).
func (e *env) registerAPIActions() {}

func TestPadroesDeUsuariosFinais(t *testing.T) {
	e, boss := setup(t)
	var types []string
	for _, r := range e.do("GET", "/api/admin/end-user-roles", nil, boss).list() {
		types = append(types, r.(map[string]any)["type"].(string))
	}
	assert.Equal(t, []string{"authenticated", "public"}, types)

	st := e.do("GET", "/api/admin/end-user-settings", nil, boss).data()
	assert.Equal(t, e.endRoleID("authenticated"), st["defaultRoleId"])
	assert.Equal(t, true, st["oneAccountPerEmail"])
	assert.Equal(t, true, st["registrationEnabled"])
	assert.Equal(t, false, st["emailConfirmationRequired"])

	tpls := e.do("GET", "/api/admin/email-templates", nil, boss).list()
	require.Len(t, tpls, 2)
	assert.Equal(t, "email-confirmation", tpls[0].(map[string]any)["key"])
	assert.Contains(t, tpls[0].(map[string]any)["bodyHtml"], "{{.URL}}")

	// Start é idempotente e não desfaz alterações.
	startCtx, stop := context.WithCancel(context.Background())
	t.Cleanup(stop)
	require.NoError(t, e.app.Start(startCtx))
	e.settingsPatch(boss, map[string]any{"registrationEnabled": false})
	require.NoError(t, e.app.EndUsers.EnsureDefaults(context.Background()))
	assert.Equal(t, false, e.do("GET", "/api/admin/end-user-settings", nil, boss).data()["registrationEnabled"])
}

func TestCadastroELogin(t *testing.T) {
	e, boss := setup(t)
	r := e.register("maria", "Maria@Exemplo.com")
	require.Equal(t, 201, r.Status, r.Body)
	user := r.data()["user"].(map[string]any)
	assert.Equal(t, true, user["confirmed"])
	assert.Equal(t, "authenticated", user["role"].(map[string]any)["type"])
	assert.NotContains(t, user, "passwordHash")
	assert.NotContains(t, r.Body["data"], "password")
	access := r.data()["tokens"].(map[string]any)["accessToken"].(string)

	me := e.do("GET", "/api/users/me", nil, access)
	require.Equal(t, 200, me.Status)
	assert.Equal(t, "maria", me.data()["username"])
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, "").Status)

	// Login por e-mail (sem diferenciar maiúsculas) e por nome de usuário.
	assert.Equal(t, 200, e.userLogin("maria@exemplo.com", userPass).Status)
	assert.Equal(t, 200, e.userLogin("MARIA", userPass).Status)
	wrong, ghost := e.userLogin("maria", "errada-errada-1"), e.userLogin("ninguem", "errada-errada-1")
	assert.Equal(t, 401, wrong.Status)
	assert.Equal(t, wrong.Body, ghost.Body, "não revela se a conta existe")

	// Duplicidade.
	assert.Equal(t, "already_taken", e.register("outra", "maria@exemplo.com").errCode(), "e-mail repetido")
	assert.Equal(t, "already_taken", e.register("MARIA", "outra@exemplo.com").errCode(), "usuário repetido")

	// Validação.
	for name, body := range map[string]map[string]any{
		"e-mail inválido":     {"username": "joao", "email": "x", "password": userPass},
		"usuário curto":       {"username": "jo", "email": "j@exemplo.com", "password": userPass},
		"usuário com espaço":  {"username": "jo ao", "email": "j@exemplo.com", "password": userPass},
		"senha curta":         {"username": "joao", "email": "j@exemplo.com", "password": "curta"},
		"sem campos":          {},
		"campo não permitido": {"username": "joao", "email": "j@exemplo.com", "password": userPass, "roleId": e.endRoleID("public")},
		"confirmed forçado":   {"username": "joao", "email": "j@exemplo.com", "password": userPass, "confirmed": true},
	} {
		assert.Equal(t, 422, e.do("POST", "/api/auth/local/register", body, "").Status, name)
	}
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM end_users").Scan(&n))
	assert.Equal(t, 1, n)

	// Cadastro desabilitado.
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"registrationEnabled": false}).Status)
	rd := e.register("joao", "j@exemplo.com")
	assert.Equal(t, 403, rd.Status)
	assert.Equal(t, "registration_disabled", rd.errCode())
}

func TestUmaContaPorEmailEmailsCompartilhados(t *testing.T) {
	e, boss := setup(t)
	e.userToken("ana", "casa@exemplo.com")
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"oneAccountPerEmail": false}).Status)
	senha2 := "outra-senha-forte-2"
	r := e.do("POST", "/api/auth/local/register", map[string]any{"username": "beto", "email": "casa@exemplo.com", "password": senha2}, "")
	require.Equal(t, 201, r.Status, r.Body)
	// Login por e-mail escolhe a conta cuja senha confere.
	assert.Equal(t, "ana", e.userLogin("casa@exemplo.com", userPass).data()["user"].(map[string]any)["username"])
	assert.Equal(t, "beto", e.userLogin("casa@exemplo.com", senha2).data()["user"].(map[string]any)["username"])
	assert.Equal(t, 401, e.userLogin("casa@exemplo.com", "nenhuma-senha-1").Status)
	// Voltar a exigir não apaga o que existe, mas bloqueia novos.
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"oneAccountPerEmail": true}).Status)
	assert.Equal(t, 409, e.register("carla", "casa@exemplo.com").Status)
}

func TestCamposExtrasNoCadastro(t *testing.T) {
	e, boss := setup(t)
	body := map[string]any{"username": "joao", "email": "j@exemplo.com", "password": userPass, "telefone": "11999990000", "idade": 30}
	r := e.do("POST", "/api/auth/local/register", body, "")
	require.Equal(t, 422, r.Status)
	fields := r.Body["error"].(map[string]any)["details"].(map[string]any)["fields"].([]any)
	assert.Len(t, fields, 2, "lista os campos não permitidos")

	// Nomes reservados e inválidos não entram na lista de permitidos.
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"registerAllowedFields": []string{"role"}}).Status)
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"registerAllowedFields": []string{"a b"}}).Status)
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"registerAllowedFields": []string{"telefone", "idade", "idade"}}).Status)
	assert.Equal(t, []any{"idade", "telefone"}, e.do("GET", "/api/admin/end-user-settings", nil, boss).data()["registerAllowedFields"])

	r = e.do("POST", "/api/auth/local/register", body, "")
	require.Equal(t, 201, r.Status, r.Body)
	profile := r.data()["user"].(map[string]any)["profile"].(map[string]any)
	assert.Equal(t, "11999990000", profile["telefone"])
	assert.EqualValues(t, 30, profile["idade"])
	body2 := map[string]any{"username": "ana", "email": "a@exemplo.com", "password": userPass, "telefone": map[string]any{"x": 1}}
	assert.Equal(t, 422, e.do("POST", "/api/auth/local/register", body2, "").Status, "só valores simples")
	body2["telefone"] = strings.Repeat("x", 501)
	assert.Equal(t, 422, e.do("POST", "/api/auth/local/register", body2, "").Status)
}

func TestConfirmacaoDeEmail(t *testing.T) {
	e, boss := setup(t)
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"emailConfirmationRequired": true}).Status)

	r := e.register("maria", "maria@exemplo.com")
	require.Equal(t, 201, r.Status, r.Body)
	assert.Nil(t, r.data()["tokens"], "sem tokens até confirmar")
	assert.Equal(t, false, r.data()["user"].(map[string]any)["confirmed"])
	require.Equal(t, 1, e.mail.Count())
	mail := e.mail.Last()
	assert.Equal(t, "maria@exemplo.com", mail.To)
	assert.Equal(t, "Confirme seu e-mail", mail.Subject)
	assert.Contains(t, mail.HTML, "Olá, maria!")
	token := e.mail.LastParam("confirmation")
	require.NotEmpty(t, token)
	assert.Contains(t, mail.HTML, "http://app.test/api/auth/email-confirmation?confirmation=")

	// Senha errada não revela o estado; senha certa informa que falta confirmar.
	assert.Equal(t, 401, e.userLogin("maria", "errada-errada-1").Status)
	nl := e.userLogin("maria", userPass)
	assert.Equal(t, 403, nl.Status)
	assert.Equal(t, "email_not_confirmed", nl.errCode())

	assert.Equal(t, 400, e.do("GET", "/api/auth/email-confirmation?confirmation=errado", nil, "").Status)
	assert.Equal(t, 400, e.do("GET", "/api/auth/email-confirmation", nil, "").Status)

	// Reenvio invalida o token anterior; e-mail desconhecido não envia nada.
	assert.Equal(t, 204, e.do("POST", "/api/auth/send-email-confirmation", map[string]any{"email": "maria@exemplo.com"}, "").Status)
	require.Equal(t, 2, e.mail.Count())
	newToken := e.mail.LastParam("confirmation")
	assert.NotEqual(t, token, newToken)
	assert.Equal(t, 204, e.do("POST", "/api/auth/send-email-confirmation", map[string]any{"email": "ghost@exemplo.com"}, "").Status)
	assert.Equal(t, 2, e.mail.Count())
	assert.Equal(t, 400, e.do("GET", "/api/auth/email-confirmation?confirmation="+url.QueryEscape(token), nil, "").Status, "token antigo invalidado")

	// Redirecionamento configurado.
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"emailConfirmationRedirectUrl": "https://front.exemplo.com/confirmado"}).Status)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/auth/email-confirmation?confirmation="+url.QueryEscape(newToken), nil)
	req.RemoteAddr = "10.0.0.1:1"
	rec := newRecorder()
	e.h.ServeHTTP(rec, req)
	assert.Equal(t, 302, rec.Code)
	assert.Equal(t, "https://front.exemplo.com/confirmado", rec.Header().Get("Location"))

	assert.Equal(t, 200, e.userLogin("maria", userPass).Status)
	assert.Equal(t, 204, e.do("POST", "/api/auth/send-email-confirmation", map[string]any{"email": "maria@exemplo.com"}, "").Status)
	assert.Equal(t, 2, e.mail.Count(), "já confirmado: nada a enviar")

	// Sem URL configurada, responde JSON.
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"emailConfirmationRedirectUrl": ""}).Status)
	e.register("joao", "joao@exemplo.com")
	assert.Equal(t, 200, e.do("GET", "/api/auth/email-confirmation?confirmation="+url.QueryEscape(e.mail.LastParam("confirmation")), nil, "").Status)
}

func TestRefreshLogoutEBloqueio(t *testing.T) {
	e, boss := setup(t)
	access, refresh, id := e.userToken("maria", "maria@exemplo.com")

	r := e.do("POST", "/api/auth/refresh", map[string]any{"refreshToken": refresh}, "")
	require.Equal(t, 200, r.Status, r.Body)
	refresh2 := r.data()["refreshToken"].(string)
	// Reuso do token antigo derruba a família inteira.
	assert.Equal(t, 401, e.do("POST", "/api/auth/refresh", map[string]any{"refreshToken": refresh}, "").Status)
	assert.Equal(t, 401, e.do("POST", "/api/auth/refresh", map[string]any{"refreshToken": refresh2}, "").Status)
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, access).Status)

	// Logout.
	a2 := e.userLogin("maria", userPass).data()["tokens"].(map[string]any)["accessToken"].(string)
	assert.Equal(t, 204, e.do("POST", "/api/auth/logout", nil, a2).Status)
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, a2).Status)
	assert.Equal(t, 401, e.do("POST", "/api/auth/logout", nil, "").Status)

	// Bloqueio: efeito imediato sobre access e refresh, e no login.
	tk := e.userLogin("maria", userPass).data()["tokens"].(map[string]any)
	a3, r3 := tk["accessToken"].(string), tk["refreshToken"].(string)
	assert.Equal(t, 200, e.do("GET", "/api/users/me", nil, a3).Status)
	blocked := e.do("POST", "/api/admin/end-users/"+id+"/block", nil, boss)
	require.Equal(t, 200, blocked.Status, blocked.Body)
	assert.Equal(t, true, blocked.data()["blocked"])
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, a3).Status)
	assert.Equal(t, 401, e.do("POST", "/api/auth/refresh", map[string]any{"refreshToken": r3}, "").Status)
	bl := e.userLogin("maria", userPass)
	assert.Equal(t, 403, bl.Status)
	assert.Equal(t, "account_blocked", bl.errCode())
	assert.Equal(t, 401, e.userLogin("maria", "errada-errada-1").Status, "senha errada continua genérica")

	assert.Equal(t, 200, e.do("POST", "/api/admin/end-users/"+id+"/unblock", nil, boss).Status)
	assert.Equal(t, 200, e.userLogin("maria", userPass).Status)
}

func TestRedefinicaoETrocaDeSenhaDeUsuario(t *testing.T) {
	e, boss := setup(t)
	access, _, _ := e.userToken("maria", "maria@exemplo.com")

	assert.Equal(t, 204, e.do("POST", "/api/auth/forgot-password", map[string]any{"email": "ghost@exemplo.com"}, "").Status)
	assert.Zero(t, e.mail.Count())

	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"resetPasswordUrl": "https://front.exemplo.com/nova-senha"}).Status)
	assert.Equal(t, 204, e.do("POST", "/api/auth/forgot-password", map[string]any{"email": "maria@exemplo.com"}, "").Status)
	require.Equal(t, 1, e.mail.Count())
	assert.Equal(t, "Redefinição de senha", e.mail.Last().Subject)
	assert.Contains(t, e.mail.Last().HTML, "https://front.exemplo.com/nova-senha?token=")
	token := e.mail.LastParam("token")

	assert.Equal(t, 422, e.do("POST", "/api/auth/reset-password", map[string]any{"token": token, "password": "curta"}, "").Status)
	assert.Equal(t, 400, e.do("POST", "/api/auth/reset-password", map[string]any{"token": "x", "password": "nova-senha-123"}, "").Status)
	assert.Equal(t, 204, e.do("POST", "/api/auth/reset-password", map[string]any{"token": token, "password": "nova-senha-123"}, "").Status)
	assert.Equal(t, 400, e.do("POST", "/api/auth/reset-password", map[string]any{"token": token, "password": "outra-senha-123"}, "").Status, "uso único")
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, access).Status, "sessões antigas revogadas")
	assert.Equal(t, 401, e.userLogin("maria", userPass).Status)
	tk := e.userLogin("maria", "nova-senha-123")
	require.Equal(t, 200, tk.Status)

	// Token de redefinição que não é de usuário final (aqui, um token de outro tipo) é recusado.
	assert.Equal(t, 400, e.do("POST", "/api/auth/reset-password", map[string]any{"token": "token-de-outro-tipo", "password": "nova-senha-456"}, "").Status)

	// Trocar a senha logado.
	a1 := tk.data()["tokens"].(map[string]any)["accessToken"].(string)
	a2 := e.userLogin("maria", "nova-senha-123").data()["tokens"].(map[string]any)["accessToken"].(string)
	assert.Equal(t, 422, e.do("POST", "/api/auth/change-password", map[string]any{"currentPassword": "errada-errada-1", "newPassword": "terceira-senha-1"}, a1).Status)
	assert.Equal(t, 422, e.do("POST", "/api/auth/change-password", map[string]any{"currentPassword": "nova-senha-123", "newPassword": "curta"}, a1).Status)
	assert.Equal(t, 401, e.do("POST", "/api/auth/change-password", map[string]any{"currentPassword": "x", "newPassword": "y"}, "").Status)
	assert.Equal(t, 204, e.do("POST", "/api/auth/change-password", map[string]any{"currentPassword": "nova-senha-123", "newPassword": "terceira-senha-1"}, a1).Status)
	assert.Equal(t, 200, e.do("GET", "/api/users/me", nil, a1).Status, "a sessão atual segue")
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, a2).Status, "as demais são revogadas")
	assert.Equal(t, 200, e.userLogin("maria", "terceira-senha-1").Status)
}

func TestTokensDeAdminEUsuarioFinalNaoSeMisturam(t *testing.T) {
	e, boss := setup(t)
	userAccess, _, _ := e.userToken("maria", "maria@exemplo.com")
	assert.Equal(t, 401, e.do("GET", "/api/users/me", nil, boss).Status, "token do MT Login na API pública")
	assert.Equal(t, 401, e.do("GET", "/api/admin/auth/me", nil, userAccess).Status, "JWT de usuário final na API admin")
	assert.Equal(t, 401, e.do("GET", "/api/admin/roles", nil, userAccess).Status)
	// Mesmo forjado com o segredo da API: o painel só aceita RS256 do MT Login.
	id := uuid.New()
	forged, _ := auth.NewJWTIssuer(e.cfg.JWTAPISecret, "admin", 60e9).Issue(id, uuid.New())
	assert.Equal(t, 401, e.do("GET", "/api/admin/auth/me", nil, forged).Status)
}

func newRecorder() *recorder { return &recorder{ResponseRecorder: newHTTPRecorder()} }

func TestRolesDeUsuarioFinal(t *testing.T) {
	e, boss := setup(t)
	e.registerAPIActions()

	r := e.do("POST", "/api/admin/end-user-roles", map[string]any{"name": "Moderadores", "description": "Moderam", "permissions": []string{"api::document.update", "api::document.find"}}, boss)
	require.Equal(t, 201, r.Status, r.Body)
	id := r.data()["id"].(string)
	assert.Equal(t, "custom-moderadores", r.data()["type"])
	assert.Equal(t, false, r.data()["isSystem"])
	perms := r.data()["permissions"].([]any)
	require.Len(t, perms, 2)
	first := perms[0].(map[string]any)
	assert.Equal(t, "api::document.find", first["action"])
	assert.Equal(t, "GET", first["method"], "devolve a rota vinculada")
	assert.Equal(t, "/api/documents", first["route"])
	assert.Equal(t, "custom-moderadores-2", e.do("POST", "/api/admin/end-user-roles", map[string]any{"name": "Moderadores"}, boss).data()["type"])

	bad := func(perm ...string) resp {
		return e.do("POST", "/api/admin/end-user-roles", map[string]any{"name": "X", "permissions": perm}, boss)
	}
	assert.Equal(t, 422, bad("api::nao.existe").Status)
	assert.Equal(t, 422, bad("admin::settings.users.read").Status, "ação de admin não é permitida")
	assert.Equal(t, 422, e.do("POST", "/api/admin/end-user-roles", map[string]any{"name": ""}, boss).Status)

	// Editar: sem "permissions" não altera; com a lista, substitui.
	u := e.do("PUT", "/api/admin/end-user-roles/"+id, map[string]any{"name": "Moderação", "description": "x"}, boss)
	require.Equal(t, 200, u.Status)
	assert.Len(t, u.data()["permissions"], 2)
	assert.Equal(t, "custom-moderadores", u.data()["type"], "o tipo é imutável")
	u = e.do("PUT", "/api/admin/end-user-roles/"+id, map[string]any{"name": "Moderação", "permissions": []string{"api::document.findOne"}}, boss)
	assert.Len(t, u.data()["permissions"], 1)
	assert.Equal(t, 422, e.do("PUT", "/api/admin/end-user-roles/"+id, map[string]any{"name": "M", "permissions": []string{"nada"}}, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/api/admin/end-user-roles/"+uuid.NewString(), map[string]any{"name": "M"}, boss).Status)

	// Excluir uma role move os usuários para a Public (na mesma transação).
	_, _, uid := e.userToken("maria", "maria@exemplo.com")
	require.Equal(t, 200, e.do("PUT", "/api/admin/end-users/"+uid+"/role", map[string]any{"roleId": id}, boss).Status)
	assert.EqualValues(t, 1, e.do("GET", "/api/admin/end-user-roles/"+id, nil, boss).data()["userCount"])
	require.Equal(t, 204, e.do("DELETE", "/api/admin/end-user-roles/"+id, nil, boss).Status)
	moved := e.do("GET", "/api/admin/end-users/"+uid, nil, boss)
	assert.Equal(t, "public", moved.data()["role"].(map[string]any)["type"])
	assert.Equal(t, 404, e.do("GET", "/api/admin/end-user-roles/"+id, nil, boss).Status)
	logs := e.do("GET", "/api/admin/audit-logs?action=end-user-role.delete", nil, boss)
	require.NotEmpty(t, logs.list())
	assert.EqualValues(t, 1, logs.list()[0].(map[string]any)["payload"].(map[string]any)["after"].(map[string]any)["usersMovedToPublic"])

	// Public e Authenticated não são excluídas; a role padrão também não.
	assert.Equal(t, "role_protected", e.do("DELETE", "/api/admin/end-user-roles/"+e.endRoleID("public"), nil, boss).errCode())
	assert.Equal(t, "role_protected", e.do("DELETE", "/api/admin/end-user-roles/"+e.endRoleID("authenticated"), nil, boss).errCode())
	mod := e.do("POST", "/api/admin/end-user-roles", map[string]any{"name": "Padrão novo"}, boss).data()["id"].(string)
	require.Equal(t, 200, e.settingsPatch(boss, map[string]any{"defaultRoleId": mod}).Status)
	assert.Equal(t, "role_is_default", e.do("DELETE", "/api/admin/end-user-roles/"+mod, nil, boss).errCode())
	assert.Equal(t, "custom-padrao-novo", e.do("GET", "/api/admin/end-user-roles/"+mod, nil, boss).data()["type"])
	// Novos cadastros recebem a role padrão configurada.
	r2 := e.register("joao", "j@exemplo.com")
	assert.Equal(t, mod, r2.data()["user"].(map[string]any)["role"].(map[string]any)["id"])
	// A Public não pode ser a padrão; role inexistente também não.
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"defaultRoleId": e.endRoleID("public")}).Status)
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"defaultRoleId": uuid.NewString()}).Status)
}

func TestAutorizacaoNaAPIPublicaEPolicyIsOwner(t *testing.T) {
	e, boss := setup(t)
	e.registerAPIActions()
	owners := map[string]uuid.UUID{}
	ok := func(w http.ResponseWriter, r *http.Request) { httpapi.JSON(w, 200, map[string]string{"ok": "1"}, nil) }
	lookup := policy.IsOwner(func(_ context.Context, r *http.Request) (uuid.UUID, error) {
		owner, found := owners[chi.URLParam(r, "id")]
		if !found {
			return uuid.Nil, httpapi.ErrNotFound()
		}
		return owner, nil
	})
	id := e.app.Identifier
	e.app.Router.Group(func(r chi.Router) {
		r.Use(id.Identify)
		r.With(id.RequireAction("api::document.find")).Get("/api/t/documents", ok)
		r.With(id.RequireAction("api::document.create")).Post("/api/t/documents", ok)
		r.With(id.RequireAction("api::document.update"), policy.Require(lookup)).Put("/api/t/documents/{id}", ok)
	})

	// Sem permissões: anônimo → 401, autenticado → 403.
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, "").Status)
	maria, _, mariaID := e.userToken("maria", "maria@exemplo.com")
	assert.Equal(t, 403, e.do("GET", "/api/t/documents", nil, maria).Status)

	// Public libera a leitura para anônimos; Authenticated cria e atualiza.
	put := func(role string, actions ...string) {
		rid := e.endRoleID(role)
		r := e.do("PUT", "/api/admin/end-user-roles/"+rid, map[string]any{"name": role, "permissions": actions}, boss)
		require.Equal(t, 200, r.Status, r.Body)
	}
	put("public", "api::document.find")
	put("authenticated", "api::document.find", "api::document.create", "api::document.update")

	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, "").Status, "mudança vale na hora, sem reiniciar")
	assert.Equal(t, 401, e.do("POST", "/api/t/documents", map[string]any{}, "").Status, "anônimo não cria")
	assert.Equal(t, 200, e.do("POST", "/api/t/documents", map[string]any{}, maria).Status)
	assert.Equal(t, 200, e.do("GET", "/api/t/documents", nil, maria).Status)

	// Token inválido nunca vira anônimo; tokens de API ainda não existem.
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, "lixo").Status)
	assert.Equal(t, 401, e.do("GET", "/api/t/documents", nil, "gat_abc").Status)
	req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/t/documents", nil)
	req.Header.Set("Authorization", "Basic abc")
	req.RemoteAddr = "10.0.0.1:1"
	rec := newRecorder()
	e.h.ServeHTTP(rec, req)
	assert.Equal(t, 401, rec.Code)

	// Policy is-owner (depois da permissão).
	docID := uuid.NewString()
	owners[docID] = uuid.MustParse(mariaID)
	assert.Equal(t, 200, e.do("PUT", "/api/t/documents/"+docID, map[string]any{}, maria).Status, "dono")
	joao, _, _ := e.userToken("joao", "j@exemplo.com")
	forbidden := e.do("PUT", "/api/t/documents/"+docID, map[string]any{}, joao)
	assert.Equal(t, 403, forbidden.Status, "outro usuário")
	assert.Equal(t, "is-owner", forbidden.Body["error"].(map[string]any)["details"].(map[string]any)["policy"])
	assert.Equal(t, 404, e.do("PUT", "/api/t/documents/"+uuid.NewString(), map[string]any{}, maria).Status)
	assert.Equal(t, 401, e.do("PUT", "/api/t/documents/"+docID, map[string]any{}, "").Status, "anônimo sem permissão da ação: 401 antes da policy")

	// Remover a permissão revoga na hora.
	put("authenticated", "api::document.find")
	assert.Equal(t, 403, e.do("PUT", "/api/t/documents/"+docID, map[string]any{}, maria).Status)
	// Usuário movido para a Public perde as permissões de Authenticated.
	put("authenticated", "api::document.find", "api::document.create")
	jid := e.do("GET", "/api/users/me", nil, joao).data()["id"].(string)
	require.Equal(t, 200, e.do("PUT", "/api/admin/end-users/"+jid+"/role", map[string]any{"roleId": e.endRoleID("public")}, boss).Status)
	assert.Equal(t, 403, e.do("POST", "/api/t/documents", map[string]any{}, joao).Status)
}

func TestSincronizacaoRemovePermissoesDeUsuarioFinalOrfas(t *testing.T) {
	e, _ := setup(t)
	ctx := context.Background()
	q := db.New(e.pool)
	rid := uuid.MustParse(e.endRoleID("authenticated"))
	require.NoError(t, q.InsertEndUserPermission(ctx, db.InsertEndUserPermissionParams{RoleID: rid, Action: "api::antiga.find"}))
	startCtx, stop := context.WithCancel(ctx)
	t.Cleanup(stop)
	require.NoError(t, e.app.Start(startCtx))
	actions, err := q.ListEndUserPermissionActions(ctx, rid)
	require.NoError(t, err)
	assert.Empty(t, actions)
}

func TestAdministracaoDeUsuariosFinais(t *testing.T) {
	e, boss := setup(t)
	for _, n := range []string{"ana", "bia", "caio"} {
		e.userToken(n, n+"@exemplo.com")
	}
	list := e.do("GET", "/api/admin/end-users", nil, boss)
	require.Equal(t, 200, list.Status)
	assert.Len(t, list.list(), 3)
	assert.EqualValues(t, 3, list.meta()["total"])
	count := func(q string) int { return len(e.do("GET", "/api/admin/end-users?"+q, nil, boss).list()) }
	assert.Equal(t, 1, count("q=bia"))
	assert.Equal(t, 3, count("confirmed=true"))
	assert.Equal(t, 0, count("blocked=true"))
	assert.Equal(t, 3, count("roleId="+e.endRoleID("authenticated")))
	assert.Equal(t, 2, count("pageSize=2"))
	assert.Equal(t, "caio", e.do("GET", "/api/admin/end-users?sort=username:desc", nil, boss).list()[0].(map[string]any)["username"])
	for _, bad := range []string{"sort=password_hash", "confirmed=talvez", "roleId=x", "page=0"} {
		assert.Equal(t, 422, e.do("GET", "/api/admin/end-users?"+bad, nil, boss).Status, bad)
	}

	// CRUD.
	c := e.do("POST", "/api/admin/end-users", map[string]any{"username": "dora", "email": "dora@exemplo.com", "password": userPass, "confirmed": true}, boss)
	require.Equal(t, 201, c.Status, c.Body)
	id := c.data()["id"].(string)
	assert.Equal(t, 200, e.userLogin("dora", userPass).Status)
	assert.Equal(t, 409, e.do("POST", "/api/admin/end-users", map[string]any{"username": "dora2", "email": "dora@exemplo.com", "password": userPass}, boss).Status)
	assert.Equal(t, 422, e.do("POST", "/api/admin/end-users", map[string]any{"username": "x1", "email": "x@exemplo.com", "password": userPass, "roleId": uuid.NewString()}, boss).Status)

	u := e.do("PUT", "/api/admin/end-users/"+id, map[string]any{"username": "dora_s", "email": "dora.s@exemplo.com", "password": "senha-nova-forte-1"}, boss)
	require.Equal(t, 200, u.Status, u.Body)
	assert.Equal(t, 401, e.userLogin("dora", userPass).Status)
	assert.Equal(t, 200, e.userLogin("dora_s", "senha-nova-forte-1").Status)
	assert.Equal(t, 409, e.do("PUT", "/api/admin/end-users/"+id, map[string]any{"username": "ana", "email": "dora.s@exemplo.com"}, boss).Status)
	assert.Equal(t, 409, e.do("PUT", "/api/admin/end-users/"+id, map[string]any{"username": "dora_s", "email": "ana@exemplo.com"}, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/api/admin/end-users/"+uuid.NewString(), map[string]any{"username": "abc", "email": "a@b.com"}, boss).Status)

	// Marcar como confirmado e alterar role.
	pend := e.do("POST", "/api/admin/end-users", map[string]any{"username": "edu", "email": "edu@exemplo.com", "password": userPass}, boss).data()["id"].(string)
	assert.Equal(t, true, e.do("POST", "/api/admin/end-users/"+pend+"/confirm", nil, boss).data()["confirmed"])
	assert.Equal(t, 422, e.do("PUT", "/api/admin/end-users/"+pend+"/role", map[string]any{"roleId": uuid.NewString()}, boss).Status)

	assert.Equal(t, 204, e.do("DELETE", "/api/admin/end-users/"+id, nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/api/admin/end-users/"+id, nil, boss).Status)
	assert.Equal(t, 404, e.do("DELETE", "/api/admin/end-users/"+id, nil, boss).Status)

	// Auditoria sem a senha.
	logs := e.do("GET", "/api/admin/audit-logs?entity=end_user&action=end-user.update", nil, boss)
	require.Len(t, logs.list(), 1)
	raw := logs.list()[0].(map[string]any)["payload"]
	assert.NotContains(t, strings.ToLower(toJSON(raw)), "password")
}

func TestConfiguracoesETemplatesDeEmail(t *testing.T) {
	e, boss := setup(t)
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"resetPasswordUrl": "javascript:alert(1)"}).Status)
	assert.Equal(t, 422, e.settingsPatch(boss, map[string]any{"emailConfirmationRedirectUrl": "/relativa"}).Status)
	assert.Equal(t, 200, e.settingsPatch(boss, map[string]any{"resetPasswordUrl": "http://localhost:3000/reset"}).Status)
	logs := e.do("GET", "/api/admin/audit-logs?action=end-user-settings.update", nil, boss)
	require.NotEmpty(t, logs.list())
	p := logs.list()[0].(map[string]any)["payload"].(map[string]any)
	assert.Equal(t, "", p["before"].(map[string]any)["resetPasswordUrl"])
	assert.Equal(t, "http://localhost:3000/reset", p["after"].(map[string]any)["resetPasswordUrl"])

	tpl := map[string]any{"senderName": "Equipe", "senderEmail": "equipe@exemplo.com", "replyTo": "ajuda@exemplo.com",
		"subject": "Olá {{.User.Username}}", "bodyHtml": `<p>Oi {{.User.Username}} ({{.User.Email}}) — <a href="{{.URL}}">link</a> {{.ServerURL}} {{.Token}}</p>`}
	r := e.do("PUT", "/api/admin/email-templates/reset-password", tpl, boss)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Equal(t, "Olá {{.User.Username}}", e.do("GET", "/api/admin/email-templates/reset-password", nil, boss).data()["subject"])

	invalid := map[string]map[string]any{
		"sintaxe":              {"bodyHtml": "<p>{{.User.Username</p>"},
		"variável inexistente": {"bodyHtml": "<p>{{.Desconhecida}}</p>"},
		"assunto inválido":     {"subject": "{{.Nada}}"},
		"remetente inválido":   {"senderEmail": "x"},
		"responder para":       {"replyTo": "x"},
	}
	for name, patch := range invalid {
		body := map[string]any{}
		for k, v := range tpl {
			body[k] = v
		}
		for k, v := range patch {
			body[k] = v
		}
		assert.Equal(t, 422, e.do("PUT", "/api/admin/email-templates/reset-password", body, boss).Status, name)
	}
	assert.Equal(t, 404, e.do("GET", "/api/admin/email-templates/nao-existe", nil, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/api/admin/email-templates/nao-existe", tpl, boss).Status)

	// O e-mail enviado usa o template editado.
	e.userToken("maria", "maria@exemplo.com")
	e.do("POST", "/api/auth/forgot-password", map[string]any{"email": "maria@exemplo.com"}, "")
	m := e.mail.Last()
	assert.Equal(t, "Olá maria", m.Subject)
	assert.Equal(t, "Equipe", m.FromName)
	assert.Equal(t, "equipe@exemplo.com", m.From)
	assert.Equal(t, "ajuda@exemplo.com", m.ReplyTo)
	assert.Contains(t, m.HTML, "maria@exemplo.com")
	assert.Contains(t, m.HTML, "http://localhost:3000/reset?token=")
	assert.Contains(t, m.HTML, "http://app.test")
}

func TestRotasAdminDeUsuariosFinaisExigemPermissao(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("vazio")
	nobody, _ := e.newAdmin(boss, "nobody@exemplo.com", "vazio")
	id, rid := uuid.NewString(), uuid.NewString()
	routes := []struct{ method, path string }{
		{"GET", "/api/admin/end-user-roles"}, {"POST", "/api/admin/end-user-roles"}, {"GET", "/api/admin/end-user-roles/" + rid},
		{"PUT", "/api/admin/end-user-roles/" + rid}, {"DELETE", "/api/admin/end-user-roles/" + rid},
		{"GET", "/api/admin/end-users"}, {"POST", "/api/admin/end-users"}, {"GET", "/api/admin/end-users/" + id}, {"PUT", "/api/admin/end-users/" + id},
		{"DELETE", "/api/admin/end-users/" + id}, {"POST", "/api/admin/end-users/" + id + "/block"}, {"POST", "/api/admin/end-users/" + id + "/unblock"},
		{"PUT", "/api/admin/end-users/" + id + "/role"}, {"POST", "/api/admin/end-users/" + id + "/confirm"},
		{"GET", "/api/admin/end-user-settings"}, {"PUT", "/api/admin/end-user-settings"},
		{"GET", "/api/admin/email-templates"}, {"GET", "/api/admin/email-templates/reset-password"}, {"PUT", "/api/admin/email-templates/reset-password"},
	}
	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			assert.Equal(t, 401, e.do(rt.method, rt.path, map[string]any{}, "").Status)
			assert.Equal(t, 403, e.do(rt.method, rt.path, map[string]any{}, nobody).Status)
		})
	}
	// Permissão parcial: só leitura de templates.
	e.grant("vazio", "admin::settings.email-templates.read", "", "[]")
	assert.Equal(t, 200, e.do("GET", "/api/admin/email-templates", nil, nobody).Status)
	assert.Equal(t, 403, e.do("PUT", "/api/admin/email-templates/reset-password", map[string]any{}, nobody).Status)
}

func TestLimiteDeRequisicoesNaAPIPublica(t *testing.T) {
	e := newEnv(t, func(c *config.Config) { c.RateLimitAuthPerMin = 3 })
	for range 3 {
		assert.Equal(t, 401, e.userLogin("maria", "errada-errada-1").Status)
	}
	r := e.userLogin("maria", "errada-errada-1")
	assert.Equal(t, 429, r.Status)
	assert.Equal(t, "rate_limited", r.errCode())
	assert.Equal(t, 401, e.userLogin("outra", "errada-errada-1").Status, "outro identificador, contagem própria")
	for range 3 {
		e.do("POST", "/api/auth/forgot-password", map[string]any{"email": "a@exemplo.com"}, "")
	}
	assert.Equal(t, 429, e.do("POST", "/api/auth/forgot-password", map[string]any{"email": "a@exemplo.com"}, "").Status)
}
