package document

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"slices"
	"sort"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// decodePatch lê o corpo JSON como objeto. Chaves fora de known → 422 (campo desconhecido
// ou não permitido nesta rota). Para a API administrativa, known é o conjunto de campos
// do Schema: campos de sistema são aceitos aqui e recusados depois por RejectedWrite,
// com a mensagem de permissão.
func decodePatch(r *http.Request, known func(key string) bool) (Patch, error) {
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return Patch{}, &httpapi.APIError{Status: 413, Code: "payload_too_large", Message: "O corpo da requisição excede o limite permitido."}
		}
		return Patch{}, httpapi.ErrBadRequest("Corpo da requisição inválido.")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil || fields == nil {
		return Patch{}, httpapi.ErrBadRequest("Corpo da requisição inválido: esperado um objeto JSON.")
	}
	var p Patch
	var issues []httpapi.FieldError
	bad := func(k, m string) { issues = append(issues, httpapi.FieldError{Field: k, Message: m}) }
	for k, v := range fields {
		if !known(k) {
			bad(k, "Campo desconhecido ou não permitido nesta operação.")
			continue
		}
		p.Keys = append(p.Keys, k)
		var err error
		switch k {
		case "titulo":
			p.Titulo, err = asString(v)
		case "conteudo":
			p.Conteudo, err = asString(v)
		case "status":
			p.Status, err = asString(v)
		case "observacao_interna":
			p.ObservacaoInterna, err = asString(v)
		case "department_id":
			if string(v) != "null" {
				var id uuid.UUID
				if err = json.Unmarshal(v, &id); err == nil {
					p.DepartmentID = &id
				}
			}
		}
		if err != nil {
			bad(k, "Tipo de valor inválido.")
		}
	}
	if len(issues) > 0 {
		sort.Slice(issues, func(i, j int) bool { return issues[i].Field < issues[j].Field })
		return Patch{}, httpapi.ErrValidation(map[string]any{"fields": issues})
	}
	slices.Sort(p.Keys)
	return p, nil
}

// asString decodifica uma string JSON; null é tratado como tipo inválido (os campos de texto não são anuláveis).
func asString(raw json.RawMessage) (*string, error) {
	var s string
	if string(raw) == "null" {
		return nil, errors.New("null")
	}
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
