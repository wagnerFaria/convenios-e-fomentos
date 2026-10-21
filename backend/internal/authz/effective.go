package authz

import (
	"context"
	"slices"
)

// EffectivePermission é a permissão efetiva de um sujeito para uma ação/recurso,
// já unida entre todas as roles. Serve para o front esconder botões.
type EffectivePermission struct {
	Action     string   `json:"action"`
	Subject    string   `json:"subject,omitempty"`
	Fields     []string `json:"fields"`     // null = todos os campos
	Conditions []string `json:"conditions"` // vazio = sem condição
}

// Effective calcula as permissões efetivas do sujeito. Super Admin recebe todas as ações de admin.
func (e *Engine) Effective(ctx context.Context, s Subject) ([]EffectivePermission, error) {
	if s.Type != SubjectAdmin {
		return []EffectivePermission{}, nil
	}
	if s.IsSuperAdmin() {
		out := []EffectivePermission{}
		for _, a := range e.actions.All() {
			if a.Domain == DomainAdmin {
				out = append(out, EffectivePermission{Action: a.ID, Subject: a.Resource, Conditions: []string{}})
			}
		}
		return out, nil
	}
	type key struct{ action, subject string }
	merged := map[key]*EffectivePermission{}
	var order []key
	for _, role := range s.Roles {
		perms, err := e.cache.get(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, p := range perms {
			if _, ok := e.actions.Get(p.Action); !ok {
				continue
			}
			k := key{p.Action, p.Subject}
			cur, seen := merged[k]
			if !seen {
				merged[k] = &EffectivePermission{Action: p.Action, Subject: p.Subject, Fields: slices.Clone(p.Fields), Conditions: slices.Clone(p.Conditions)}
				if merged[k].Conditions == nil {
					merged[k].Conditions = []string{}
				}
				order = append(order, k)
				continue
			}
			// Campos: "todos" prevalece. Condições: sem condição prevalece.
			if cur.Fields != nil {
				if p.Fields == nil {
					cur.Fields = nil
				} else {
					cur.Fields = fieldSetFromList(cur.Fields).Union(Fields(p.Fields...)).Names()
				}
			}
			if len(cur.Conditions) > 0 {
				if len(p.Conditions) == 0 {
					cur.Conditions = []string{}
				} else {
					for _, c := range p.Conditions {
						if !slices.Contains(cur.Conditions, c) {
							cur.Conditions = append(cur.Conditions, c)
						}
					}
				}
			}
		}
	}
	out := make([]EffectivePermission, 0, len(order))
	for _, k := range order {
		out = append(out, *merged[k])
	}
	return out, nil
}
