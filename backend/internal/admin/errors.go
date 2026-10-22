package admin

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

var (
	errSuperAdminImmutable = httpapi.ErrConflict("super_admin_immutable", "A role Super Admin não pode ter permissões alteradas, ser duplicada nem ser excluída.")
	errLastSuperAdmin      = httpapi.ErrConflict("last_super_admin", "É necessário manter ao menos um administrador ativo com a role Super Admin.")
	errCannotModifySelf    = httpapi.ErrConflict("cannot_modify_self", "Você não pode executar esta ação sobre a sua própria conta.")
	errCannotModifyOwnRole = httpapi.ErrConflict("cannot_modify_own_role", "Você não pode alterar as permissões de uma role que você possui.")

	errTargetMorePrivileged = &httpapi.APIError{Status: 403, Code: "target_more_privileged", Message: "Este administrador tem permissões que você não possui; apenas quem tem acesso equivalente ou maior pode alterá-lo."}
)

// errPrivilegeEscalation recusa conceder o que o próprio ator não possui.
func errPrivilegeEscalation(missing []authz.Permission) error {
	return &httpapi.APIError{Status: 403, Code: "privilege_escalation", Message: "Você só pode conceder permissões que você mesmo possui.",
		Details: map[string]any{"permissions": permViews(missing)}}
}

func errRoleInUse(n int64) error {
	e := httpapi.ErrConflict("role_in_use", fmt.Sprintf("A role está atribuída a %d administrador(es) e não pode ser excluída. Remova-a deles antes.", n))
	e.Details = map[string]any{"userCount": n}
	return e
}

func fieldIssue(field, msg string) error {
	return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: field, Message: msg}}})
}

// notFound converte "sem linhas" em 404.
func notFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return httpapi.ErrNotFound()
	}
	return err
}
