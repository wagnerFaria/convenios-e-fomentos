package admin

import (
	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/audit"
	"github.com/wagner/convenios-e-fomentos/backend/internal/auth"
	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
)

// Actor é o administrador que executa uma operação, usado para autorização
// fina (ex.: só Super Admin concede a role Super Admin) e auditoria.
type Actor struct {
	ID      uuid.UUID
	Subject authz.Subject
	IP      string
}

// ActorOf monta o Actor a partir do administrador autenticado.
func ActorOf(p *auth.AdminPrincipal, ip string) Actor {
	return Actor{ID: p.ID, Subject: SubjectOf(p), IP: ip}
}

func (a Actor) entry(action, entity string, entityID uuid.UUID, before, after any) audit.Entry {
	payload := map[string]any{}
	if before != nil {
		payload["before"] = before
	}
	if after != nil {
		payload["after"] = after
	}
	return audit.Entry{ActorType: "admin", ActorID: a.ID.String(), Action: action, Entity: entity, EntityID: entityID.String(), Payload: payload, IP: a.IP}
}
