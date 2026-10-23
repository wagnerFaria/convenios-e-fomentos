package catalog

import (
	"context"
	_ "embed"
	"encoding/csv"
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// municipiosCSV é a lista de municípios do IBGE (código;nome;UF), baixada de
// servicodados.ibge.gov.br/api/v1/localidades/municipios.
//
//go:embed data/municipios.csv
var municipiosCSV string

// UFs são as 27 unidades da federação.
var UFs = []string{"AC", "AL", "AP", "AM", "BA", "CE", "DF", "ES", "GO", "MA", "MT", "MS", "MG", "PA", "PB", "PR", "PE", "PI", "RJ", "RN", "RS", "RO", "RR", "SC", "SP", "SE", "TO"}

// Normalize devolve o texto em minúsculas, sem acentos e com espaços simples: a forma usada na busca.
func Normalize(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, err := transform.String(t, s)
	if err != nil {
		out = s
	}
	return strings.Join(strings.Fields(strings.ToLower(out)), " ")
}

// Municipality é um município do IBGE.
type Municipality struct {
	IBGECode string `json:"ibgeCode"`
	Name     string `json:"name"`
	UF       string `json:"uf"`
}

type municipalityRow struct{ code, name, uf string }

func parseMunicipalities() ([]municipalityRow, error) {
	r := csv.NewReader(strings.NewReader(municipiosCSV))
	r.Comma = ';'
	recs, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("lendo municípios embutidos: %w", err)
	}
	out := make([]municipalityRow, 0, len(recs))
	for _, rec := range recs {
		if len(rec) != 3 {
			return nil, fmt.Errorf("linha de município inválida: %v", rec)
		}
		out = append(out, municipalityRow{code: rec[0], name: rec[1], uf: rec[2]})
	}
	return out, nil
}

// ensureMunicipalities carrega os municípios quando faltam. Como a lista é de referência (somente
// leitura), a carga é completa e idempotente.
func (s *Service) ensureMunicipalities(ctx context.Context) error {
	rows, err := parseMunicipalities()
	if err != nil {
		return err
	}
	have, err := s.q.CountMunicipalities(ctx)
	if err != nil {
		return err
	}
	if have >= int64(len(rows)) {
		return nil
	}
	p := db.InsertMunicipalitiesParams{}
	for _, m := range rows {
		p.Codes = append(p.Codes, m.code)
		p.Names = append(p.Names, m.name)
		p.Ufs = append(p.Ufs, m.uf)
		p.SearchNames = append(p.SearchNames, Normalize(m.name))
	}
	return s.q.InsertMunicipalities(ctx, p)
}

func fieldError(field, msg string) error {
	return httpapi.ErrValidation(map[string]any{"fields": []httpapi.FieldError{{Field: field, Message: msg}}})
}

// ListMunicipalities busca municípios por UF e por prefixo do nome (sem acentos nem maiúsculas).
func (s *Service) ListMunicipalities(ctx context.Context, uf, search string, page httpapi.Page) ([]Municipality, int, error) {
	var ufp, sp *string
	if uf != "" {
		uf = strings.ToUpper(uf)
		if !isUF(uf) {
			return nil, 0, fieldError("uf", "Informe a sigla de uma unidade da federação.")
		}
		ufp = &uf
	}
	if n := Normalize(search); n != "" {
		if len([]rune(n)) > 60 {
			return nil, 0, fieldError("q", "Use até 60 caracteres.")
		}
		sp = &n
	}
	total, err := s.q.CountMunicipalitiesFiltered(ctx, db.CountMunicipalitiesFilteredParams{Uf: ufp, Search: sp})
	if err != nil {
		return nil, 0, err
	}
	rows, err := s.q.ListMunicipalities(ctx, db.ListMunicipalitiesParams{Uf: ufp, Search: sp, Limit: int32(page.PageSize), Offset: int32(page.Offset())}) //nolint:gosec // limitado por ParsePage
	if err != nil {
		return nil, 0, err
	}
	out := make([]Municipality, len(rows))
	for i, r := range rows {
		out[i] = Municipality{IBGECode: r.IbgeCode, Name: r.Name, UF: r.Uf}
	}
	return out, int(total), nil
}

// GetMunicipality devolve um município pelo código IBGE.
func (s *Service) GetMunicipality(ctx context.Context, code string) (*Municipality, error) {
	r, err := s.q.GetMunicipality(ctx, code)
	if err != nil {
		return nil, notFound(err)
	}
	return &Municipality{IBGECode: r.IbgeCode, Name: r.Name, UF: r.Uf}, nil
}

func isUF(s string) bool {
	for _, u := range UFs {
		if u == s {
			return true
		}
	}
	return false
}
