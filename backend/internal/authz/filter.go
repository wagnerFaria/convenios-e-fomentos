package authz

import (
	"fmt"
	"reflect"
	"strings"
)

// Op é um operador de comparação permitido em filtros.
type Op string

// Operadores suportados.
const (
	OpEq  Op = "eq"
	OpNeq Op = "neq"
	OpIn  Op = "in"
)

type filterKind int

const (
	kindTrue filterKind = iota
	kindFalse
	kindLeaf
	kindAnd
	kindOr
)

// Filter é uma representação tipada e segura de uma restrição de registros:
// comparações (campo + operador + valor) combinadas com and/or. Nunca contém
// SQL: a conversão para WHERE é parametrizada e só aceita campos de uma lista
// explícita (ver Schema e Filter.ToSQL).
type Filter struct {
	kind     filterKind
	Field    string
	Op       Op
	Values   []any
	Children []Filter
}

// Getter lê o valor de um campo lógico de um registro em memória.
type Getter func(field string) (any, bool)

// FilterTrue não restringe nada.
func FilterTrue() Filter { return Filter{kind: kindTrue} }

// FilterFalse não casa com nenhum registro.
func FilterFalse() Filter { return Filter{kind: kindFalse} }

// Eq casa registros em que field == v. Valor nulo nunca casa (semântica do SQL).
func Eq(field string, v any) Filter {
	if isNil(v) {
		return FilterFalse()
	}
	return Filter{kind: kindLeaf, Field: field, Op: OpEq, Values: []any{v}}
}

// Neq casa registros em que field != v (campos nulos não casam).
func Neq(field string, v any) Filter {
	if isNil(v) {
		return FilterFalse()
	}
	return Filter{kind: kindLeaf, Field: field, Op: OpNeq, Values: []any{v}}
}

// In casa registros cujo field está em vals. Lista vazia não casa com nada.
func In(field string, vals ...any) Filter {
	clean := make([]any, 0, len(vals))
	for _, v := range vals {
		if !isNil(v) {
			clean = append(clean, v)
		}
	}
	if len(clean) == 0 {
		return FilterFalse()
	}
	return Filter{kind: kindLeaf, Field: field, Op: OpIn, Values: clean}
}

// And combina filtros com E, simplificando constantes.
func And(fs ...Filter) Filter {
	var out []Filter
	for _, f := range fs {
		switch f.kind {
		case kindTrue:
		case kindFalse:
			return FilterFalse()
		case kindAnd:
			out = append(out, f.Children...)
		default:
			out = append(out, f)
		}
	}
	switch len(out) {
	case 0:
		return FilterTrue()
	case 1:
		return out[0]
	}
	return Filter{kind: kindAnd, Children: out}
}

// Or combina filtros com OU, simplificando constantes.
func Or(fs ...Filter) Filter {
	var out []Filter
	for _, f := range fs {
		switch f.kind {
		case kindFalse:
		case kindTrue:
			return FilterTrue()
		case kindOr:
			out = append(out, f.Children...)
		default:
			out = append(out, f)
		}
	}
	switch len(out) {
	case 0:
		return FilterFalse()
	case 1:
		return out[0]
	}
	return Filter{kind: kindOr, Children: out}
}

// IsTrue informa se o filtro não restringe nada.
func (f Filter) IsTrue() bool { return f.kind == kindTrue }

// IsFalse informa se o filtro nega tudo.
func (f Filter) IsFalse() bool { return f.kind == kindFalse }

// Match avalia o filtro contra um registro em memória. Campo ausente ou nulo
// nunca casa, espelhando o comportamento do SQL.
func (f Filter) Match(get Getter) bool {
	switch f.kind {
	case kindTrue:
		return true
	case kindFalse:
		return false
	case kindAnd:
		for _, c := range f.Children {
			if !c.Match(get) {
				return false
			}
		}
		return true
	case kindOr:
		for _, c := range f.Children {
			if c.Match(get) {
				return true
			}
		}
		return false
	}
	raw, ok := get(f.Field)
	if !ok {
		return false
	}
	got, ok := normalize(raw)
	if !ok {
		return false
	}
	switch f.Op {
	case OpEq:
		want, _ := normalize(f.Values[0])
		return got == want
	case OpNeq:
		want, _ := normalize(f.Values[0])
		return got != want
	case OpIn:
		for _, v := range f.Values {
			if want, _ := normalize(v); want == got {
				return true
			}
		}
	}
	return false
}

// ToSQL converte o filtro em uma cláusula WHERE parametrizada ($n a partir de
// firstArg). column resolve o nome lógico do campo para a coluna; campos fora da
// lista permitida geram erro. Nenhum valor é concatenado ao SQL.
func (f Filter) ToSQL(column func(field string) (string, bool), firstArg int) (string, []any, error) {
	var args []any
	next := firstArg
	var build func(Filter) (string, error)
	build = func(f Filter) (string, error) {
		switch f.kind {
		case kindTrue:
			return "TRUE", nil
		case kindFalse:
			return "FALSE", nil
		case kindAnd, kindOr:
			parts := make([]string, 0, len(f.Children))
			for _, c := range f.Children {
				s, err := build(c)
				if err != nil {
					return "", err
				}
				parts = append(parts, s)
			}
			sep := " AND "
			if f.kind == kindOr {
				sep = " OR "
			}
			return "(" + strings.Join(parts, sep) + ")", nil
		}
		col, ok := column(f.Field)
		if !ok {
			return "", fmt.Errorf("campo %q não permitido em filtros", f.Field)
		}
		placeholders := make([]string, 0, len(f.Values))
		for _, v := range f.Values {
			placeholders = append(placeholders, fmt.Sprintf("$%d", next))
			next++
			args = append(args, v)
		}
		switch f.Op {
		case OpEq:
			return col + " = " + placeholders[0], nil
		case OpNeq:
			return col + " <> " + placeholders[0], nil
		case OpIn:
			return col + " IN (" + strings.Join(placeholders, ", ") + ")", nil
		}
		return "", fmt.Errorf("operador %q desconhecido", f.Op)
	}
	sql, err := build(f)
	if err != nil {
		return "", nil, err
	}
	return sql, args, nil
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Pointer && rv.IsNil()
}

// normalize converte valores comuns (uuid, ponteiros, números, strings) para
// uma forma comparável. O segundo retorno é false para nulo.
func normalize(v any) (string, bool) {
	if isNil(v) {
		return "", false
	}
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return fmt.Sprint(rv.Interface()), true
}
