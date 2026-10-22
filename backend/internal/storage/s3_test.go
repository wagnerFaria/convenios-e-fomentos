package storage_test

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/wagner/convenios-e-fomentos/backend/internal/storage"
)

// minioImage: as imagens minio/minio deixaram de ser publicadas; esta é a da Chainguard.
const minioImage = "chainguard/minio:latest"

// Sobe um MinIO real atrás de um proxy que imita o HAProxy: o navegador usa /storage/... e o proxy
// remove o prefixo mantendo o Host. É o desenho de produção: a assinatura SigV4 cobre Host e caminho.
func TestS3PreassinadoPeloMesmoDominio(t *testing.T) {
	if testing.Short() {
		t.Skip("teste de integração (ignorado com -short)")
	}
	ctx := context.Background()
	ctr, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{Started: true, ContainerRequest: testcontainers.ContainerRequest{
		Image: minioImage, Cmd: []string{"server", "/data"}, ExposedPorts: []string{"9000/tcp"},
		Env:        map[string]string{"MINIO_ROOT_USER": "teste", "MINIO_ROOT_PASSWORD": "teste-secret-123"},
		WaitingFor: wait.ForHTTP("/minio/health/live").WithPort("9000/tcp"),
	}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = ctr.Terminate(ctx) })
	endpoint, err := ctr.PortEndpoint(ctx, "9000/tcp", "")
	require.NoError(t, err)

	admin, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4("teste", "teste-secret-123", "")})
	require.NoError(t, err)
	require.NoError(t, admin.MakeBucket(ctx, "documentos", minio.MakeBucketOptions{}))

	backend := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
		r.Out.URL.Scheme, r.Out.URL.Host = "http", endpoint
		r.Out.URL.Path = strings.TrimPrefix(r.In.URL.Path, "/storage")
		r.Out.URL.RawPath = ""
		r.Out.Host = r.In.Host // o Host original precisa chegar ao MinIO
	}}
	proxy := httptest.NewServer(backend)
	t.Cleanup(proxy.Close)

	s, err := storage.NewS3(storage.S3Config{Endpoint: endpoint, PublicURL: proxy.URL + "/storage",
		AccessKey: "teste", SecretKey: "teste-secret-123", Bucket: "documentos", Region: "us-east-1"})
	require.NoError(t, err)

	post := func(up storage.Upload, ctype string, body []byte) int {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		for k, v := range up.Fields {
			require.NoError(t, mw.WriteField(k, v))
		}
		fw, err := mw.CreateFormFile("file", "x")
		require.NoError(t, err)
		_, _ = fw.Write(body)
		require.NoError(t, mw.Close())
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, up.URL, &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		res, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		_, _ = io.Copy(io.Discard, res.Body)
		return res.StatusCode
	}

	key := "documents/d1/f1"
	up, err := s.PresignUpload(ctx, key, "application/pdf", 64, time.Minute)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(up.URL, proxy.URL+"/storage/documentos"), up.URL)

	t.Run("recusa tamanho acima do declarado", func(t *testing.T) {
		assert.GreaterOrEqual(t, post(up, "application/pdf", bytes.Repeat([]byte("a"), 65)), 400)
		_, err := s.Stat(ctx, key)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
	t.Run("envia o arquivo", func(t *testing.T) {
		require.Equal(t, 204, post(up, "application/pdf", []byte("%PDF-1.7 conteudo")))
		info, err := s.Stat(ctx, key)
		require.NoError(t, err)
		assert.EqualValues(t, 17, info.Size)
		assert.Equal(t, "application/pdf", info.ContentType)
		head, err := s.Head(ctx, key, 5)
		require.NoError(t, err)
		assert.Equal(t, "%PDF-", string(head))
	})
	t.Run("recusa outro tipo de conteúdo", func(t *testing.T) {
		other, err := s.PresignUpload(ctx, "documents/d1/f2", "application/pdf", 64, time.Minute)
		require.NoError(t, err)
		other.Fields["Content-Type"] = "text/html"
		assert.GreaterOrEqual(t, post(other, "text/html", []byte("<script>")), 400)
	})
	t.Run("recusa trocar a chave do objeto", func(t *testing.T) {
		other, err := s.PresignUpload(ctx, "documents/d1/f3", "application/pdf", 64, time.Minute)
		require.NoError(t, err)
		other.Fields["key"] = "documents/d1/f1"
		assert.GreaterOrEqual(t, post(other, "application/pdf", []byte("%PDF-")), 400)
	})
	t.Run("download pré-assinado como anexo", func(t *testing.T) {
		u, exp, err := s.PresignDownload(ctx, key, "Edital nº 1.pdf", "application/pdf", time.Minute)
		require.NoError(t, err)
		assert.True(t, exp.After(time.Now()))
		res, err := http.Get(u) //nolint:gosec,noctx // URL do próprio teste
		require.NoError(t, err)
		defer func() { _ = res.Body.Close() }()
		body, _ := io.ReadAll(res.Body)
		require.Equal(t, 200, res.StatusCode, string(body))
		assert.Equal(t, "%PDF-1.7 conteudo", string(body))
		assert.Contains(t, res.Header.Get("Content-Disposition"), "attachment")
		assert.Contains(t, res.Header.Get("Content-Disposition"), "filename*=utf-8''Edital%20n%C2%BA%201.pdf", "nome com acento preservado")
		assert.Equal(t, "application/pdf", res.Header.Get("Content-Type"))

		// Sem a assinatura, ou com o objeto alterado, não há acesso.
		bare, _ := url.Parse(u)
		bare.RawQuery = ""
		res2, err := http.Get(bare.String()) //nolint:gosec,noctx // URL do próprio teste
		require.NoError(t, err)
		_ = res2.Body.Close()
		assert.Equal(t, 403, res2.StatusCode)
	})
	t.Run("remove o objeto", func(t *testing.T) {
		require.NoError(t, s.Delete(ctx, key))
		require.NoError(t, s.Delete(ctx, key), "remover de novo não é erro")
		_, err := s.Stat(ctx, key)
		assert.ErrorIs(t, err, storage.ErrNotFound)
	})
}
