package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeRecusaNUL(t *testing.T) {
	type inner struct{ Notes []string }
	type body struct {
		Title string
		Ptr   *string
		Tags  map[string]string
		In    inner
	}
	decode := func(js string) error {
		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(js))
		var b body
		return Decode(r, &b)
	}
	require.NoError(t, decode(`{"Title":"ok","Tags":{"a":"b"},"In":{"Notes":["x"]}}`))
	for name, js := range map[string]string{
		"campo":    `{"Title":"a\u0000b"}`,
		"ponteiro": `{"Ptr":"\u0000"}`,
		"mapa":     `{"Tags":{"k":"v\u0000"}}`,
		"chave":    `{"Tags":{"k\u0000":"v"}}`,
		"lista":    `{"In":{"Notes":["ok","\u0000"]}}`,
	} {
		t.Run(name, func(t *testing.T) {
			var api *APIError
			require.ErrorAs(t, decode(js), &api)
			assert.Equal(t, 422, api.Status)
		})
	}
}

func TestRejectNUL(t *testing.T) {
	h := RejectNUL(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for target, want := range map[string]int{
		"/x?q=abc":     http.StatusNoContent,
		"/x?q=a%00b":   http.StatusBadRequest,
		"/x%00y":       http.StatusBadRequest,
		"/x?q=%25":     http.StatusNoContent,
		"/x?q=a%2500b": http.StatusNoContent, // "%00" escrito como texto (o % foi escapado) não é NUL
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, target, nil))
		assert.Equal(t, want, rec.Code, target)
	}
}
