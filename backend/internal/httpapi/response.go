// Pacote httpapi contém o roteador, middlewares e helpers de resposta/erro.
package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/wagner/convenios-e-fomentos/backend/internal/logging"
)

// Meta carrega informações de paginação.
type Meta struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	Total     int `json:"total"`
	PageCount int `json:"pageCount"`
}

type successBody struct {
	Data any   `json:"data"`
	Meta *Meta `json:"meta,omitempty"`
}

// APIError é o formato padronizado de erro da API.
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string { return e.Code + ": " + e.Message }

type errorBody struct {
	Error *APIError `json:"error"`
}

// Construtores dos erros padronizados.
func ErrBadRequest(msg string) *APIError {
	return &APIError{Status: 400, Code: "bad_request", Message: msg}
}
func ErrUnauthorized() *APIError {
	return &APIError{Status: 401, Code: "unauthorized", Message: "Autenticação necessária ou credenciais inválidas."}
}
func ErrForbidden() *APIError {
	return &APIError{Status: 403, Code: "forbidden", Message: "Você não tem permissão para executar esta ação."}
}
func ErrNotFound() *APIError {
	return &APIError{Status: 404, Code: "not_found", Message: "Recurso não encontrado."}
}
func ErrConflict(code, msg string) *APIError {
	return &APIError{Status: 409, Code: code, Message: msg}
}
func ErrValidation(details any) *APIError {
	return &APIError{Status: 422, Code: "validation_error", Message: "Os dados enviados são inválidos.", Details: details}
}
func ErrTooManyRequests() *APIError {
	return &APIError{Status: 429, Code: "rate_limited", Message: "Muitas requisições. Tente novamente em instantes."}
}

// JSON escreve o envelope de sucesso { "data": ..., "meta": ... }.
func JSON(w http.ResponseWriter, status int, data any, meta *Meta) {
	writeJSON(w, status, successBody{Data: data, Meta: meta})
}

// Error escreve o envelope de erro. Erros que não são *APIError viram 500
// genérico: o detalhe vai apenas para o log, junto do request ID.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		id := logging.RequestID(r.Context())
		slog.ErrorContext(r.Context(), "erro interno", "error", err, "request_id", id)
		apiErr = &APIError{
			Status:  500,
			Code:    "internal_error",
			Message: "Erro interno do servidor.",
			Details: map[string]string{"requestId": id},
		}
	}
	writeJSON(w, apiErr.Status, errorBody{Error: apiErr})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
