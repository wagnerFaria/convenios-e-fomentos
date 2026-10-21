package httpapi

import (
	"net/http"
	"strconv"
	"strings"
)

// Page é a paginação pedida (page e pageSize).
type Page struct{ Page, PageSize int }

// Offset devolve o deslocamento para a consulta.
func (p Page) Offset() int { return (p.Page - 1) * p.PageSize }

// Meta monta os metadados de paginação para o total informado.
func (p Page) Meta(total int) *Meta {
	count := (total + p.PageSize - 1) / p.PageSize
	return &Meta{Page: p.Page, PageSize: p.PageSize, Total: total, PageCount: count}
}

func badQuery(field, msg string) error {
	return ErrValidation(map[string]any{"fields": []FieldError{{Field: field, Message: msg}}})
}

// ParsePage lê page (padrão 1) e pageSize (padrão def). Valores inválidos
// geram 422; pageSize acima de max é reduzido a max.
func ParsePage(r *http.Request, def, max int) (Page, error) {
	p := Page{Page: 1, PageSize: def}
	q := r.URL.Query()
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return p, badQuery("page", "Deve ser um inteiro maior ou igual a 1.")
		}
		p.Page = n
	}
	if v := q.Get("pageSize"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return p, badQuery("pageSize", "Deve ser um inteiro maior ou igual a 1.")
		}
		p.PageSize = min(n, max)
	}
	return p, nil
}

// Sort é uma ordenação já resolvida para uma coluna da lista permitida.
type Sort struct {
	Column string
	Desc   bool
}

// SQL devolve "coluna ASC|DESC". A coluna só pode vir da lista permitida.
func (s Sort) SQL() string {
	if s.Desc {
		return s.Column + " DESC"
	}
	return s.Column + " ASC"
}

// ParseSort lê "sort=campo:asc|desc". allowed mapeia campo da API → coluna;
// qualquer campo fora dele gera 422 (evita injeção em ORDER BY).
func ParseSort(r *http.Request, allowed map[string]string, def Sort) (Sort, error) {
	v := r.URL.Query().Get("sort")
	if v == "" {
		return def, nil
	}
	field, dir, _ := strings.Cut(v, ":")
	col, ok := allowed[field]
	if !ok {
		return def, badQuery("sort", "Campo de ordenação não permitido.")
	}
	switch strings.ToLower(dir) {
	case "", "asc":
		return Sort{Column: col}, nil
	case "desc":
		return Sort{Column: col, Desc: true}, nil
	}
	return def, badQuery("sort", "A direção deve ser asc ou desc.")
}
