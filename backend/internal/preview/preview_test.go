package preview

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGotenbergConverte(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/forms/libreoffice/convert", r.URL.Path)
		require.NoError(t, r.ParseMultipartForm(1<<20))
		f, hdr, err := r.FormFile("files")
		require.NoError(t, err)
		body, _ := io.ReadAll(f)
		assert.Equal(t, "docx-fake", string(body))
		assert.True(t, strings.HasSuffix(hdr.Filename, ".docx"))
		_, _ = w.Write([]byte("%PDF-1.7 convertido"))
	}))
	defer srv.Close()

	pdf, err := NewGotenberg(srv.URL, time.Second).ToPDF(context.Background(), "Termo.docx", strings.NewReader("docx-fake"))
	require.NoError(t, err)
	assert.Equal(t, "%PDF-1.7 convertido", string(pdf))
}

func TestGotenbergFalhas(t *testing.T) {
	cases := map[string]http.HandlerFunc{
		"status de erro": func(w http.ResponseWriter, _ *http.Request) { http.Error(w, "boom", http.StatusInternalServerError) },
		"não é PDF":      func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("<html>")) },
	}
	for name, h := range cases {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(h)
			defer srv.Close()
			_, err := NewGotenberg(srv.URL, time.Second).ToPDF(context.Background(), "a.docx", strings.NewReader("x"))
			require.ErrorIs(t, err, ErrUnavailable)
		})
	}
	t.Run("serviço fora do ar", func(t *testing.T) {
		_, err := NewGotenberg("http://127.0.0.1:1", time.Second).ToPDF(context.Background(), "a.docx", strings.NewReader("x"))
		require.ErrorIs(t, err, ErrUnavailable)
	})
}
