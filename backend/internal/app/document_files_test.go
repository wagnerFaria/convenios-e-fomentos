package app_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pdfType  = "application/pdf"
	docxType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
)

func filesPath(docID string, rest ...string) string {
	return "/api/admin/content/documents/" + docID + "/files" + strings.Join(rest, "")
}

// startUpload autoriza o envio e devolve o id do arquivo e a chave do objeto.
func (e *env) startUpload(token, docID, name, ctype string, size int64) (fileID, key string) {
	e.t.Helper()
	r := e.do("POST", filesPath(docID), map[string]any{"filename": name, "content_type": ctype, "size_bytes": size}, token)
	require.Equal(e.t, 201, r.Status, r.Body)
	d := r.data()
	_, leaked := d["file"].(map[string]any)["object_key"]
	require.False(e.t, leaked, "a chave do objeto não sai da API")
	fields := d["upload"].(map[string]any)["fields"].(map[string]any)
	return d["file"].(map[string]any)["id"].(string), fields["key"].(string)
}

func TestArquivosFluxoCompleto(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})

	id, key := e.startUpload(boss, doc, "edital.pdf", pdfType, 1000)
	assert.Equal(t, "documents/"+doc+"/"+id, key, "chave gerada pelo servidor, sem o nome do cliente")

	// Confirmar antes de enviar → 409.
	r := e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, "file_not_uploaded", r.errCode())
	assert.Equal(t, 404, e.do("GET", filesPath(doc, "/", id, "/download"), nil, boss).Status, "pendente não baixa")

	e.store.Put(key, pdfType, append([]byte("%PDF-1.7\n"), make([]byte, 100)...))
	r = e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Equal(t, "confirmado", r.data()["status"])
	assert.EqualValues(t, 109, r.data()["size_bytes"], "o tamanho passa a ser o real")
	assert.Equal(t, 200, e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss).Status, "idempotente")

	r = e.do("GET", filesPath(doc), nil, boss)
	require.Equal(t, 200, r.Status)
	assert.Len(t, r.Body["data"], 1)

	r = e.do("GET", filesPath(doc, "/", id, "/download"), nil, boss)
	require.Equal(t, 200, r.Status, r.Body)
	assert.Contains(t, r.data()["url"], key)
	assert.NotEmpty(t, r.data()["expires_at"])

	assert.Equal(t, 204, e.do("DELETE", filesPath(doc, "/", id), nil, boss).Status)
	assert.False(t, e.store.Has(key), "o objeto sai do armazenamento")
	assert.Equal(t, 404, e.do("GET", filesPath(doc, "/", id, "/download"), nil, boss).Status)
}

func TestArquivosValidacaoDoEnvio(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	const mb = int64(1) << 20
	cases := []struct {
		name  string
		file  string
		ctype string
		size  int64
	}{
		{"tipo não permitido", "a.exe", "application/x-msdownload", 10},
		{"extensão diferente do tipo", "a.docx", pdfType, 10},
		{"sem extensão", "a", pdfType, 10},
		{"pdf acima de 200 MB", "a.pdf", pdfType, 200*mb + 1},
		{"docx acima de 200 MB", "a.docx", docxType, 200*mb + 1},
		{"imagem acima de 25 MB", "a.png", "image/png", 25*mb + 1},
		{"tamanho zero", "a.pdf", pdfType, 0},
		{"nome vazio", "  ", pdfType, 10},
	}
	for _, c := range cases {
		r := e.do("POST", filesPath(doc), map[string]any{"filename": c.file, "content_type": c.ctype, "size_bytes": c.size}, boss)
		assert.Equal(t, 422, r.Status, c.name)
	}
	// Limites exatos são aceitos.
	for _, c := range []struct {
		file, ctype string
		size        int64
	}{{"a.pdf", pdfType, 200 * mb}, {"a.docx", docxType, 200 * mb}, {"a.jpeg", "image/jpeg", 25 * mb}, {"A.PNG", "image/png", 25 * mb}} {
		r := e.do("POST", filesPath(doc), map[string]any{"filename": c.file, "content_type": c.ctype, "size_bytes": c.size}, boss)
		assert.Equal(t, 201, r.Status, c.file)
	}
	// Caminho enviado pelo cliente é reduzido ao nome base.
	r := e.do("POST", filesPath(doc), map[string]any{"filename": `..\..\etc/passwd.pdf`, "content_type": pdfType, "size_bytes": 10}, boss)
	require.Equal(t, 201, r.Status)
	assert.Equal(t, "passwd.pdf", r.data()["file"].(map[string]any)["filename"])
	// Campo desconhecido é recusado.
	assert.Equal(t, 400, e.do("POST", filesPath(doc), map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 1, "object_key": "x"}, boss).Status)
	assert.Equal(t, 404, e.do("POST", filesPath("00000000-0000-0000-0000-000000000000"), map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 1}, boss).Status)
}

func TestArquivosConfirmacaoRecusaObjetoInvalido(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	cases := []struct {
		name  string
		ctype string
		file  string
		data  []byte
		stype string
	}{
		{"conteúdo não é PDF", pdfType, "a.pdf", []byte("MZ\x90\x00 executável"), pdfType},
		{"tipo armazenado diferente", pdfType, "a.pdf", []byte("%PDF-1.4"), "text/html"},
		{"png falso", "image/png", "a.png", []byte("GIF89a......"), "image/png"},
		{"imagem acima de 25 MB", "image/png", "a.png", append([]byte("\x89PNG\r\n\x1a\n"), make([]byte, 25<<20)...), "image/png"},
	}
	for _, c := range cases {
		id, key := e.startUpload(boss, doc, c.file, c.ctype, 100)
		e.store.Put(key, c.stype, c.data)
		r := e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss)
		assert.Equal(t, 422, r.Status, c.name)
		assert.False(t, e.store.Has(key), c.name+": objeto recusado é apagado")
		assert.Equal(t, 404, e.do("GET", filesPath(doc, "/", id, "/download"), nil, boss).Status, c.name+": registro também")
	}
	assert.Equal(t, 0, e.store.Len())
}

func TestArquivosLimitePorDocumento(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	for i := 0; i < 50; i++ {
		require.Equal(t, 201, e.do("POST", filesPath(doc), map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 1}, boss).Status)
	}
	r := e.do("POST", filesPath(doc), map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 1}, boss)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, "too_many_files", r.errCode())
}

func TestArquivosApiPublicaSoDeDocumentoPublicado(t *testing.T) {
	e, boss := setup(t)
	e.runSeed(false)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	id, key := e.startUpload(boss, doc, "edital.pdf", pdfType, 100)
	pendente, _ := e.startUpload(boss, doc, "outro.pdf", pdfType, 100)
	e.store.Put(key, pdfType, []byte("%PDF-1.4"))
	require.Equal(t, 200, e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss).Status)

	pubList := "/api/documents/" + doc + "/files"
	pubDown := func(f string) string { return pubList + "/" + f + "/download" }
	assert.Equal(t, 404, e.do("GET", pubList, nil, "").Status, "rascunho não revela arquivos")
	assert.Equal(t, 404, e.do("GET", pubDown(id), nil, "").Status)

	require.Equal(t, 200, e.do("POST", "/api/admin/content/documents/"+doc+"/publish", nil, boss).Status)
	r := e.do("GET", pubList, nil, "")
	require.Equal(t, 200, r.Status, r.Body)
	list := r.Body["data"].([]any)
	require.Len(t, list, 1, "só arquivos confirmados")
	assert.Equal(t, id, list[0].(map[string]any)["id"])
	assert.NotContains(t, toJSON(r.Body), "documents/", "a chave do objeto não vaza")

	r = e.do("GET", pubDown(id), nil, "")
	require.Equal(t, 200, r.Status, r.Body)
	assert.Contains(t, r.data()["url"], key)
	assert.Equal(t, 404, e.do("GET", pubDown(pendente), nil, "").Status, "pendente não baixa")

	// Arquivo de outro documento não é alcançável por este.
	outro := e.docID(boss, map[string]any{"titulo": "Outro"})
	require.Equal(t, 200, e.do("POST", "/api/admin/content/documents/"+outro+"/publish", nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", "/api/documents/"+outro+"/files/"+id+"/download", nil, "").Status)

	// Despublicar fecha o acesso de novo.
	require.Equal(t, 200, e.do("POST", "/api/admin/content/documents/"+doc+"/unpublish", nil, boss).Status)
	assert.Equal(t, 404, e.do("GET", pubDown(id), nil, "").Status)
}

func TestArquivosPermissoesPorRegistro(t *testing.T) {
	e, boss, login := seeded(t)
	grantLegacyReviewerDocs(e)
	author := login("author@dev.local")
	reviewer := login("revisor@dev.local")

	mine := e.docID(author, map[string]any{"titulo": "Meu"})
	theirs := e.docID(boss, map[string]any{"titulo": "Do chefe"})
	body := map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 10}

	// Author só envia/exclui em documentos que criou (condição is-creator), mas lê todos.
	assert.Equal(t, 201, e.do("POST", filesPath(mine), body, author).Status)
	assert.Equal(t, 403, e.do("POST", filesPath(theirs), body, author).Status)
	assert.Equal(t, 200, e.do("GET", filesPath(theirs), nil, author).Status)

	// Revisor só lê.
	assert.Equal(t, 200, e.do("GET", filesPath(mine), nil, reviewer).Status)
	assert.Equal(t, 403, e.do("POST", filesPath(mine), body, reviewer).Status)

	// Sem token → 401; anônimo não alcança a API administrativa.
	assert.Equal(t, 401, e.do("GET", filesPath(mine), nil, "").Status)
	assert.Equal(t, 401, e.do("POST", filesPath(mine), body, "").Status)
}

func TestArquivosExcluirDocumentoApagaObjetos(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	for i := 0; i < 2; i++ {
		id, key := e.startUpload(boss, doc, "a.pdf", pdfType, 10)
		e.store.Put(key, pdfType, []byte("%PDF-1.4"))
		require.Equal(t, 200, e.do("POST", filesPath(doc, "/", id, "/confirm"), nil, boss).Status)
	}
	require.Equal(t, 2, e.store.Len())
	require.Equal(t, 204, e.do("DELETE", "/api/admin/content/documents/"+doc, nil, boss).Status)
	assert.Equal(t, 0, e.store.Len(), "objetos removidos junto com o documento")
	var n int
	require.NoError(t, e.pool.QueryRow(context.Background(), "SELECT count(*) FROM document_files").Scan(&n))
	assert.Equal(t, 0, n)
}

func TestArquivosLimpezaDeEnviosPendentes(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	oldID, oldKey := e.startUpload(boss, doc, "velho.pdf", pdfType, 10)
	_, newKey := e.startUpload(boss, doc, "novo.pdf", pdfType, 10)
	e.store.Put(oldKey, pdfType, []byte("%PDF-1.4"))
	e.store.Put(newKey, pdfType, []byte("%PDF-1.4"))
	_, err := e.pool.Exec(context.Background(), "UPDATE document_files SET created_at = $2 WHERE id = $1", oldID, time.Now().Add(-25*time.Hour))
	require.NoError(t, err)

	n, err := e.app.Documents.CleanupPendingFiles(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.False(t, e.store.Has(oldKey))
	assert.True(t, e.store.Has(newKey), "envio recente é preservado")
}

func TestArquivosSemArmazenamentoConfigurado(t *testing.T) {
	e, boss := setup(t)
	doc := e.docID(boss, map[string]any{"titulo": "Edital"})
	e.app.Documents.SetStorage(nil, time.Minute, time.Minute, nil)
	r := e.do("POST", filesPath(doc), map[string]any{"filename": "a.pdf", "content_type": pdfType, "size_bytes": 1}, boss)
	assert.Equal(t, 503, r.Status)
	assert.Equal(t, "storage_unavailable", r.errCode())
	// O restante dos documentos segue funcionando.
	assert.Equal(t, 204, e.do("DELETE", "/api/admin/content/documents/"+doc, nil, boss).Status)
}
