package authz

import (
	"slices"
	"sort"
)

// FieldSet é um conjunto de campos permitidos; o valor zero não permite nenhum.
type FieldSet struct {
	all   bool
	names map[string]struct{}
}

// AllFields permite todos os campos.
func AllFields() FieldSet { return FieldSet{all: true} }

// Fields permite apenas os campos informados.
func Fields(names ...string) FieldSet {
	m := make(map[string]struct{}, len(names))
	for _, n := range names {
		m[n] = struct{}{}
	}
	return FieldSet{names: m}
}

// IsAll informa se todos os campos são permitidos.
func (s FieldSet) IsAll() bool { return s.all }

// Allows informa se o campo é permitido.
func (s FieldSet) Allows(name string) bool {
	if s.all {
		return true
	}
	_, ok := s.names[name]
	return ok
}

// Names devolve os nomes em ordem alfabética (vazio quando IsAll).
func (s FieldSet) Names() []string {
	out := make([]string, 0, len(s.names))
	for n := range s.names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// Union une dois conjuntos; "todos" prevalece.
func (s FieldSet) Union(o FieldSet) FieldSet {
	if s.all || o.all {
		return AllFields()
	}
	return Fields(slices.Concat(s.Names(), o.Names())...)
}

// fieldSetFromList converte a lista persistida (nil = todos os campos).
func fieldSetFromList(list []string) FieldSet {
	if list == nil {
		return AllFields()
	}
	return Fields(list...)
}
