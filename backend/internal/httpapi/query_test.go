package httpapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

func req(query string) *http.Request {
	return httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/x?"+query, nil)
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		nome, query string
		want        httpapi.Page
		erro        bool
	}{
		{"padrões", "", httpapi.Page{Page: 1, PageSize: 25}, false},
		{"informado", "page=3&pageSize=10", httpapi.Page{Page: 3, PageSize: 10}, false},
		{"limite máximo", "pageSize=999", httpapi.Page{Page: 1, PageSize: 100}, false},
		{"page zero", "page=0", httpapi.Page{}, true},
		{"page texto", "page=a", httpapi.Page{}, true},
		{"pageSize negativo", "pageSize=-1", httpapi.Page{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			p, err := httpapi.ParsePage(req(tt.query), 25, 100)
			if tt.erro {
				var apiErr *httpapi.APIError
				require.ErrorAs(t, err, &apiErr)
				assert.Equal(t, 422, apiErr.Status)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, p)
		})
	}
	p := httpapi.Page{Page: 3, PageSize: 10}
	assert.Equal(t, 20, p.Offset())
	assert.Equal(t, &httpapi.Meta{Page: 3, PageSize: 10, Total: 25, PageCount: 3}, p.Meta(25))
	assert.Equal(t, 0, p.Meta(0).PageCount)
}

func TestParseSort(t *testing.T) {
	allowed := map[string]string{"email": "a.email", "createdAt": "a.created_at"}
	def := httpapi.Sort{Column: "a.id"}
	tests := []struct {
		nome, query string
		want        string
		erro        bool
	}{
		{"padrão", "", "a.id ASC", false},
		{"asc implícito", "sort=email", "a.email ASC", false},
		{"desc", "sort=createdAt:desc", "a.created_at DESC", false},
		{"direção em maiúsculas", "sort=email:DESC", "a.email DESC", false},
		{"campo fora da lista", "sort=password_hash", "", true},
		{"injeção", "sort=email%3BDROP+TABLE+x", "", true},
		{"direção inválida", "sort=email:up", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			s, err := httpapi.ParseSort(req(tt.query), allowed, def)
			if tt.erro {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, s.SQL())
		})
	}
}
