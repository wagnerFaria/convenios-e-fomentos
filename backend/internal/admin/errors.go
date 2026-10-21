package admin

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

var (
	errSuperAdminImmutable = httpapi.ErrConflict("super_admin_immutable", "A role Super Admin não pode ter permissões alteradas, ser duplicada nem ser excluída.")
	errLastSuperAdmin      = httpapi.ErrConflict("last_super_admin", "É necessário manter ao menos um administrador ativo com a role Super Admin.")
	errCannotModifySelf    = httpapi.ErrConflict("cannot_modify_self", "Você não pode executar esta ação sobre a sua própria conta.")
	errInvitationPending   = httpapi.ErrConflict("invitation_pending", "O convite ainda não foi aceito; a conta será ativada quando o administrador definir a senha.")
)

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
