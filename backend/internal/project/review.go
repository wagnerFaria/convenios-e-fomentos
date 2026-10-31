package project

import (
	"time"

	"github.com/google/uuid"
)

// Conferência da análise. Aqui está só a regra pura (sem banco): quais itens do projeto a equipe confere e como se calcula o
// "tudo aprovado". As ações (assumir, aprovar, desfazer) estão em review_service.go.

// Tipos de item conferido. O alvo usa as mesmas chaves dos pedidos de complementação.
const (
	ReviewSection  = "section"  // alvo: nome da seção (proponent, institution, proposal, funding, document-options)
	ReviewDocument = "document" // alvo: "codigo" ou "codigo@id-da-emenda" do checklist
	ReviewExtra    = "extra"    // alvo: id do documento adicional

	ReviewPending  = "pending"
	ReviewApproved = "approved"
)

// reviewSections é a ordem em que as seções aparecem na conferência.
var reviewSections = []string{"proponent", "institution", "proposal", "funding", "document-options"}

const reasonMandatoryMissing = "Documento obrigatório sem arquivo nem justificativa."

// reviewTarget é um item que a equipe pode conferir (ou que impede a conferência) no estado atual do projeto.
type reviewTarget struct {
	Kind, Target, Label string
	// Files são os arquivos confirmados do item: a equipe precisa abrir cada um antes de aprovar.
	Files []DocFile
	// Blocking marca o impedimento: não há o que aprovar, e a equipe devolve ou arquiva.
	Blocking bool
	Reason   string
}

func (t reviewTarget) key() string { return t.Kind + ":" + t.Target }

// reviewTargets calcula, a partir do estado atual, os itens que entram na conferência:
//   - as cinco seções, sempre;
//   - os documentos do checklist com arquivo confirmado ou justificativa;
//   - os documentos adicionais com arquivo confirmado;
//   - como impedimento, o documento obrigatório sem arquivo nem justificativa.
//
// Documento opcional vazio não entra.
func reviewTargets(cl *Checklist) []reviewTarget {
	out := make([]reviewTarget, 0, len(reviewSections)+len(cl.Items)+len(cl.ExtraDocuments))
	for _, sec := range reviewSections {
		out = append(out, reviewTarget{Kind: ReviewSection, Target: sec, Label: returnableSections[sec]})
	}
	for _, it := range cl.Items {
		label := it.Name
		if it.AmendmentLabel != "" {
			label += " (" + it.AmendmentLabel + ")"
		}
		t := reviewTarget{Kind: ReviewDocument, Target: it.key(), Label: label}
		switch {
		case it.Status == ItemDone || it.Status == ItemJustified:
			t.Files = confirmed(it.Files)
		case it.Status == ItemMissing && it.Mandatory:
			t.Blocking, t.Reason = true, reasonMandatoryMissing
		default:
			continue
		}
		out = append(out, t)
	}
	for _, x := range cl.ExtraDocuments {
		if x.File == nil || x.File.Status != FileConfirmed {
			continue
		}
		out = append(out, reviewTarget{Kind: ReviewExtra, Target: x.ID.String(), Label: x.Title, Files: []DocFile{*x.File}})
	}
	return out
}

func confirmed(files []DocFile) []DocFile {
	var out []DocFile
	for _, f := range files {
		if f.Status == FileConfirmed {
			out = append(out, f)
		}
	}
	return out
}

// find acha o item pelo tipo e alvo.
func findTarget(targets []reviewTarget, kind, target string) (reviewTarget, bool) {
	for _, t := range targets {
		if t.Kind == kind && t.Target == target {
			return t, true
		}
	}
	return reviewTarget{}, false
}

// unopened devolve os arquivos do item que o administrador ainda não abriu.
func unopened(t reviewTarget, opened map[uuid.UUID]bool) []DocFile {
	var out []DocFile
	for _, f := range t.Files {
		if !opened[f.ID] {
			out = append(out, f)
		}
	}
	return out
}

// reviewRow é a situação gravada de um item.
type reviewRow struct {
	Status string
	By     string
	At     *time.Time
}

// ReviewFile é um arquivo de um item conferido. Opened é de quem consulta: cada Analista vê o que ele mesmo abriu.
type ReviewFile struct {
	ID       uuid.UUID `json:"id"`
	Filename string    `json:"filename"`
	Opened   bool      `json:"opened"`
}

// ReviewItem é um item na visão da equipe.
type ReviewItem struct {
	Kind       string       `json:"kind"`
	Target     string       `json:"target"`
	Label      string       `json:"label"`
	Status     string       `json:"status"`
	Blocking   bool         `json:"blocking"`
	Reason     string       `json:"reason,omitempty"`
	ReviewedBy string       `json:"reviewedBy,omitempty"`
	ReviewedAt *time.Time   `json:"reviewedAt,omitempty"`
	Files      []ReviewFile `json:"files"`
}

// ReviewSummary conta os itens. Total não inclui os impedimentos, que não são conferíveis.
type ReviewSummary struct {
	Total    int `json:"total"`
	Approved int `json:"approved"`
	Pending  int `json:"pending"`
	Blockers int `json:"blockers"`
}

// Review é a conferência do projeto, na visão da equipe.
type Review struct {
	Items   []ReviewItem  `json:"items"`
	Summary ReviewSummary `json:"summary"`
	// AllApproved: há itens, todos aprovados e nenhum impedimento. É a condição para indicar o departamento.
	AllApproved bool `json:"allApproved"`
}

// buildReview junta os itens conferíveis com o que foi gravado e com os arquivos que o administrador abriu.
func buildReview(targets []reviewTarget, saved map[string]reviewRow, opened map[uuid.UUID]bool) Review {
	r := Review{Items: make([]ReviewItem, 0, len(targets))}
	for _, t := range targets {
		it := ReviewItem{Kind: t.Kind, Target: t.Target, Label: t.Label, Status: ReviewPending, Blocking: t.Blocking, Reason: t.Reason, Files: []ReviewFile{}}
		for _, f := range t.Files {
			it.Files = append(it.Files, ReviewFile{ID: f.ID, Filename: f.Filename, Opened: opened[f.ID]})
		}
		if row, ok := saved[t.key()]; ok && !t.Blocking {
			it.Status, it.ReviewedBy, it.ReviewedAt = row.Status, row.By, row.At
			if row.Status != ReviewApproved {
				it.ReviewedBy, it.ReviewedAt = "", nil
			}
		}
		switch {
		case t.Blocking:
			r.Summary.Blockers++
		default:
			r.Summary.Total++
			if it.Status == ReviewApproved {
				r.Summary.Approved++
			}
		}
		r.Items = append(r.Items, it)
	}
	r.Summary.Pending = r.Summary.Total - r.Summary.Approved
	r.AllApproved = r.Summary.Total > 0 && r.Summary.Pending == 0 && r.Summary.Blockers == 0
	return r
}
