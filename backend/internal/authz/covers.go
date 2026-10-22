package authz

import (
	"context"
	"slices"

	"github.com/google/uuid"
)

// withinFields informa se a lista de campos p cabe em q (nil = todos os campos).
func withinFields(p, q []string) bool {
	if q == nil {
		return true
	}
	if p == nil {
		return false
	}
	for _, f := range p {
		if !slices.Contains(q, f) {
			return false
		}
	}
	return true
}

// withinConditions informa se as condições p são tão restritivas quanto as de q.
// Condições se combinam com OR: sem condição é o mais amplo, e menos condições é mais restrito.
func withinConditions(p, q []string) bool {
	if len(q) == 0 {
		return true
	}
	if len(p) == 0 {
		return false
	}
	for _, c := range p {
		if !slices.Contains(q, c) {
			return false
		}
	}
	return true
}

type permKey struct{ action, subject string }

// keyOf identifica a permissão por ação e recurso. Recurso vazio vale o recurso da ação
// registrada, pois permissões gravadas sem recurso e as normalizadas são equivalentes.
func (e *Engine) keyOf(action, subject string) permKey {
	if subject == "" {
		if a, ok := e.actions.Get(action); ok {
			subject = a.Resource
		}
	}
	return permKey{action, subject}
}

// Widened devolve as permissões de after que ampliam o que before já concedia:
// as novas e as mais amplas (mais campos ou menos condições) que a de mesma ação e recurso.
// Permissões iguais ou mais restritas não aparecem.
func (e *Engine) Widened(before, after []Permission) []Permission {
	prev := make(map[permKey]Permission, len(before))
	for _, p := range before {
		prev[e.keyOf(p.Action, p.Subject)] = p
	}
	var out []Permission
	for _, p := range after {
		q, ok := prev[e.keyOf(p.Action, p.Subject)]
		if ok && withinFields(p.Fields, q.Fields) && withinConditions(p.Conditions, q.Conditions) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// Covers devolve as permissões de perms que o sujeito não cobre, ou seja, que lhe dariam
// mais poder do que ele já tem. Serve para impedir escalonamento de privilégio: só se
// concede o que se possui. Super Admin cobre tudo.
func (e *Engine) Covers(ctx context.Context, s Subject, perms []Permission) ([]Permission, error) {
	if s.IsSuperAdmin() {
		return nil, nil
	}
	eff, err := e.Effective(ctx, s)
	if err != nil {
		return nil, err
	}
	have := make(map[permKey]EffectivePermission, len(eff))
	for _, p := range eff {
		have[e.keyOf(p.Action, p.Subject)] = p
	}
	var missing []Permission
	for _, p := range perms {
		h, ok := have[e.keyOf(p.Action, p.Subject)]
		if !ok || !withinFields(p.Fields, h.Fields) || !withinConditions(p.Conditions, h.Conditions) {
			missing = append(missing, p)
		}
	}
	return missing, nil
}

// CoversRoles é como Covers, para o conjunto de permissões das roles informadas.
func (e *Engine) CoversRoles(ctx context.Context, s Subject, roleIDs []uuid.UUID) ([]Permission, error) {
	if s.IsSuperAdmin() {
		return nil, nil
	}
	var perms []Permission
	for _, id := range roleIDs {
		rp, err := e.cache.get(ctx, id)
		if err != nil {
			return nil, err
		}
		perms = append(perms, rp...)
	}
	return e.Covers(ctx, s, perms)
}

// CoversSubject é como Covers, para as permissões efetivas de outro sujeito. Um ator só pode
// agir sobre quem não tem mais poder do que ele.
func (e *Engine) CoversSubject(ctx context.Context, actor, target Subject) ([]Permission, error) {
	if actor.IsSuperAdmin() {
		return nil, nil
	}
	eff, err := e.Effective(ctx, target)
	if err != nil {
		return nil, err
	}
	perms := make([]Permission, len(eff))
	for i, p := range eff {
		perms[i] = Permission{Action: p.Action, Subject: p.Subject, Fields: p.Fields, Conditions: p.Conditions}
	}
	return e.Covers(ctx, actor, perms)
}
