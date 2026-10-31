package project

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func confirmedFile(name string) DocFile {
	return DocFile{ID: uuid.New(), Filename: name, Status: FileConfirmed}
}

func item(code string, status string, mandatory bool, files ...DocFile) ChecklistItem {
	return ChecklistItem{Code: code, Name: "Doc " + code, Status: status, Mandatory: mandatory, Files: files}
}

func keysOf(ts []reviewTarget) []string {
	var out []string
	for _, t := range ts {
		out = append(out, t.key())
	}
	return out
}

func TestItensConferiveis(t *testing.T) {
	extra := uuid.New()
	pending := DocFile{ID: uuid.New(), Filename: "x.pdf", Status: "pending"}
	cl := &Checklist{
		Items: []ChecklistItem{
			item("a", ItemDone, true, confirmedFile("a.pdf"), pending),
			item("b", ItemJustified, true),
			item("c", ItemMissing, true),  // obrigatório vazio: impedimento
			item("d", ItemMissing, false), // opcional vazio: fora
			{Code: "e", Name: "Doc e", AmendmentID: "em1", AmendmentLabel: "Emenda 1/2026", Status: ItemDone, Mandatory: true, Files: []DocFile{confirmedFile("e.pdf")}},
		},
		ExtraDocuments: []ExtraDocument{
			{ID: extra, Title: "Foto da obra", File: &DocFile{ID: uuid.New(), Filename: "foto.jpg", Status: FileConfirmed}},
			{ID: uuid.New(), Title: "Sem arquivo"},
			{ID: uuid.New(), Title: "Arquivo pendente", File: &pending},
		},
	}
	ts := reviewTargets(cl)

	assert.Equal(t, []string{
		"section:proponent", "section:institution", "section:proposal", "section:funding", "section:document-options",
		"document:a", "document:b", "document:c", "document:e@em1", "extra:" + extra.String(),
	}, keysOf(ts))

	byKey := map[string]reviewTarget{}
	for _, x := range ts {
		byKey[x.key()] = x
	}
	assert.Len(t, byKey["document:a"].Files, 1, "só o arquivo confirmado precisa ser aberto")
	assert.Empty(t, byKey["document:b"].Files, "justificado não tem arquivo para abrir")
	assert.True(t, byKey["document:c"].Blocking)
	assert.Equal(t, reasonMandatoryMissing, byKey["document:c"].Reason)
	assert.Equal(t, "Doc e (Emenda 1/2026)", byKey["document:e@em1"].Label)
	assert.Equal(t, "Foto da obra", byKey["extra:"+extra.String()].Label)
	assert.Equal(t, "Proponente", byKey["section:proponent"].Label)
}

func TestTudoAprovado(t *testing.T) {
	f := confirmedFile("a.pdf")
	cl := &Checklist{Items: []ChecklistItem{item("a", ItemDone, true, f)}}
	ts := reviewTargets(cl)
	now := time.Now()
	approveAll := func() map[string]reviewRow {
		m := map[string]reviewRow{}
		for _, x := range ts {
			m[x.key()] = reviewRow{Status: ReviewApproved, By: "Ana", At: &now}
		}
		return m
	}

	t.Run("nada aprovado", func(t *testing.T) {
		r := buildReview(ts, nil, nil)
		assert.False(t, r.AllApproved)
		assert.Equal(t, ReviewSummary{Total: 6, Pending: 6}, r.Summary)
	})
	t.Run("tudo aprovado", func(t *testing.T) {
		r := buildReview(ts, approveAll(), map[uuid.UUID]fileSeen{f.ID: {Opened: true, Read: true}})
		assert.True(t, r.AllApproved)
		assert.Equal(t, ReviewSummary{Total: 6, Approved: 6}, r.Summary)
		assert.Equal(t, "Ana", r.Items[0].ReviewedBy)
		require.Len(t, r.Items[5].Files, 1)
		assert.True(t, r.Items[5].Files[0].Opened)
		assert.True(t, r.Items[5].Files[0].Read)
	})
	t.Run("reprovado conta à parte, mostra a justificativa e impede", func(t *testing.T) {
		saved := approveAll()
		saved["section:funding"] = reviewRow{Status: ReviewRejected, Note: "Valores não batem.", By: "Ana", At: &now}
		r := buildReview(ts, saved, nil)
		assert.False(t, r.AllApproved)
		assert.Equal(t, ReviewSummary{Total: 6, Approved: 5, Rejected: 1}, r.Summary)
		assert.Equal(t, ReviewRejected, r.Items[3].Status)
		assert.Equal(t, "Valores não batem.", r.Items[3].Note)
		assert.Equal(t, "Ana", r.Items[3].ReviewedBy)
		assert.Empty(t, r.Items[0].Note, "aprovado não tem justificativa")
	})
	t.Run("um item pendente impede", func(t *testing.T) {
		saved := approveAll()
		saved["section:funding"] = reviewRow{Status: ReviewPending}
		r := buildReview(ts, saved, nil)
		assert.False(t, r.AllApproved)
		assert.Equal(t, 1, r.Summary.Pending)
		assert.Empty(t, r.Items[3].ReviewedBy, "pendente não mostra quem aprovou")
	})
	t.Run("impedimento impede mesmo com o resto aprovado", func(t *testing.T) {
		withBlocker := reviewTargets(&Checklist{Items: []ChecklistItem{item("a", ItemDone, true, f), item("c", ItemMissing, true)}})
		saved := map[string]reviewRow{}
		for _, x := range withBlocker {
			saved[x.key()] = reviewRow{Status: ReviewApproved}
		}
		r := buildReview(withBlocker, saved, nil)
		assert.False(t, r.AllApproved)
		assert.Equal(t, 1, r.Summary.Blockers)
		assert.Equal(t, 6, r.Summary.Total, "impedimento não conta como item a conferir")
		blocked := r.Items[len(r.Items)-1]
		assert.True(t, blocked.Blocking)
		assert.Equal(t, ReviewPending, blocked.Status, "mesmo que haja linha gravada, o impedimento nunca aparece aprovado")
	})
	t.Run("sem nenhum item não há o que aprovar", func(t *testing.T) {
		assert.False(t, buildReview(nil, nil, nil).AllApproved)
	})
}

func TestArquivosNaoVistosENaoLidos(t *testing.T) {
	a, b := confirmedFile("a.pdf"), confirmedFile("b.pdf")
	tg := reviewTarget{Kind: ReviewDocument, Target: "x", Files: []DocFile{a, b}}
	opened := func(ids ...uuid.UUID) map[uuid.UUID]fileSeen {
		m := map[uuid.UUID]fileSeen{}
		for _, id := range ids {
			m[id] = fileSeen{Opened: true}
		}
		return m
	}
	assert.Equal(t, []DocFile{a, b}, unseen(tg, nil, false))
	assert.Equal(t, []DocFile{b}, unseen(tg, opened(a.ID), false))
	assert.Empty(t, unseen(tg, opened(a.ID, b.ID), false))
	assert.Empty(t, unseen(reviewTarget{Kind: ReviewSection, Target: "proponent"}, nil, true), "seção não exige abertura nem leitura")

	// Ler até o fim é mais que ver: vistos mas não lidos ainda faltam para aprovar.
	assert.Equal(t, []DocFile{a, b}, unseen(tg, opened(a.ID, b.ID), true))
	seen := opened(a.ID, b.ID)
	seen[a.ID] = fileSeen{Opened: true, Read: true}
	assert.Equal(t, []DocFile{b}, unseen(tg, seen, true))
	// Ler sem constar como visto não vale (o servidor sempre grava os dois juntos).
	assert.Equal(t, []DocFile{a, b}, unseen(tg, map[uuid.UUID]fileSeen{a.ID: {Read: true}, b.ID: {Read: true}}, true))
}
