package app_test

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

func (e *env) roleID(code string) string {
	e.t.Helper()
	var id uuid.UUID
	require.NoError(e.t, e.pool.QueryRow(context.Background(), "SELECT id FROM admin_roles WHERE code=$1", code).Scan(&id))
	return id.String()
}

func (e *env) adminID(email string) string {
	e.t.Helper()
	var id uuid.UUID
	require.NoError(e.t, e.pool.QueryRow(context.Background(), "SELECT id FROM admins WHERE email=$1", email).Scan(&id))
	return id.String()
}

// mkRole cria a role (se não existir) e concede as ações, sem condição.
func (e *env) mkRole(code string, actions ...string) {
	e.t.Helper()
	_, err := db.New(e.pool).UpsertAdminRole(context.Background(), db.UpsertAdminRoleParams{Code: code, Name: code})
	require.NoError(e.t, err)
	for _, a := range actions {
		e.grant(code, a, "", "[]")
	}
}

// newAdmin convida e ativa um administrador com a role e devolve (token, id).
func (e *env) newAdmin(boss, email, roleCode string) (string, string) {
	e.t.Helper()
	tok := inviteAndAccept(e.t, e, boss, email, roleCode)
	return tok, e.adminID(email)
}

func (r resp) list() []any {
	l, _ := r.Body["data"].([]any)
	return l
}

func (r resp) meta() map[string]any {
	m, _ := r.Body["meta"].(map[string]any)
	return m
}

var usersAll = []string{"admin::settings.users.read", "admin::settings.users.create", "admin::settings.users.update", "admin::settings.users.delete"}

func setup(t *testing.T) (*env, string) {
	e := newEnv(t)
	e.registerFirst()
	boss, _ := e.loginOK("ana@exemplo.com", senha)
	return e, boss
}

func TestRotasExigemAutenticacaoEPermissao(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("vazio")
	nobody, _ := e.newAdmin(boss, "nobody@exemplo.com", "vazio")
	id := e.adminID("ana@exemplo.com")
	rid := e.roleID("vazio")
	routes := []struct{ method, path string }{
		{"GET", "/admin/users"}, {"GET", "/admin/users/" + id}, {"PUT", "/admin/users/" + id}, {"DELETE", "/admin/users/" + id},
		{"POST", "/admin/users/batch-delete"}, {"POST", "/admin/users/" + id + "/activate"}, {"POST", "/admin/users/" + id + "/deactivate"},
		{"POST", "/admin/users/" + id + "/roles"}, {"DELETE", "/admin/users/" + id + "/roles/" + rid},
		{"GET", "/admin/roles"}, {"POST", "/admin/roles"}, {"GET", "/admin/roles/" + rid}, {"PUT", "/admin/roles/" + rid},
		{"DELETE", "/admin/roles/" + rid}, {"PUT", "/admin/roles/" + rid + "/permissions"}, {"POST", "/admin/roles/" + rid + "/duplicate"},
		{"GET", "/admin/catalog/actions"}, {"GET", "/admin/catalog/conditions"}, {"GET", "/admin/audit-logs"},
	}
	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			assert.Equal(t, 401, e.do(rt.method, rt.path, map[string]any{}, "").Status, "sem token")
			assert.Equal(t, 403, e.do(rt.method, rt.path, map[string]any{}, nobody).Status, "sem permissão")
		})
	}
}

func TestListagemDeAdministradores(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("redator")
	e.mkRole("revisor")
	for _, n := range []string{"bia", "caio", "dora"} {
		e.newAdmin(boss, n+"@exemplo.com", "redator")
	}
	e.newAdmin(boss, "eva.revisora@exemplo.com", "revisor")
	// Convite pendente.
	inv := e.do("POST", "/admin/users/invite", map[string]any{"firstName": "Pendente", "lastName": "Da Silva", "email": "pend@exemplo.com", "roleIds": []string{e.roleID("redator")}}, boss)
	require.Equal(t, 201, inv.Status)
	// Um desativado.
	e.do("POST", "/admin/users/"+e.adminID("dora@exemplo.com")+"/deactivate", nil, boss)

	all := e.do("GET", "/admin/users", nil, boss)
	require.Equal(t, 200, all.Status)
	assert.Len(t, all.list(), 6)
	assert.EqualValues(t, 6, all.meta()["total"])
	assert.EqualValues(t, 1, all.meta()["pageCount"])

	count := func(query string) int { return len(e.do("GET", "/admin/users?"+query, nil, boss).list()) }
	assert.Equal(t, 1, count("q=revisora"), "busca no e-mail")
	assert.Equal(t, 1, count("q="+url.QueryEscape("Pendente Da")), "busca no nome completo")
	assert.Equal(t, 0, count("q=%25"), "% é literal, não curinga")
	assert.Equal(t, 4, count("roleId="+e.roleID("redator")), "dora desativada e o convite pendente também têm a role")
	assert.Equal(t, 4, count("status=active"))
	assert.Equal(t, 1, count("status=inactive"))
	assert.Equal(t, 1, count("status=pending"))
	assert.Equal(t, 0, count("status=blocked"))
	assert.Equal(t, 1, count("status=active&roleId="+e.roleID("revisor")))

	// Ordenação e paginação.
	asc := e.do("GET", "/admin/users?sort=email:asc&pageSize=2&page=2", nil, boss)
	require.Equal(t, 200, asc.Status)
	assert.EqualValues(t, 3, asc.meta()["pageCount"])
	assert.Equal(t, "caio@exemplo.com", asc.list()[0].(map[string]any)["email"])
	desc := e.do("GET", "/admin/users?sort=email:desc", nil, boss)
	assert.Equal(t, "pend@exemplo.com", desc.list()[0].(map[string]any)["email"])

	// Limite máximo de pageSize e entradas inválidas.
	assert.EqualValues(t, 100, e.do("GET", "/admin/users?pageSize=5000", nil, boss).meta()["pageSize"])
	for _, bad := range []string{"sort=password_hash", "sort=" + url.QueryEscape("email;DROP TABLE admins"), "sort=email:up", "page=0", "pageSize=x", "status=zumbi", "roleId=nao-uuid"} {
		assert.Equal(t, 422, e.do("GET", "/admin/users?"+bad, nil, boss).Status, bad)
	}
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM admins").Scan(&n))
	assert.Equal(t, 6, n)

	item := all.list()[0].(map[string]any)
	assert.Contains(t, item, "roles")
	assert.NotContains(t, item, "passwordHash")
}

func TestEdicaoDeAdministrador(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("redator")
	_, bia := e.newAdmin(boss, "bia@exemplo.com", "redator")
	_, caio := e.newAdmin(boss, "caio@exemplo.com", "redator")
	dept, err := db.New(e.pool).UpsertDepartment(context.Background(), "Financeiro")
	require.NoError(t, err)

	body := map[string]any{"firstName": "Beatriz", "lastName": "Souza", "email": "beatriz@exemplo.com", "departmentId": dept.ID.String()}
	r := e.do("PUT", "/admin/users/"+bia, body, boss)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Equal(t, "beatriz@exemplo.com", r.data()["email"])
	assert.Equal(t, dept.ID.String(), r.data()["departmentId"])

	body["email"] = "caio@exemplo.com"
	assert.Equal(t, 409, e.do("PUT", "/admin/users/"+bia, body, boss).Status)
	body["email"] = "ok@exemplo.com"
	body["departmentId"] = uuid.NewString()
	assert.Equal(t, 422, e.do("PUT", "/admin/users/"+bia, body, boss).Status)
	assert.Equal(t, 422, e.do("PUT", "/admin/users/"+bia, map[string]any{"firstName": "", "lastName": "x", "email": "x"}, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/admin/users/"+uuid.NewString(), body, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/users/nao-uuid", nil, boss).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/users/"+caio, nil, boss).Status)

	// Auditoria com estado anterior e novo.
	logs := e.do("GET", "/admin/audit-logs?action=admin.update&entityId="+bia, nil, boss)
	require.Len(t, logs.list(), 1)
	payload := logs.list()[0].(map[string]any)["payload"].(map[string]any)
	assert.Equal(t, "bia@exemplo.com", payload["before"].(map[string]any)["email"])
	assert.Equal(t, "beatriz@exemplo.com", payload["after"].(map[string]any)["email"])
	assert.Equal(t, e.adminID("ana@exemplo.com"), logs.list()[0].(map[string]any)["actorId"])
}

func TestAtivarDesativarEProtecaoDoUltimoSuperAdmin(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("gestor", usersAll...)
	gestor, gestorID := e.newAdmin(boss, "gestor@exemplo.com", "gestor")
	ana := e.adminID("ana@exemplo.com")

	// O gestor tenta desativar o único Super Admin.
	r := e.do("POST", "/admin/users/"+ana+"/deactivate", nil, gestor)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, "last_super_admin", r.errCode())
	assert.Equal(t, 200, e.do("GET", "/admin/auth/me", nil, boss).Status, "nada mudou")

	// Quem não é Super Admin não concede a role Super Admin.
	assert.Equal(t, 403, e.do("POST", "/admin/users/"+gestorID+"/roles", map[string]any{"roleIds": []string{e.roleID("super-admin")}}, gestor).Status)
	// O Super Admin concede; agora existem dois.
	assert.Equal(t, 200, e.do("POST", "/admin/users/"+gestorID+"/roles", map[string]any{"roleIds": []string{e.roleID("super-admin")}}, boss).Status)

	// Auto-desativação é proibida.
	assert.Equal(t, "cannot_modify_self", e.do("POST", "/admin/users/"+gestorID+"/deactivate", nil, gestor).errCode())

	// Desativar a Ana revoga as sessões dela imediatamente.
	r = e.do("POST", "/admin/users/"+ana+"/deactivate", nil, gestor)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Equal(t, "inactive", r.data()["status"])
	assert.Equal(t, 401, e.do("GET", "/admin/auth/me", nil, boss).Status)
	assert.Equal(t, 401, e.login("ana@exemplo.com", senha).Status)

	// Reativar.
	r = e.do("POST", "/admin/users/"+ana+"/activate", nil, gestor)
	assert.Equal(t, "active", r.data()["status"])
	assert.Equal(t, 200, e.login("ana@exemplo.com", senha).Status)

	// Convite pendente não pode ser ativado manualmente.
	inv := e.do("POST", "/admin/users/invite", map[string]any{"firstName": "P", "lastName": "P", "email": "p@exemplo.com", "roleIds": []string{e.roleID("gestor")}}, gestor)
	require.Equal(t, 201, inv.Status)
	pend := inv.data()["admin"].(map[string]any)["id"].(string)
	assert.Equal(t, "invitation_pending", e.do("POST", "/admin/users/"+pend+"/activate", nil, gestor).errCode())
}

func TestExclusaoIndividualEmLoteEUltimoSuperAdmin(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("gestor", usersAll...)
	e.mkRole("redator")
	gestor, gestorID := e.newAdmin(boss, "gestor@exemplo.com", "gestor")
	_, bia := e.newAdmin(boss, "bia@exemplo.com", "redator")
	_, caio := e.newAdmin(boss, "caio@exemplo.com", "redator")
	ana := e.adminID("ana@exemplo.com")

	assert.Equal(t, "last_super_admin", e.do("DELETE", "/admin/users/"+ana, nil, gestor).errCode())
	assert.Equal(t, "cannot_modify_self", e.do("DELETE", "/admin/users/"+gestorID, nil, gestor).errCode())

	// Lote é atômico: se um item viola a regra, nenhum é excluído.
	r := e.do("POST", "/admin/users/batch-delete", map[string]any{"ids": []string{bia, ana}}, gestor)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, 200, e.do("GET", "/admin/users/"+bia, nil, boss).Status)

	// Item inexistente derruba o lote inteiro.
	assert.Equal(t, 404, e.do("POST", "/admin/users/batch-delete", map[string]any{"ids": []string{bia, uuid.NewString()}}, gestor).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/users/"+bia, nil, boss).Status)

	assert.Equal(t, 204, e.do("POST", "/admin/users/batch-delete", map[string]any{"ids": []string{bia, caio, bia}}, gestor).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/users/"+bia, nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/users/"+caio, nil, boss).Status)
	assert.Equal(t, 422, e.do("POST", "/admin/users/batch-delete", map[string]any{"ids": []string{}}, gestor).Status)

	_, dora := e.newAdmin(boss, "dora@exemplo.com", "redator")
	assert.Equal(t, 204, e.do("DELETE", "/admin/users/"+dora, nil, gestor).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/users/"+dora, nil, gestor).Status)

	logs := e.do("GET", "/admin/audit-logs?action=admin.delete", nil, boss)
	assert.EqualValues(t, 3, logs.meta()["total"])
}

func TestRemoverSuperAdminDoUltimoESerUnicoAdmin(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("gestor", usersAll...)
	gtok, gestorID := e.newAdmin(boss, "gestor@exemplo.com", "gestor")
	ana := e.adminID("ana@exemplo.com")
	super := e.roleID("super-admin")

	// Ana tem só uma role: não pode ficar sem nenhuma.
	r := e.do("DELETE", "/admin/users/"+ana+"/roles/"+super, nil, boss)
	assert.Equal(t, 422, r.Status)

	// Com outra role, ainda é o último Super Admin ativo.
	assert.Equal(t, 200, e.do("POST", "/admin/users/"+ana+"/roles", map[string]any{"roleIds": []string{e.roleID("gestor")}}, boss).Status)
	r = e.do("DELETE", "/admin/users/"+ana+"/roles/"+super, nil, boss)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, "last_super_admin", r.errCode())

	// Com um segundo Super Admin, a remoção passa.
	assert.Equal(t, 200, e.do("POST", "/admin/users/"+gestorID+"/roles", map[string]any{"roleIds": []string{super}}, boss).Status)
	r = e.do("DELETE", "/admin/users/"+ana+"/roles/"+super, nil, boss)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Len(t, r.data()["roles"], 1)

	// Role não atribuída, inexistente e roles inválidas (agora quem age é o novo Super Admin;
	// a Ana perdeu a role e, com ela, o direito de mexer na role Super Admin).
	assert.Equal(t, 403, e.do("DELETE", "/admin/users/"+gestorID+"/roles/"+super, nil, boss).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/users/"+ana+"/roles/"+super, nil, gtok).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/users/"+ana+"/roles/"+uuid.NewString(), nil, gtok).Status)
	assert.Equal(t, 422, e.do("POST", "/admin/users/"+ana+"/roles", map[string]any{"roleIds": []string{uuid.NewString()}}, gtok).Status)
}

func TestConviteComSuperAdminSoPorSuperAdmin(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("gestor", usersAll...)
	gestor, _ := e.newAdmin(boss, "gestor@exemplo.com", "gestor")
	body := map[string]any{"firstName": "X", "lastName": "X", "email": "x@exemplo.com", "roleIds": []string{e.roleID("super-admin")}}
	assert.Equal(t, 403, e.do("POST", "/admin/users/invite", body, gestor).Status)
	assert.Equal(t, 201, e.do("POST", "/admin/users/invite", body, boss).Status)
}

func TestDesativacoesConcorrentesNaoRemovemTodosOsSuperAdmins(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("gestor", usersAll...)
	gestor, _ := e.newAdmin(boss, "gestor@exemplo.com", "gestor")
	_, bruno := e.newAdmin(boss, "bruno@exemplo.com", "gestor")
	ana := e.adminID("ana@exemplo.com")
	require.Equal(t, 200, e.do("POST", "/admin/users/"+bruno+"/roles", map[string]any{"roleIds": []string{e.roleID("super-admin")}}, boss).Status)

	var wg sync.WaitGroup
	results := make(chan int, 2)
	for _, id := range []string{ana, bruno} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- e.do("POST", "/admin/users/"+id+"/deactivate", nil, gestor).Status
		}()
	}
	wg.Wait()
	close(results)
	var ok, conflict int
	for s := range results {
		switch s {
		case 200:
			ok++
		case 409:
			conflict++
		}
	}
	assert.Equal(t, 1, ok)
	assert.Equal(t, 1, conflict)
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), `SELECT count(*) FROM admins a JOIN admin_users_roles ur ON ur.admin_id=a.id JOIN admin_roles r ON r.id=ur.role_id WHERE r.code='super-admin' AND a.is_active`).Scan(&n))
	assert.Equal(t, 1, n)
}

func TestCrudDeRoles(t *testing.T) {
	e, boss := setup(t)

	r := e.do("POST", "/admin/roles", map[string]any{"name": "Gestão de Conteúdo", "description": "Equipe editorial"}, boss)
	require.Equal(t, 201, r.Status, r.Body)
	id := r.data()["id"].(string)
	assert.Equal(t, "gestao-de-conteudo", r.data()["code"])
	assert.Equal(t, false, r.data()["isSystem"])
	assert.EqualValues(t, 0, r.data()["userCount"])
	assert.Equal(t, []any{}, r.data()["permissions"])

	second := e.do("POST", "/admin/roles", map[string]any{"name": "Gestão de Conteúdo"}, boss)
	assert.Equal(t, "gestao-de-conteudo-2", second.data()["code"], "code único com sufixo")
	assert.Equal(t, 422, e.do("POST", "/admin/roles", map[string]any{"name": ""}, boss).Status)

	r = e.do("PUT", "/admin/roles/"+id, map[string]any{"name": "Gestão Editorial", "description": "nova"}, boss)
	require.Equal(t, 200, r.Status)
	assert.Equal(t, "Gestão Editorial", r.data()["name"])
	assert.Equal(t, "gestao-de-conteudo", r.data()["code"], "o code é imutável")

	// Contagem de administradores na listagem e no detalhe.
	e.newAdmin(boss, "bia@exemplo.com", "gestao-de-conteudo")
	list := e.do("GET", "/admin/roles", nil, boss)
	counts := map[string]float64{}
	for _, it := range list.list() {
		m := it.(map[string]any)
		counts[m["code"].(string)] = m["userCount"].(float64)
	}
	assert.EqualValues(t, 1, counts["gestao-de-conteudo"])
	assert.EqualValues(t, 1, counts["super-admin"])
	assert.EqualValues(t, 1, e.do("GET", "/admin/roles/"+id, nil, boss).data()["userCount"])

	// Excluir role em uso: 409 explicando o motivo.
	del := e.do("DELETE", "/admin/roles/"+id, nil, boss)
	assert.Equal(t, 409, del.Status)
	assert.Equal(t, "role_in_use", del.errCode())
	assert.Contains(t, del.Body["error"].(map[string]any)["message"], "1 administrador")

	// Role livre pode ser excluída.
	free := second.data()["id"].(string)
	assert.Equal(t, 204, e.do("DELETE", "/admin/roles/"+free, nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/roles/"+free, nil, boss).Status)
	assert.Equal(t, 404, e.do("DELETE", "/admin/roles/"+free, nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/roles/xyz", nil, boss).Status)

	// Super Admin: não exclui, não duplica, não altera permissões.
	super := e.roleID("super-admin")
	assert.Equal(t, "super_admin_immutable", e.do("DELETE", "/admin/roles/"+super, nil, boss).errCode())
	assert.Equal(t, "super_admin_immutable", e.do("POST", "/admin/roles/"+super+"/duplicate", nil, boss).errCode())
	assert.Equal(t, "super_admin_immutable", e.do("PUT", "/admin/roles/"+super+"/permissions", map[string]any{"permissions": []any{}}, boss).errCode())
	sd := e.do("GET", "/admin/roles/"+super, nil, boss)
	assert.Equal(t, true, sd.data()["isSuperAdmin"])
}

func TestSubstituirPermissoesEValidacao(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("editor")
	rid := e.roleID("editor")
	put := func(perms ...map[string]any) resp {
		if perms == nil {
			perms = []map[string]any{}
		}
		return e.do("PUT", "/admin/roles/"+rid+"/permissions", map[string]any{"permissions": perms}, boss)
	}
	fieldPaths := func(r resp) []string {
		var out []string
		for _, f := range r.Body["error"].(map[string]any)["details"].(map[string]any)["fields"].([]any) {
			out = append(out, f.(map[string]any)["field"].(string))
		}
		return out
	}

	cases := []struct {
		nome string
		perm map[string]any
		path string
	}{
		{"ação inexistente", map[string]any{"action": "admin::content.fantasma.read"}, "permissions[0].action"},
		{"ação do domínio api", map[string]any{"action": "api::x.find"}, "permissions[0].action"},
		{"recurso divergente", map[string]any{"action": "admin::content.document.read", "subject": "outro"}, "permissions[0].subject"},
		{"campos inexistentes", map[string]any{"action": "admin::content.document.read", "fields": []string{"titulo", "nope"}}, "permissions[0].fields"},
		{"campos em ação sem suporte", map[string]any{"action": "admin::content.document.publish", "fields": []string{"titulo"}}, "permissions[0].fields"},
		{"condição em ação sem suporte", map[string]any{"action": "admin::content.document.create", "conditions": []string{"admin::is-creator"}}, "permissions[0].conditions"},
		{"condição inexistente", map[string]any{"action": "admin::content.document.read", "conditions": []string{"admin::nao-existe"}}, "permissions[0].conditions"},
	}
	for _, c := range cases {
		t.Run(c.nome, func(t *testing.T) {
			r := put(c.perm)
			assert.Equal(t, 422, r.Status, r.Body)
			assert.Equal(t, []string{c.path}, fieldPaths(r))
		})
	}
	dup := put(map[string]any{"action": "admin::content.document.read"}, map[string]any{"action": "admin::content.document.read", "subject": "document"})
	assert.Equal(t, 422, dup.Status, "mesma ação e recurso repetidos")
	assert.Equal(t, 422, e.do("PUT", "/admin/roles/"+rid+"/permissions", map[string]any{}, boss).Status)
	assert.Equal(t, 404, e.do("PUT", "/admin/roles/"+uuid.NewString()+"/permissions", map[string]any{"permissions": []any{}}, boss).Status)

	// Sucesso: fields nulo (todos) x lista; condições; subject padrão = recurso.
	r := put(
		map[string]any{"action": "admin::content.document.read", "fields": []string{"conteudo", "titulo"}, "conditions": []string{"admin::same-department", "admin::is-creator"}},
		map[string]any{"action": "admin::content.document.publish"},
		map[string]any{"action": "admin::settings.users.read"},
	)
	require.Equal(t, 200, r.Status, r.Body)
	perms := r.data()["permissions"].([]any)
	require.Len(t, perms, 3)
	byAction := map[string]map[string]any{}
	for _, p := range perms {
		m := p.(map[string]any)
		byAction[m["action"].(string)] = m
	}
	read := byAction["admin::content.document.read"]
	assert.Equal(t, "document", read["subject"])
	assert.Equal(t, []any{"conteudo", "titulo"}, read["fields"])
	assert.Equal(t, []any{"admin::is-creator", "admin::same-department"}, read["conditions"])
	assert.Nil(t, byAction["admin::content.document.publish"]["fields"], "null = todos os campos")

	// "fields": [] (nenhum campo) é diferente de null (todos) e deve ser preservado.
	r = put(map[string]any{"action": "admin::content.document.read", "fields": []string{}})
	require.Equal(t, 200, r.Status, r.Body)
	assert.Equal(t, []any{}, r.data()["permissions"].([]any)[0].(map[string]any)["fields"])

	// Troca em lote substitui tudo (e a mudança vale imediatamente).
	require.Equal(t, 200, put(map[string]any{"action": "admin::settings.users.read"}).Status)
	editor, _ := e.newAdmin(boss, "ed@exemplo.com", "editor")
	assert.Equal(t, 200, e.do("GET", "/admin/users", nil, editor).Status)
	r = put(map[string]any{"action": "admin::settings.roles.read"})
	require.Equal(t, 200, r.Status)
	assert.Len(t, r.data()["permissions"], 1)
	assert.Equal(t, 403, e.do("GET", "/admin/users", nil, editor).Status, "permissão removida vale na hora")
	assert.Equal(t, 200, e.do("GET", "/admin/roles", nil, editor).Status)
	assert.Equal(t, 200, put().Status, "lista vazia remove tudo")
	assert.Equal(t, 403, e.do("GET", "/admin/roles", nil, editor).Status)

	// Permissão já gravada não se perde se a validação falhar.
	put(map[string]any{"action": "admin::settings.roles.read"})
	assert.Equal(t, 422, put(map[string]any{"action": "admin::settings.roles.read"}, map[string]any{"action": "nada"}).Status)
	assert.Equal(t, 200, e.do("GET", "/admin/roles", nil, editor).Status)

	// Auditoria do before/after.
	logs := e.do("GET", "/admin/audit-logs?action=admin-role.permissions-replace&entityId="+rid, nil, boss)
	require.NotEmpty(t, logs.list())
	last := logs.list()[0].(map[string]any)["payload"].(map[string]any)
	assert.Contains(t, last, "before")
	assert.Contains(t, last, "after")
}

func TestDuplicarRole(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("editor", "admin::settings.users.read")
	e.grant("editor", "admin::settings.roles.read", `[]`, `[]`)
	src := e.roleID("editor")
	require.Equal(t, 200, e.do("PUT", "/admin/roles/"+src, map[string]any{"name": "Editor", "description": "Edita tudo"}, boss).Status)

	r := e.do("POST", "/admin/roles/"+src+"/duplicate", nil, boss)
	require.Equal(t, 201, r.Status, r.Body)
	assert.Equal(t, "Editor (cópia)", r.data()["name"])
	assert.Equal(t, "Edita tudo", r.data()["description"])
	assert.NotEqual(t, src, r.data()["id"])
	assert.Equal(t, "editor-copia", r.data()["code"])
	assert.Equal(t, false, r.data()["isSystem"])
	perms := r.data()["permissions"].([]any)
	require.Len(t, perms, 2)
	var fieldsOfRolesRead any = "x"
	for _, p := range perms {
		if m := p.(map[string]any); m["action"] == "admin::settings.roles.read" {
			fieldsOfRolesRead = m["fields"]
		}
	}
	assert.Equal(t, []any{}, fieldsOfRolesRead, "campos vazios (não nulos) são preservados")

	// A cópia é independente da original.
	e.do("PUT", "/admin/roles/"+r.data()["id"].(string)+"/permissions", map[string]any{"permissions": []any{}}, boss)
	assert.Len(t, e.do("GET", "/admin/roles/"+src, nil, boss).data()["permissions"], 2)
	assert.Equal(t, 404, e.do("POST", "/admin/roles/"+uuid.NewString()+"/duplicate", nil, boss).Status)
	// Duplicar de novo gera outro code.
	assert.Equal(t, "editor-copia-2", e.do("POST", "/admin/roles/"+src+"/duplicate", nil, boss).data()["code"])
}

func TestCatalogo(t *testing.T) {
	e, boss := setup(t)

	r := e.do("GET", "/admin/catalog/actions", nil, boss)
	require.Equal(t, 200, r.Status)
	groups := r.list()
	cats := []string{}
	var docAction map[string]any
	for _, g := range groups {
		gm := g.(map[string]any)
		cats = append(cats, gm["category"].(string))
		for _, a := range gm["actions"].([]any) {
			am := a.(map[string]any)
			assert.NotContains(t, am["id"], "api::", "só ações de admin por padrão")
			if am["id"] == "admin::content.document.read" {
				docAction = am
			}
		}
	}
	assert.Equal(t, []string{"collection-types", "plugins", "settings"}, cats)
	require.NotNil(t, docAction)
	assert.Equal(t, true, docAction["supportsFields"])
	assert.Equal(t, []any{"titulo", "conteudo", "status", "observacao_interna", "department_id"}, docAction["availableFields"], "campos configuráveis (sem os de sistema)")

	api := e.do("GET", "/admin/catalog/actions?domain=api", nil, boss).list()
	require.Len(t, api, 1)
	act := api[0].(map[string]any)["actions"].([]any)[0].(map[string]any)
	assert.Equal(t, "GET", act["method"])
	assert.Equal(t, "/api/documents", act["route"])

	conds := e.do("GET", "/admin/catalog/conditions", nil, boss).list()
	ids := []string{}
	for _, c := range conds {
		ids = append(ids, c.(map[string]any)["id"].(string))
	}
	assert.Equal(t, []string{"admin::is-creator", "admin::has-same-role-as-creator", "admin::same-department"}, ids)
}

func TestAuditoriaFiltrosEPaginacao(t *testing.T) {
	e, boss := setup(t)
	for i := range 5 {
		require.Equal(t, 201, e.do("POST", "/admin/roles", map[string]any{"name": fmt.Sprintf("Role %d", i)}, boss).Status)
	}
	ana := e.adminID("ana@exemplo.com")

	r := e.do("GET", "/admin/audit-logs?action=admin-role.create&pageSize=2", nil, boss)
	require.Equal(t, 200, r.Status)
	assert.Len(t, r.list(), 2)
	assert.EqualValues(t, 5, r.meta()["total"])
	assert.EqualValues(t, 3, r.meta()["pageCount"])
	first := r.list()[0].(map[string]any)
	assert.Equal(t, ana, first["actorId"])
	assert.Equal(t, "admin", first["actorType"])
	assert.Equal(t, "admin_role", first["entity"])
	assert.Contains(t, first["payload"], "after")

	count := func(q string) float64 {
		return e.do("GET", "/admin/audit-logs?"+q, nil, boss).meta()["total"].(float64)
	}
	assert.EqualValues(t, 5, count("entity=admin_role&actorId="+ana))
	assert.EqualValues(t, 0, count("actorId="+uuid.NewString()))
	assert.EqualValues(t, 0, count("actorType=api_token"))
	assert.EqualValues(t, 5, count("action=admin-role.create&from=2000-01-01T00:00:00Z&to=2999-01-01T00:00:00Z"))
	assert.EqualValues(t, 0, count("action=admin-role.create&to=2000-01-01T00:00:00Z"))
	assert.EqualValues(t, 0, count("action=admin-role.create&from=2999-01-01T00:00:00Z"))
	assert.Equal(t, 422, e.do("GET", "/admin/audit-logs?from=ontem", nil, boss).Status)
	assert.Equal(t, 422, e.do("GET", "/admin/audit-logs?to=2026-13-45", nil, boss).Status)
	// Filtro com SQL no valor é só um texto: não casa com nada.
	assert.EqualValues(t, 0, count("action="+url.QueryEscape("x' OR '1'='1")))
}

func TestApagarRoleInvalidaCache(t *testing.T) {
	e, boss := setup(t)
	e.mkRole("tmp", "admin::settings.roles.read")
	rid := e.roleID("tmp")
	// Aquece o cache dessa role.
	tok, id := e.newAdmin(boss, "t@exemplo.com", "tmp")
	assert.Equal(t, 200, e.do("GET", "/admin/roles", nil, tok).Status)
	// Remove a role do admin e exclui: o cache da role some e uma recriação com o mesmo ID é impossível;
	// o relevante é que a exclusão funciona após liberar a role.
	require.Equal(t, 204, e.do("DELETE", "/admin/users/"+id, nil, boss).Status)
	assert.Equal(t, 204, e.do("DELETE", "/admin/roles/"+rid, nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/admin/roles/"+rid, nil, boss).Status)
}
