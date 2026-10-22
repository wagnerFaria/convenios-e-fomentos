package admin

import (
	"context"
	"slices"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

// As verificações abaixo impedem escalonamento de privilégio: quem não é Super Admin só concede
// o que possui, não altera a própria conta nem as roles que tem, e não age sobre quem tem mais poder.

// ensureCovers recusa se o ator não possuir todas as permissões informadas.
func (s *Service) ensureCovers(ctx context.Context, actor Actor, perms []authz.Permission) error {
	missing, err := s.engine.Covers(ctx, actor.Subject, perms)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return errPrivilegeEscalation(missing)
	}
	return nil
}

// ensureCoversRoles recusa atribuir roles cujas permissões o ator não possui.
func (s *Service) ensureCoversRoles(ctx context.Context, actor Actor, roleIDs []uuid.UUID) error {
	missing, err := s.engine.CoversRoles(ctx, actor.Subject, roleIDs)
	if err != nil {
		return err
	}
	if len(missing) > 0 {
		return errPrivilegeEscalation(missing)
	}
	return nil
}

// ensureNotOwnAccount impede que quem não é Super Admin altere perfis ou lotação da própria conta.
func ensureNotOwnAccount(actor Actor, id uuid.UUID) error {
	if actor.ID == id && !actor.Subject.IsSuperAdmin() {
		return errCannotModifySelf
	}
	return nil
}

// ensureCanManage recusa agir sobre administradores com permissões que o ator não possui.
func (s *Service) ensureCanManage(ctx context.Context, q *db.Queries, actor Actor, targets ...db.Admin) error {
	if actor.Subject.IsSuperAdmin() {
		return nil
	}
	for _, t := range targets {
		if t.ID == actor.ID {
			continue
		}
		roles, err := q.ListRolesForAdmin(ctx, t.ID)
		if err != nil {
			return err
		}
		target := authz.Subject{Type: authz.SubjectAdmin, ID: t.ID, DepartmentID: t.DepartmentID}
		for _, r := range roles {
			target.Roles = append(target.Roles, authz.Role{ID: r.ID, Code: r.Code})
		}
		missing, err := s.engine.CoversSubject(ctx, actor.Subject, target)
		if err != nil {
			return err
		}
		if len(missing) > 0 {
			return errTargetMorePrivileged
		}
	}
	return nil
}

// ownsRole informa se o ator possui a role.
func ownsRole(actor Actor, roleID uuid.UUID) bool {
	return slices.Contains(actor.Subject.RoleIDs(), roleID)
}
