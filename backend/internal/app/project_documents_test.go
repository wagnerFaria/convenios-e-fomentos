package app_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func docsPath(id string, rest ...string) string {
	return "/api/admin/projects/" + id + "/documents" + strings.Join(rest, "")
}

// checklist devolve os itens do projeto indexados por "código" ou "código@emenda".
func (f *fundingProject) checklist(tok string) (map[string]map[string]any, resp) {
	f.t.Helper()
	r := f.do("GET", docsPath(f.id), nil, tok)
	require.Equal(f.t, 200, r.Status, r.Body)
	out := map[string]map[string]any{}
	for _, it := range r.data()["items"].([]any) {
		m := it.(map[string]any)
		k := m["code"].(string)
		if a, _ := m["amendmentId"].(string); a != "" {
			k += "@" + a
		}
		out[k] = m
	}
	return out, r
}

var pdfBytes = append([]byte("%PDF-1.7\n"), make([]byte, 100)...)

// upload anexa um PDF ao documento (pedido, envio ao armazenamento e confirmação) e devolve o id do arquivo.
func (f *fundingProject) upload(tok, code, amendment string) string {
	f.t.Helper()
	body := map[string]any{"filename": code + ".pdf", "contentType": pdfType, "sizeBytes": len(pdfBytes)}
	if amendment != "" {
		body["amendmentId"] = amendment
	}
	r := f.do("POST", docsPath(f.id, "/", code, "/files"), body, tok)
	require.Equal(f.t, 201, r.Status, r.Body)
	file := r.data()["file"].(map[string]any)
	_, leaked := file["objectKey"]
	require.False(f.t, leaked, "a chave do objeto não sai da API")
	key := r.data()["upload"].(map[string]any)["fields"].(map[string]any)["key"].(string)
	f.store.Put(key, pdfType, pdfBytes)
	c := f.do("POST", docsPath(f.id, "/files/", file["id"].(string), "/confirm"), nil, tok)
	require.Equal(f.t, 200, c.Status, c.Body)
	return file["id"].(string)
}

func TestChecklistPorInstrumentoENatureza(t *testing.T) {
	p := newProjEnv(t)

	// Sem instrumento não há checklist.
	id, _ := p.create(p.maria)
	none := &fundingProject{projEnv: p, id: id}
	items, r := none.checklist(p.maria)
	assert.Empty(t, items)
	assert.EqualValues(t, 0, r.data()["summary"].(map[string]any)["total"])

	// Convênio de prefeitura: checklist de prefeituras (23 itens).
	conv := p.newFundingProject("convenio", "municipal_government", nil)
	items, r = conv.checklist(p.maria)
	assert.Len(t, items, 23)
	for _, c := range []string{"oficio_solicitacao", "requerimento_apoio", "plano_trabalho_sigcon", "comprovante_envio_sigcon", "certidao_habilitacao", "termo_referencia",
		"pesquisa_precos", "declaracao_contrapartida", "quadro_detalhamento_despesa", "loa_publicada", "cartao_cnpj", "documento_dirigente", "diploma_prefeito", "comprovante_endereco_dirigente"} {
		assert.Contains(t, items, c)
	}
	for _, c := range []string{"nao_impedimento", "estatuto_atas", "experiencia_previa", "oficio_deputado", "projeto_basico_obra", "ato_calamidade"} {
		assert.NotContains(t, items, c)
	}
	req := items["requerimento_apoio"]
	assert.Equal(t, "C2", req["source"])
	assert.Equal(t, "Anexo II", req["template"].(map[string]any)["name"])
	assert.Contains(t, req["template"].(map[string]any)["url"], "https://www.sedec.mt.gov.br/documents/d/sedec/")
	assert.Equal(t, "missing", req["status"])
	assert.EqualValues(t, 23, r.data()["summary"].(map[string]any)["missing"])

	// Convênio de entidade privada: o checklist continua o de prefeituras, sem o que é de ente público.
	priv := p.newFundingProject("convenio", "association", nil)
	items, _ = priv.checklist(p.maria)
	assert.Len(t, items, 20, "sem quadro de despesa, LOA e diploma")
	for _, c := range []string{"quadro_detalhamento_despesa", "loa_publicada", "diploma_prefeito"} {
		assert.NotContains(t, items, c)
	}
	assert.Contains(t, items, "cartao_cnpj")

	// Termo de Fomento (OSC): checklist de OSC (24 itens), com os anexos próprios.
	fom := p.newFundingProject("fomento", "association", nil)
	items, _ = fom.checklist(p.maria)
	assert.Len(t, items, 24)
	for _, c := range []string{"nao_impedimento", "execucao_direta", "compatibilidade_institucional", "ausencia_conflito_interesses", "declaracao_despesas_indiretas",
		"declaracao_remuneracao_dirigentes", "termo_ciencia_dirigentes", "experiencia_previa", "estatuto_atas", "comprovacao_domicilio"} {
		assert.Contains(t, items, c)
	}
	for _, c := range []string{"declaracao_contrapartida", "cartao_cnpj", "comprovante_envio_sigcon", "pesquisa_precos"} {
		assert.NotContains(t, items, c)
	}
	assert.Equal(t, "F2", items["requerimento_apoio"]["source"])
	assert.Equal(t, "Anexo IV", items["declaracao_compatibilidade_precos"]["template"].(map[string]any)["name"], "no Fomento o Anexo IV é a declaração de preços")
}

func TestChecklistCondicional(t *testing.T) {
	p := newProjEnv(t)

	// Obra, calamidade e evento artístico acrescentam documentos.
	fp := p.newFundingProject("convenio", "municipal_government", func(b map[string]any) {
		b["objectNature"], b["calamity"], b["hasArtisticHiring"] = "works", true, true
	})
	items, _ := fp.checklist(p.maria)
	assert.Len(t, items, 23+3+1-1, "+3 de obra, +1 ato de calamidade, -1 declaração de contrapartida")
	for _, c := range []string{"projeto_basico_obra", "licencas_ambientais", "propriedade_imovel", "ato_calamidade"} {
		assert.Contains(t, items, c)
	}
	assert.NotContains(t, items, "declaracao_contrapartida", "calamidade dispensa a contrapartida")
	assert.Contains(t, items["projeto_basico_obra"]["name"], "ART")

	// Autorização do Governador só aparece quando marcada nos recursos.
	assert.NotContains(t, items, "autorizacao_governador")
	fp.save("funding", map[string]any{"type": "own", "ownAmountCents": 100_000, "artistFeesCents": 10, "governorWaiver": true})
	items, _ = fp.checklist(p.maria)
	assert.Contains(t, items, "autorizacao_governador")

	// Situações informadas na própria etapa de documentos.
	fom := p.newFundingProject("fomento", "association", nil)
	r := fom.save("document-options", map[string]any{"exclusiveSupplier": true, "publicSpaceUse": true, "reducedDeadline": true})
	assert.Equal(t, map[string]any{"exclusiveSupplier": true, "publicSpaceUse": true, "reducedDeadline": true}, r.data()["documentOptions"])
	items, _ = fom.checklist(p.maria)
	assert.Len(t, items, 24+3)
	for _, c := range []string{"exclusividade", "uso_espaco_publico", "reducao_prazo"} {
		assert.Contains(t, items, c)
	}
	assert.Equal(t, "Anexo XVI", items["reducao_prazo"]["template"].(map[string]any)["name"])
	// Uso de espaço público é só do Termo de Fomento: no Convênio a marcação não acrescenta nada.
	conv := p.newFundingProject("convenio", "municipal_government", nil)
	conv.save("document-options", map[string]any{"publicSpaceUse": true})
	items, _ = conv.checklist(p.maria)
	assert.NotContains(t, items, "uso_espaco_publico")
	assert.Equal(t, 400, conv.try("document-options", map[string]any{"talvez": true}).Status, "campo desconhecido")

	// Chamamento: edital quando realizado; justificativa quando dispensado ou inexigível.
	tender := p.newFundingProject("fomento", "association", nil)
	tender.save("funding", map[string]any{"type": "own", "ownAmountCents": 100_000, "tenderStatus": "held", "tenderNoticeNumber": "1/2026", "tenderApprovedOn": "2026-03-01"})
	items, _ = tender.checklist(p.maria)
	assert.Contains(t, items, "chamamento_edital")
	assert.NotContains(t, items, "chamamento_justificativa")
	tender.save("funding", map[string]any{"type": "own", "ownAmountCents": 100_000, "tenderStatus": "waived", "tenderLegalBasis": "art19_i", "tenderJustification": "x"})
	items, _ = tender.checklist(p.maria)
	assert.Contains(t, items, "chamamento_justificativa")
	assert.NotContains(t, items, "chamamento_edital")
}

func TestChecklistPorEmenda(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	dep := p.parliamentarian("Fulano de Tal", "state")
	year := time.Now().Year()
	r := fp.save("funding", map[string]any{"type": "amendment", "amendments": []any{amendment(dep, "1234", year, 100), amendment(dep, "99", year, 200)}})
	ams := r.data()["funding"].(map[string]any)["amendments"].([]any)
	a1, a2 := ams[0].(map[string]any)["id"].(string), ams[1].(map[string]any)["id"].(string)
	require.NotEmpty(t, a1)
	require.NotEqual(t, a1, a2)

	items, _ := fp.checklist(p.maria)
	assert.Len(t, items, 23+4, "ofício do deputado e da Casa Civil para cada uma das duas emendas")
	for _, a := range []string{a1, a2} {
		for _, c := range []string{"oficio_deputado", "oficio_casa_civil"} {
			assert.Contains(t, items, c+"@"+a)
		}
	}
	assert.Equal(t, fmt.Sprintf("Emenda 1234/%d — Fulano de Tal", year), items["oficio_deputado@"+a1]["amendmentLabel"])
	assert.NotContains(t, items, "oficio_deputado", "exigência por emenda não existe sem emenda")
}

func TestUploadDeDocumentoFluxoCompleto(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)

	r := p.do("POST", docsPath(fp.id, "/requerimento_apoio/files"), map[string]any{"filename": "req.pdf", "contentType": pdfType, "sizeBytes": 1000}, p.maria)
	require.Equal(t, 201, r.Status, r.Body)
	file := r.data()["file"].(map[string]any)
	fileID := file["id"].(string)
	key := r.data()["upload"].(map[string]any)["fields"].(map[string]any)["key"].(string)
	assert.Equal(t, "projects/"+fp.id+"/"+fileID, key, "chave gerada pelo servidor, sem o nome do cliente")
	assert.Equal(t, "pending", file["status"])

	// Pendente não conta como entregue, não baixa e não confirma sem o envio.
	items, _ := fp.checklist(p.maria)
	assert.Equal(t, "missing", items["requerimento_apoio"]["status"])
	assert.Len(t, items["requerimento_apoio"]["files"], 1)
	c := p.do("POST", docsPath(fp.id, "/files/", fileID, "/confirm"), nil, p.maria)
	assert.Equal(t, 409, c.Status)
	assert.Equal(t, "file_not_uploaded", c.errCode())
	assert.Equal(t, 404, p.do("GET", docsPath(fp.id, "/files/", fileID, "/download"), nil, p.maria).Status)

	p.store.Put(key, pdfType, pdfBytes)
	c = p.do("POST", docsPath(fp.id, "/files/", fileID, "/confirm"), nil, p.maria)
	require.Equal(t, 200, c.Status, c.Body)
	assert.Equal(t, "confirmed", c.data()["status"])
	assert.EqualValues(t, len(pdfBytes), c.data()["sizeBytes"], "o tamanho passa a ser o real")
	assert.Equal(t, 200, p.do("POST", docsPath(fp.id, "/files/", fileID, "/confirm"), nil, p.maria).Status, "idempotente")

	items, r = fp.checklist(p.maria)
	assert.Equal(t, "done", items["requerimento_apoio"]["status"])
	assert.EqualValues(t, 1, r.data()["summary"].(map[string]any)["done"])
	assert.EqualValues(t, 22, r.data()["summary"].(map[string]any)["missing"])
	// O projeto reflete o resumo e a pendência agregada.
	proj := p.do("GET", "/api/admin/projects/"+fp.id, nil, p.maria)
	assert.EqualValues(t, 22, proj.data()["documents"].(map[string]any)["missing"])
	assert.Equal(t, "documents_missing", issueFields(proj)["documents.documents"])

	d := p.do("GET", docsPath(fp.id, "/files/", fileID, "/download"), nil, p.maria)
	require.Equal(t, 200, d.Status, d.Body)
	assert.Contains(t, d.data()["url"], key)
	assert.NotEmpty(t, d.data()["expiresAt"])

	assert.Equal(t, 204, p.do("DELETE", docsPath(fp.id, "/files/", fileID), nil, p.maria).Status)
	assert.False(t, p.store.Has(key), "o objeto sai do armazenamento")
	items, _ = fp.checklist(p.maria)
	assert.Equal(t, "missing", items["requerimento_apoio"]["status"])
	var n int
	require.NoError(t, p.pool.QueryRow(context.Background(), "SELECT count(*) FROM project_document_items WHERE project_id=$1", fp.id).Scan(&n))
	assert.Zero(t, n, "o item vazio some")
	assert.Equal(t, 404, p.do("DELETE", docsPath(fp.id, "/files/", fileID), nil, p.maria).Status)
}

func TestUploadValidaOPedidoEOConteudo(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	const mb = int64(1) << 20
	req := func(code string, body map[string]any) resp {
		return p.do("POST", docsPath(fp.id, "/", code, "/files"), body, p.maria)
	}
	good := func() map[string]any {
		return map[string]any{"filename": "a.pdf", "contentType": pdfType, "sizeBytes": 10}
	}

	for name, mutate := range map[string]func(map[string]any){
		"tipo não permitido":         func(b map[string]any) { b["contentType"] = "application/x-msdownload"; b["filename"] = "a.exe" },
		"extensão diferente do tipo": func(b map[string]any) { b["filename"] = "a.docx" },
		"sem extensão":               func(b map[string]any) { b["filename"] = "a" },
		"tamanho zero":               func(b map[string]any) { b["sizeBytes"] = 0 },
		"PDF acima de 200 MB":        func(b map[string]any) { b["sizeBytes"] = 200*mb + 1 },
		"imagem acima de 25 MB": func(b map[string]any) {
			b["filename"], b["contentType"], b["sizeBytes"] = "a.png", "image/png", 25*mb+1
		},
		"nome só com caminho": func(b map[string]any) { b["filename"] = "../" },
	} {
		t.Run(name, func(t *testing.T) {
			b := good()
			mutate(b)
			assert.Equal(t, 422, req("requerimento_apoio", b).Status)
		})
	}
	// Caminho no nome vira só o nome do arquivo.
	b := good()
	b["filename"] = `C:\Users\maria\docs\req.pdf`
	r := req("requerimento_apoio", b)
	require.Equal(t, 201, r.Status, r.Body)
	assert.Equal(t, "req.pdf", r.data()["file"].(map[string]any)["filename"])

	// Exigências que não se aplicam, inexistentes, e escopo de emenda.
	assert.Equal(t, "document_not_applicable", req("nao_impedimento", good()).errCode(), "documento de Fomento num Convênio")
	assert.Equal(t, "document_not_applicable", req("inexistente", good()).errCode())
	bad := good()
	bad["amendmentId"] = "00000000-0000-0000-0000-000000000000"
	assert.Equal(t, 422, req("requerimento_apoio", bad).Status, "documento que não é por emenda recusa amendmentId")

	// Conteúdo: assinatura do formato e tamanho reais são conferidos na confirmação.
	mk := func(content []byte, ctype, name string) (fileID string, key string) {
		up := req("termo_referencia", map[string]any{"filename": name, "contentType": ctype, "sizeBytes": len(content)})
		require.Equal(t, 201, up.Status, up.Body)
		fileID = up.data()["file"].(map[string]any)["id"].(string)
		key = up.data()["upload"].(map[string]any)["fields"].(map[string]any)["key"].(string)
		p.store.Put(key, ctype, content)
		return
	}
	id, key := mk([]byte("MZ não é um PDF"), pdfType, "falso.pdf")
	c := p.do("POST", docsPath(fp.id, "/files/", id, "/confirm"), nil, p.maria)
	assert.Equal(t, 422, c.Status)
	assert.False(t, p.store.Has(key), "arquivo que não passa é apagado")
	assert.Equal(t, 404, p.do("DELETE", docsPath(fp.id, "/files/", id), nil, p.maria).Status, "e o registro também")
	id, key = mk(pdfBytes, "image/png", "foto.png")
	assert.Equal(t, 422, p.do("POST", docsPath(fp.id, "/files/", id, "/confirm"), nil, p.maria).Status, "PDF enviado como PNG")
	assert.False(t, p.store.Has(key))
}

func TestLimiteDeArquivosPorDocumento(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	for range 10 {
		fp.upload(p.maria, "pesquisa_precos", "")
	}
	r := p.do("POST", docsPath(fp.id, "/pesquisa_precos/files"), map[string]any{"filename": "a.pdf", "contentType": pdfType, "sizeBytes": 10}, p.maria)
	assert.Equal(t, 409, r.Status)
	assert.Equal(t, "too_many_files", r.errCode())
	items, _ := fp.checklist(p.maria)
	assert.Len(t, items["pesquisa_precos"]["files"], 10)
}

func TestJustificativaDeNaoEnvio(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	put := func(code, text string) resp {
		return p.do("PUT", docsPath(fp.id, "/", code, "/justification"), map[string]any{"text": text}, p.maria)
	}

	r := put("cartao_cnpj", "  O cartão será emitido na assinatura.  ")
	require.Equal(t, 200, r.Status, r.Body)
	item := func(r resp, code string) map[string]any {
		for _, it := range r.data()["items"].([]any) {
			if m := it.(map[string]any); m["code"] == code {
				return m
			}
		}
		return nil
	}
	assert.Equal(t, "justified", item(r, "cartao_cnpj")["status"])
	assert.Equal(t, "O cartão será emitido na assinatura.", item(r, "cartao_cnpj")["justification"], "texto aparado")
	assert.EqualValues(t, 1, r.data()["summary"].(map[string]any)["justified"])
	assert.EqualValues(t, 22, r.data()["summary"].(map[string]any)["missing"])

	// Com arquivo, o item conta como entregue e a justificativa fica registrada.
	fp.upload(p.maria, "cartao_cnpj", "")
	items, _ := fp.checklist(p.maria)
	assert.Equal(t, "done", items["cartao_cnpj"]["status"])
	assert.NotEmpty(t, items["cartao_cnpj"]["justification"])

	// Esvaziar a justificativa de item sem arquivo apaga o item; limite de tamanho; documento fora do checklist.
	r = put("termo_referencia", "x")
	require.Equal(t, 200, r.Status)
	r = put("termo_referencia", "   ")
	assert.Equal(t, "missing", item(r, "termo_referencia")["status"])
	var n int
	require.NoError(t, p.pool.QueryRow(context.Background(), "SELECT count(*) FROM project_document_items WHERE requirement_code='termo_referencia'").Scan(&n))
	assert.Zero(t, n)
	assert.Equal(t, 422, put("termo_referencia", strings.Repeat("a", 2001)).Status)
	assert.Equal(t, "document_not_applicable", put("nao_impedimento", "x").errCode())
	assert.Equal(t, 404, p.do("PUT", docsPath("00000000-0000-0000-0000-000000000000", "/cartao_cnpj/justification"), map[string]any{"text": "x"}, p.maria).Status)

	// O projeto reflete o resumo: 23 itens, 1 entregue (cartão do CNPJ, com arquivo e justificativa).
	proj := p.do("GET", "/api/admin/projects/"+fp.id, nil, p.maria)
	assert.EqualValues(t, 22, proj.data()["documents"].(map[string]any)["missing"])
	assert.EqualValues(t, 1, proj.data()["documents"].(map[string]any)["done"])
}

func TestAnexosDeEmendaSobrevivemAEdicaoDoRecurso(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	dep := p.parliamentarian("Fulano", "state")
	year := time.Now().Year()
	save := func(rows ...any) resp {
		return fp.save("funding", map[string]any{"type": "amendment", "amendments": rows})
	}
	r := save(amendment(dep, "1", year, 100), amendment(dep, "2", year, 200))
	ams := r.data()["funding"].(map[string]any)["amendments"].([]any)
	a1, a2 := ams[0].(map[string]any)["id"].(string), ams[1].(map[string]any)["id"].(string)

	f1 := fp.upload(p.maria, "oficio_deputado", a1)
	f2 := fp.upload(p.maria, "oficio_deputado", a2)
	_ = f1
	keys := func() int { return p.store.Len() }
	require.Equal(t, 2, keys())

	// Reenviar com os ids mantém as emendas (e os anexos), mesmo alterando os valores.
	e1, e2 := amendment(dep, "1", year, 150), amendment(dep, "2", year, 250)
	e1["id"], e2["id"] = a1, a2
	r = save(e1, e2)
	ams = r.data()["funding"].(map[string]any)["amendments"].([]any)
	assert.Equal(t, a1, ams[0].(map[string]any)["id"], "ids estáveis")
	assert.Equal(t, a2, ams[1].(map[string]any)["id"])
	items, _ := fp.checklist(p.maria)
	assert.Equal(t, "done", items["oficio_deputado@"+a1]["status"])
	assert.Equal(t, "done", items["oficio_deputado@"+a2]["status"])
	assert.Equal(t, 2, keys())

	// Remover a segunda emenda apaga também o ofício dela (registro e objeto).
	r = save(e1)
	require.Equal(t, 200, r.Status)
	items, _ = fp.checklist(p.maria)
	assert.Equal(t, "done", items["oficio_deputado@"+a1]["status"])
	assert.NotContains(t, items, "oficio_deputado@"+a2)
	assert.Equal(t, 1, keys(), "o objeto da emenda removida sai do armazenamento")
	assert.Equal(t, 404, p.do("DELETE", docsPath(fp.id, "/files/", f2), nil, p.maria).Status)

	// Sem informar o id a emenda é nova: a antiga (e o anexo) saem.
	r = save(amendment(dep, "1", year, 150))
	newID := r.data()["funding"].(map[string]any)["amendments"].([]any)[0].(map[string]any)["id"].(string)
	assert.NotEqual(t, a1, newID)
	assert.Zero(t, keys())
	items, _ = fp.checklist(p.maria)
	assert.Equal(t, "missing", items["oficio_deputado@"+newID]["status"])

	// Um id de emenda de outro projeto não vale: vira emenda nova.
	other := p.newFundingProject("convenio", "municipal_government", nil)
	ro := other.save("funding", map[string]any{"type": "amendment", "amendments": []any{amendment(dep, "7", year, 10)}})
	foreign := ro.data()["funding"].(map[string]any)["amendments"].([]any)[0].(map[string]any)["id"].(string)
	e := amendment(dep, "1", year, 150)
	e["id"] = foreign
	r = save(e)
	assert.NotEqual(t, foreign, r.data()["funding"].(map[string]any)["amendments"].([]any)[0].(map[string]any)["id"])
	oi, _ := other.checklist(p.maria)
	assert.Contains(t, oi, "oficio_deputado@"+foreign, "o outro projeto não foi tocado")

	// Anexar a uma emenda inexistente ou sem informá-la.
	body := map[string]any{"filename": "a.pdf", "contentType": pdfType, "sizeBytes": 10}
	assert.Equal(t, 422, p.do("POST", docsPath(fp.id, "/oficio_deputado/files"), body, p.maria).Status, "falta amendmentId")
	body["amendmentId"] = foreign
	assert.Equal(t, 422, p.do("POST", docsPath(fp.id, "/oficio_deputado/files"), body, p.maria).Status, "emenda de outro projeto")
}

func TestExcluirProjetoApagaOsArquivos(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	fp.upload(p.maria, "requerimento_apoio", "")
	fp.upload(p.maria, "termo_referencia", "")
	require.Equal(t, 2, p.store.Len())
	require.Equal(t, 204, p.do("DELETE", "/api/admin/projects/"+fp.id, nil, p.maria).Status)
	assert.Zero(t, p.store.Len())
	var n int
	require.NoError(t, p.pool.QueryRow(context.Background(), "SELECT count(*) FROM project_files").Scan(&n))
	assert.Zero(t, n)
}

func TestDocumentosSoDoTitular(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	id := fp.upload(p.maria, "requerimento_apoio", "")
	body := map[string]any{"filename": "a.pdf", "contentType": pdfType, "sizeBytes": 10}

	// Outra pessoa não vê o projeto (404 nas leituras) nem altera (403 nas escritas).
	assert.Equal(t, 404, p.do("GET", docsPath(fp.id), nil, p.joao).Status)
	assert.Equal(t, 404, p.do("GET", docsPath(fp.id, "/files/", id, "/download"), nil, p.joao).Status)
	assert.Equal(t, 403, p.do("POST", docsPath(fp.id, "/requerimento_apoio/files"), body, p.joao).Status)
	assert.Equal(t, 403, p.do("POST", docsPath(fp.id, "/files/", id, "/confirm"), nil, p.joao).Status)
	assert.Equal(t, 403, p.do("DELETE", docsPath(fp.id, "/files/", id), nil, p.joao).Status)
	assert.Equal(t, 403, p.do("PUT", docsPath(fp.id, "/cartao_cnpj/justification"), map[string]any{"text": "x"}, p.joao).Status)
	assert.Equal(t, 403, p.do("PUT", "/api/admin/projects/"+fp.id+"/document-options", map[string]any{"version": 1}, p.joao).Status)
	assert.Equal(t, 200, p.do("GET", docsPath(fp.id, "/files/", id, "/download"), nil, p.maria).Status)

	// Quem lê todos os projetos baixa, mas não altera.
	p.mkRole("leitor", "admin::projects.read")
	staff, _ := p.newAdmin(p.boss, "bia@exemplo.com", "leitor")
	assert.Equal(t, 200, p.do("GET", docsPath(fp.id), nil, staff).Status)
	assert.Equal(t, 200, p.do("GET", docsPath(fp.id, "/files/", id, "/download"), nil, staff).Status)
	assert.Equal(t, 403, p.do("DELETE", docsPath(fp.id, "/files/", id), nil, staff).Status)
}

func TestChecklistCompletoZeraAPendenciaDeDocumentos(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("fomento", "association", nil)
	items, _ := fp.checklist(p.maria)
	require.Len(t, items, 24)
	i := 0
	for code := range items {
		if i%2 == 0 {
			fp.upload(p.maria, code, "")
		} else {
			require.Equal(t, 200, p.do("PUT", docsPath(fp.id, "/", code, "/justification"), map[string]any{"text": "Será entregue na assinatura."}, p.maria).Status)
		}
		i++
	}
	_, r := fp.checklist(p.maria)
	sum := r.data()["summary"].(map[string]any)
	assert.EqualValues(t, 24, sum["total"])
	assert.EqualValues(t, 0, sum["missing"])
	proj := p.do("GET", "/api/admin/projects/"+fp.id, nil, p.maria)
	assert.NotContains(t, issueFields(proj), "documents.documents")
}

// Todo item da carga inicial é obrigatório: o resumo dos obrigatórios acompanha o geral.
func TestChecklistMarcaTodosComoObrigatoriosPorPadrao(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("fomento", "association", nil)
	items, r := fp.checklist(p.maria)
	for code, it := range items {
		assert.Equal(t, true, it["mandatory"], code)
	}
	sum := r.data()["summary"].(map[string]any)
	assert.EqualValues(t, 24, sum["requiredTotal"])
	assert.EqualValues(t, 24, sum["requiredMissing"])
}

// Documento opcional aparece marcado, mas não conta nos obrigatórios nem gera a pendência que bloqueia a continuidade.
func TestDocumentoOpcionalNaoBloqueia(t *testing.T) {
	p := newProjEnv(t)
	_, err := p.pool.Exec(context.Background(), "UPDATE document_requirements SET mandatory = false WHERE code = 'termo_referencia'")
	require.NoError(t, err)
	fp := p.newFundingProject("fomento", "association", nil)
	items, r := fp.checklist(p.maria)
	require.Len(t, items, 24)
	assert.Equal(t, false, items["termo_referencia"]["mandatory"])
	sum := r.data()["summary"].(map[string]any)
	assert.EqualValues(t, 24, sum["total"])
	assert.EqualValues(t, 23, sum["requiredTotal"])
	assert.EqualValues(t, 24, sum["missing"])
	assert.EqualValues(t, 23, sum["requiredMissing"])

	for code, it := range items {
		if it["mandatory"] == true {
			fp.upload(p.maria, code, "")
		}
	}
	_, r = fp.checklist(p.maria)
	sum = r.data()["summary"].(map[string]any)
	assert.EqualValues(t, 0, sum["requiredMissing"])
	assert.EqualValues(t, 1, sum["missing"], "o opcional segue pendente")
	proj := p.do("GET", "/api/admin/projects/"+fp.id, nil, p.maria)
	assert.NotContains(t, issueFields(proj), "documents.documents")
	assert.EqualValues(t, 0, proj.data()["documents"].(map[string]any)["requiredMissing"])
}

func TestLimpezaDeEnviosNaoConfirmados(t *testing.T) {
	p := newProjEnv(t)
	fp := p.newFundingProject("convenio", "municipal_government", nil)
	r := p.do("POST", docsPath(fp.id, "/requerimento_apoio/files"), map[string]any{"filename": "a.pdf", "contentType": pdfType, "sizeBytes": 10}, p.maria)
	require.Equal(t, 201, r.Status)
	key := r.data()["upload"].(map[string]any)["fields"].(map[string]any)["key"].(string)
	p.store.Put(key, pdfType, pdfBytes)
	fresh := fp.upload(p.maria, "termo_referencia", "")
	_ = fresh

	n, err := p.app.Projects.CleanupPendingFiles(context.Background())
	require.NoError(t, err)
	assert.Zero(t, n, "envios recentes não são removidos")
	_, err = p.pool.Exec(context.Background(), "UPDATE project_files SET created_at = now() - interval '25 hours' WHERE status = 'pending'")
	require.NoError(t, err)
	n, err = p.app.Projects.CleanupPendingFiles(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, n)
	assert.False(t, p.store.Has(key))
	items, _ := fp.checklist(p.maria)
	assert.Equal(t, "done", items["termo_referencia"]["status"], "o confirmado fica")
}
