package authz

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"strings"
)

var identRe = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// FieldInfo descreve um campo de um recurso.
type FieldInfo struct {
	Name   string // nome lógico (JSON); usado em permissões e filtros
	Column string // coluna no banco
	System bool   // sempre legível e nunca gravável pelo cliente
}

// Schema descreve os campos de um recurso. É construído a partir das tags da
// struct do modelo, o que dispensa código específico por recurso:
//
//	type Document struct {
//	    ID    uuid.UUID `json:"id" db:"id" authz:"system"`
//	    Titulo string   `json:"titulo" db:"titulo"`
//	}
//
// "json" define o nome lógico, "db" a coluna (padrão: o nome JSON) e
// authz:"system" marca campos de sistema.
type Schema struct {
	Name   string
	fields []FieldInfo
	byName map[string]FieldInfo
}

// NewSchema lê as tags do modelo (struct ou ponteiro para struct).
func NewSchema(name string, model any) (*Schema, error) {
	t := reflect.TypeOf(model)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("schema %q: o modelo deve ser uma struct", name)
	}
	s := &Schema{Name: name, byName: map[string]FieldInfo{}}
	for i := range t.NumField() {
		f := t.Field(i)
		jsonName, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if jsonName == "" || jsonName == "-" {
			continue
		}
		col := f.Tag.Get("db")
		if col == "" {
			col = jsonName
		}
		if !identRe.MatchString(col) {
			return nil, fmt.Errorf("schema %q: coluna inválida %q", name, col)
		}
		fi := FieldInfo{Name: jsonName, Column: col, System: f.Tag.Get("authz") == "system"}
		if _, dup := s.byName[fi.Name]; dup {
			return nil, fmt.Errorf("schema %q: campo duplicado %q", name, fi.Name)
		}
		s.fields = append(s.fields, fi)
		s.byName[fi.Name] = fi
	}
	return s, nil
}

// Fields lista os nomes de todos os campos.
func (s *Schema) Fields() []string {
	out := make([]string, len(s.fields))
	for i, f := range s.fields {
		out[i] = f.Name
	}
	return out
}

// Configurable lista os campos que podem aparecer em permissões (não-sistema).
func (s *Schema) Configurable() []string {
	var out []string
	for _, f := range s.fields {
		if !f.System {
			out = append(out, f.Name)
		}
	}
	return out
}

// Has informa se o campo existe.
func (s *Schema) Has(name string) bool { _, ok := s.byName[name]; return ok }

// IsSystem informa se é campo de sistema.
func (s *Schema) IsSystem(name string) bool { return s.byName[name].System }

// Column resolve o nome lógico para a coluna; serve de lista de campos
// permitidos para Filter.ToSQL.
func (s *Schema) Column(name string) (string, bool) {
	f, ok := s.byName[name]
	return f.Column, ok
}

// UnknownFields devolve, ordenados, os nomes que não existem no recurso.
func (s *Schema) UnknownFields(names []string) []string {
	var out []string
	for _, n := range names {
		if !s.Has(n) {
			out = append(out, n)
		}
	}
	slices.Sort(out)
	return out
}

// FilterRead remove do mapa os campos não permitidos. Campos de sistema são
// sempre mantidos; chaves que não pertencem ao schema são descartadas. Retorna uma cópia.
func (s *Schema) FilterRead(m map[string]any, allowed FieldSet) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		// Chaves fora do schema nunca saem, mesmo com "todos os campos".
		if s.Has(k) && (allowed.Allows(k) || s.IsSystem(k)) {
			out[k] = v
		}
	}
	return out
}

// Project serializa v (struct) em JSON e aplica FilterRead.
func (s *Schema) Project(v any, allowed FieldSet) (map[string]any, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return s.FilterRead(m, allowed), nil
}

// RejectedWrite devolve, ordenados e sem repetição, os campos que o cliente
// não pode gravar: campos de sistema e campos fora do conjunto permitido.
func (s *Schema) RejectedWrite(keys []string, allowed FieldSet) []string {
	var out []string
	for _, k := range keys {
		if s.IsSystem(k) || !allowed.Allows(k) {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return slices.Compact(out)
}
