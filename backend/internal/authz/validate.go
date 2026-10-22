package authz

import (
	"fmt"
	"slices"
)

// Issue é um problema encontrado ao validar permissões. Path aponta o item,
// por exemplo "permissions[2].fields".
type Issue struct {
	Path    string
	Message string
}

// RegisterSchema associa o schema de campos a um recurso, para validar as
// listas de campos das permissões.
func (r *Registry) RegisterSchema(resource string, s *Schema) {
	if r.schemas == nil {
		r.schemas = map[string]*Schema{}
	}
	r.schemas[resource] = s
}

// Schema devolve o schema registrado do recurso.
func (r *Registry) Schema(resource string) (*Schema, bool) {
	s, ok := r.schemas[resource]
	return s, ok
}

// NormalizePermissions valida a lista de permissões de uma role contra o
// registro de ações, condições e schemas, e devolve a lista normalizada
// (subject padrão = recurso da ação; conditions nunca nulo). Não grava nada.
func (e *Engine) NormalizePermissions(in []Permission) ([]Permission, []Issue) {
	var issues []Issue
	out := make([]Permission, 0, len(in))
	seen := map[[2]string]bool{}
	for i, p := range in {
		path := func(field string) string { return fmt.Sprintf("permissions[%d].%s", i, field) }
		a, ok := e.actions.Get(p.Action)
		if !ok || a.Domain != DomainAdmin {
			issues = append(issues, Issue{path("action"), "Ação desconhecida ou fora do domínio administrativo."})
			continue
		}
		if p.Subject == "" {
			p.Subject = a.Resource
		} else if a.Resource != "" && p.Subject != a.Resource {
			issues = append(issues, Issue{path("subject"), fmt.Sprintf("O recurso desta ação é %q.", a.Resource)})
			continue
		}
		key := [2]string{p.Action, p.Subject}
		if seen[key] {
			issues = append(issues, Issue{path("action"), "Permissão repetida para a mesma ação e recurso."})
			continue
		}
		seen[key] = true

		if p.Fields != nil {
			if !a.SupportsFields {
				issues = append(issues, Issue{path("fields"), "Esta ação não suporta restrição de campos."})
			} else if s, ok := e.actions.Schema(a.Resource); ok {
				if unknown := s.UnknownFields(p.Fields); len(unknown) > 0 {
					issues = append(issues, Issue{path("fields"), fmt.Sprintf("Campos inexistentes: %v.", unknown)})
				}
			}
			// Atenção: lista vazia ("nenhum campo") não pode virar nil ("todos os campos").
			p.Fields = append([]string{}, slices.Compact(slices.Sorted(slices.Values(p.Fields)))...)
		}
		if len(p.Conditions) > 0 && !a.SupportsConditions {
			issues = append(issues, Issue{path("conditions"), "Esta ação não suporta condições."})
		}
		for _, c := range p.Conditions {
			if _, ok := e.conds.Get(c); !ok {
				issues = append(issues, Issue{path("conditions"), fmt.Sprintf("Condição desconhecida: %q.", c)})
			}
		}
		p.Conditions = append([]string{}, slices.Compact(slices.Sorted(slices.Values(p.Conditions)))...)
		if p.Properties == nil {
			p.Properties = map[string]any{}
		}
		out = append(out, p)
	}
	return out, issues
}
