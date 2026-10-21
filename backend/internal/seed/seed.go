// Pacote seed cria os dados iniciais do sistema. É idempotente e não destrutivo: rodar de
// novo não duplica nada nem desfaz alterações feitas depois pelos administradores.
//
//   - Roles de admin faltantes são criadas, com as permissões padrão aplicadas somente na criação.
//   - As permissões padrão de usuário final (Public/Authenticated) são aplicadas uma única vez
//     (controlado pela configuração "seed.enduser-permissions").
//   - Em desenvolvimento, cria um usuário de cada role; a senha só aparece no log na criação.
package seed

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/document"
	"github.com/wagner/convenios-e-fomentos/backend/internal/enduser"
)

// Deps são as dependências do seed.
type Deps struct {
	Pool        *pgxpool.Pool
	Hasher      auth.Hasher
	Engine      *authz.Engine
	EndUsers    *enduser.Service
	Log         *slog.Logger
	Development bool
}

// Códigos das roles de administrador criadas pelo seed.
const (
	RoleSuperAdmin = authz.SuperAdminRoleCode
	RoleEditor     = "editor"
	RoleAuthor     = "author"
	RoleReviewer   = "department-reviewer"
)

// Departamentos de exemplo.
var Departments = []string{"Financeiro", "Jurídico"}

const endUserFlag = "seed.enduser-permissions"

type roleSpec struct {
	code, name, description string
	system                  bool
	perms                   []authz.Permission
}

func adminRoles() []roleSpec {
	return []roleSpec{
		{code: RoleSuperAdmin, name: "Super Admin", description: "Acesso total ao sistema.", system: true},
		{
			code: RoleEditor, name: "Editor", description: "Cria, lê, atualiza, exclui e publica qualquer documento.", system: true,
			perms: []authz.Permission{
				{Action: document.ActionAdminCreate}, {Action: document.ActionAdminRead}, {Action: document.ActionAdminUpdate},
				{Action: document.ActionAdminDelete}, {Action: document.ActionAdminPublish},
			},
		},
		{
			code: RoleAuthor, name: "Author", description: "Cria e lê documentos; atualiza e exclui apenas os que criou. Não publica.", system: true,
			perms: []authz.Permission{
				{Action: document.ActionAdminCreate, Fields: []string{"titulo", "conteudo", "observacao_interna", "department_id"}},
				{Action: document.ActionAdminRead},
				{Action: document.ActionAdminUpdate, Fields: []string{"titulo", "conteudo", "observacao_interna", "department_id"}, Conditions: []string{"admin::is-creator"}},
				{Action: document.ActionAdminDelete, Conditions: []string{"admin::is-creator"}},
			},
		},
		{
			code: RoleReviewer, name: "Revisor do Departamento", description: "Lê documentos; só altera status e observação interna dos documentos do próprio departamento.",
			perms: []authz.Permission{
				{Action: document.ActionAdminRead},
				{Action: document.ActionAdminUpdate, Fields: []string{"status", "observacao_interna"}, Conditions: []string{"admin::same-department"}},
				{Action: document.ActionAdminPublish, Conditions: []string{"admin::same-department"}},
			},
		},
	}
}

var endUserDefaults = map[string][]string{
	enduser.RolePublic:        {document.ActionAPIFind, document.ActionAPIFindOne},
	enduser.RoleAuthenticated: {document.ActionAPIFind, document.ActionAPIFindOne, document.ActionAPICreate, document.ActionAPIUpdate, document.ActionAPIDelete},
}

// Run executa o seed.
func Run(ctx context.Context, d Deps) error {
	q := db.New(d.Pool)
	if err := d.EndUsers.EnsureDefaults(ctx); err != nil {
		return fmt.Errorf("padrões de usuário final: %w", err)
	}
	for _, name := range Departments {
		if _, err := q.UpsertDepartment(ctx, name); err != nil {
			return fmt.Errorf("departamento %q: %w", name, err)
		}
	}
	for _, spec := range adminRoles() {
		if err := d.ensureAdminRole(ctx, q, spec); err != nil {
			return fmt.Errorf("role %q: %w", spec.code, err)
		}
	}
	if err := d.ensureEndUserPermissions(ctx, q); err != nil {
		return err
	}
	if d.Development {
		if err := d.devUsers(ctx, q); err != nil {
			return err
		}
	}
	// Avisa instâncias em execução (LISTEN/NOTIFY) que as permissões podem ter mudado.
	return d.Engine.Invalidate(ctx)
}

func (d Deps) ensureAdminRole(ctx context.Context, q *db.Queries, spec roleSpec) error {
	exists, err := q.AdminRoleCodeExists(ctx, spec.code)
	if err != nil {
		return err
	}
	role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: spec.code, Name: spec.name, Description: spec.description, IsSystem: spec.system})
	if err != nil {
		return err
	}
	if exists || len(spec.perms) == 0 {
		return nil
	}
	// Role recém-criada: aplica as permissões padrão, validadas pelo mesmo caminho da API.
	perms, issues := d.Engine.NormalizePermissions(spec.perms)
	if len(issues) > 0 {
		return fmt.Errorf("permissões padrão inválidas: %v", issues)
	}
	return database.WithTx(ctx, d.Pool, func(tx pgx.Tx) error {
		tq := q.WithTx(tx)
		for _, p := range perms {
			fields, err := marshalFields(p.Fields)
			if err != nil {
				return err
			}
			conds, _ := json.Marshal(p.Conditions)
			props, _ := json.Marshal(p.Properties)
			var subject *string
			if p.Subject != "" {
				subject = &p.Subject
			}
			if _, err := tq.CreateAdminPermission(ctx, db.CreateAdminPermissionParams{RoleID: role.ID, Action: p.Action, Subject: subject, Fields: fields, Conditions: conds, Properties: props}); err != nil {
				return err
			}
		}
		return nil
	})
}

func marshalFields(f []string) ([]byte, error) {
	if f == nil {
		return nil, nil
	}
	return json.Marshal(f)
}

func (d Deps) ensureEndUserPermissions(ctx context.Context, q *db.Queries) error {
	if _, err := q.GetSetting(ctx, endUserFlag); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	return database.WithTx(ctx, d.Pool, func(tx pgx.Tx) error {
		tq := q.WithTx(tx)
		for typ, actions := range endUserDefaults {
			role, err := tq.GetEndUserRoleByType(ctx, typ)
			if err != nil {
				return fmt.Errorf("role %q: %w", typ, err)
			}
			for _, a := range actions {
				if _, ok := d.Engine.Actions().Get(a); !ok {
					return fmt.Errorf("ação %q não registrada", a)
				}
				if err := tq.InsertEndUserPermission(ctx, db.InsertEndUserPermissionParams{RoleID: role.ID, Action: a}); err != nil {
					return err
				}
			}
		}
		_, err := tq.UpsertSetting(ctx, db.UpsertSettingParams{Key: endUserFlag, Value: []byte("true")})
		return err
	})
}

type devUser struct {
	email, first, last, role, department string
}

func (d Deps) devUsers(ctx context.Context, q *db.Queries) error {
	users := []devUser{
		{"super@dev.local", "Super", "Admin", RoleSuperAdmin, ""},
		{"editor@dev.local", "Eda", "Editora", RoleEditor, ""},
		{"author@dev.local", "Augusto", "Autor", RoleAuthor, "Financeiro"},
		{"revisor@dev.local", "Regina", "Revisora", RoleReviewer, "Financeiro"},
	}
	for _, u := range users {
		if err := d.devAdmin(ctx, q, u); err != nil {
			return fmt.Errorf("usuário de desenvolvimento %s: %w", u.email, err)
		}
	}
	return d.devEndUser(ctx, q)
}

func randomPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (d Deps) devAdmin(ctx context.Context, q *db.Queries, u devUser) error {
	if _, err := q.GetAdminByEmail(ctx, u.email); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	role, err := q.UpsertAdminRole(ctx, db.UpsertAdminRoleParams{Code: u.role})
	if err != nil {
		return err
	}
	var dept *uuid.UUID
	if u.department != "" {
		rows, err := q.ListDepartmentsByName(ctx, []string{u.department})
		if err != nil || len(rows) != 1 {
			return fmt.Errorf("departamento %q: %w", u.department, err)
		}
		dept = &rows[0].ID
	}
	password, err := randomPassword()
	if err != nil {
		return err
	}
	hash, err := d.Hasher.Hash(password)
	if err != nil {
		return err
	}
	err = database.WithTx(ctx, d.Pool, func(tx pgx.Tx) error {
		tq := q.WithTx(tx)
		a, err := tq.CreateAdmin(ctx, db.CreateAdminParams{FirstName: u.first, LastName: u.last, Email: u.email, PasswordHash: &hash, IsActive: true, DepartmentID: dept})
		if err != nil {
			return err
		}
		return tq.AssignAdminRole(ctx, db.AssignAdminRoleParams{AdminID: a.ID, RoleID: role.ID})
	})
	if err != nil {
		return err
	}
	d.Log.Info("usuário de desenvolvimento criado (a senha não será exibida novamente)", "tipo", "admin", "role", u.role, "email", u.email, "senha", password)
	return nil
}

func (d Deps) devEndUser(ctx context.Context, q *db.Queries) error {
	const email, username = "usuario@dev.local", "usuario"
	if users, err := q.GetEndUsersByEmail(ctx, email); err != nil || len(users) > 0 {
		return err
	}
	role, err := q.GetEndUserRoleByType(ctx, enduser.RoleAuthenticated)
	if err != nil {
		return err
	}
	password, err := randomPassword()
	if err != nil {
		return err
	}
	hash, err := d.Hasher.Hash(password)
	if err != nil {
		return err
	}
	if _, err := q.CreateEndUser(ctx, db.CreateEndUserParams{Username: username, Email: email, PasswordHash: &hash, Confirmed: true, RoleID: role.ID, Profile: []byte("{}")}); err != nil {
		return err
	}
	d.Log.Info("usuário de desenvolvimento criado (a senha não será exibida novamente)", "tipo", "usuário final", "role", enduser.RoleAuthenticated, "email", email, "usuario", username, "senha", password)
	return nil
}
