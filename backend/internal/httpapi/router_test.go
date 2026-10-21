package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

type fakeDB struct{ err error }

func (f fakeDB) Ping(context.Context) error { return f.err }

func newTestRouter(db httpapi.Pinger) http.Handler {
	return httpapi.NewRouter(httpapi.RouterOptions{
		Log:            slog.New(slog.NewJSONHandler(io.Discard, nil)),
		DB:             db,
		AllowedOrigins: []string{"http://localhost:3000"},
		MaxBodyBytes:   1024,
		OpenAPISpec:    []byte("openapi: 3.1.0"),
	})
}

func TestRotasDeInfraestrutura(t *testing.T) {
	tests := []struct {
		nome       string
		db         httpapi.Pinger
		caminho    string
		status     int
		codigoErro string
	}{
		{"health", fakeDB{}, "/health", 200, ""},
		{"ready ok", fakeDB{}, "/ready", 200, ""},
		{"ready sem banco", fakeDB{errors.New("falhou")}, "/ready", 503, "not_ready"},
		{"rota inexistente", fakeDB{}, "/nao-existe", 404, "not_found"},
		{"docs", fakeDB{}, "/docs", 200, ""},
		{"spec", fakeDB{}, "/docs/openapi.yaml", 200, ""},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			rec := httptest.NewRecorder()
			newTestRouter(tt.db).ServeHTTP(rec, httptest.NewRequestWithContext(context.Background(), http.MethodGet, tt.caminho, nil))
			assert.Equal(t, tt.status, rec.Code)
			assert.NotEmpty(t, rec.Header().Get("X-Request-ID"))
			assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"))
			if tt.codigoErro != "" {
				var body struct {
					Error httpapi.APIError `json:"error"`
				}
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, tt.codigoErro, body.Error.Code)
				assert.Equal(t, tt.status, body.Error.Status)
			}
		})
	}
}

func TestErroInternoNaoVazaDetalhes(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/", nil)
	httpapi.Error(rec, req, errors.New("senha do banco: segredo"))
	assert.Equal(t, 500, rec.Code)
	assert.NotContains(t, rec.Body.String(), "segredo")
	assert.Contains(t, rec.Body.String(), "internal_error")
}

func TestDecodeValidacao(t *testing.T) {
	type entrada struct {
		Email string `json:"email" validate:"required,email"`
		Nome  string `json:"nome" validate:"required,min=3"`
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"email":"x","nome":"a"}`))
	var in entrada
	err := httpapi.Decode(req, &in)
	var apiErr *httpapi.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, 422, apiErr.Status)
	assert.Equal(t, "validation_error", apiErr.Code)
	assert.Contains(t, apiErr.Details.(map[string]any)["fields"], httpapi.FieldError{Field: "email", Message: "E-mail inválido."})

	req = httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/", strings.NewReader(`{"extra":1}`))
	require.ErrorAs(t, httpapi.Decode(req, &in), &apiErr)
	assert.Equal(t, 400, apiErr.Status)
}

func TestRequestIDDoClienteESanitizado(t *testing.T) {
	tests := []struct {
		nome, enviado string
		mantido       bool
	}{
		{"simples", "abc-123_X.y", true},
		{"vazio", "", false},
		{"com espaço", "abc 123", false},
		{"com quebra de linha", "abc\r\nSet-Cookie: x=y", false},
		{"com aspas e símbolos", `a"b<script>`, false},
		{"longo demais", strings.Repeat("a", 65), false},
		{"no limite", strings.Repeat("a", 64), true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/health", nil)
			if tt.enviado != "" {
				req.Header["X-Request-Id"] = []string{tt.enviado}
			}
			rec := httptest.NewRecorder()
			newTestRouter(fakeDB{}).ServeHTTP(rec, req)
			got := rec.Header().Get("X-Request-ID")
			assert.NotEmpty(t, got)
			if tt.mantido {
				assert.Equal(t, tt.enviado, got)
			} else {
				assert.NotEqual(t, tt.enviado, got)
				assert.Regexp(t, `^[a-f0-9]{24}$`, got)
			}
		})
	}
}
