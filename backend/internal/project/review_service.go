package project

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
)

// maxUnopenedListed limita quantos arquivos o erro file_not_opened lista (o total vai junto).
const maxUnopenedListed = 20

// AnalystInfo é o analista responsável, na visão da equipe.
type AnalystInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ReviewInput é o corpo de PUT /review/items: aprova (Approved) ou desfaz a aprovação de um item.
type ReviewInput struct {
	Kind     string `json:"kind"`   // section | document | extra
	Target   string `json:"target"` // seção, chave do documento ou id do documento adicional
	Approved bool   `json:"approved"`
}

func errStageForReview(stage string) error {
	if stage == StageAwaitingChecklist {
		return httpapi.ErrConflict("analysis_not_started", "Assuma a análise do projeto antes de conferir os itens.")
	}
	return errStage(stage)
}

// lockedForReview trava o projeto e confere permissão e situação (enviado e ativo).
func (s *Service) lockedForReview(ctx context.Context, q *db.Queries, id uuid.UUID, dec authz.Decision) (db.Project, error) {
	cur, err := q.GetProjectForUpdate(ctx, id)
	if err != nil {
		return db.Project{}, notFound(err)
	}
	if !dec.AllowsRecord(refOf(cur).Getter()) {
		return db.Project{}, httpapi.ErrForbidden()
	}
	if cur.Status != StatusSubmitted {
		return db.Project{}, httpapi.ErrConflict("project_not_submitted", "Só projetos enviados e ativos têm conferência.")
	}
	return cur, nil
}

// StartAnalysis assume a análise: "Aguardando análise de checklist" passa a "Em análise" e quem assumiu vira o responsável.
// É idempotente. Outro Analista que chame depois não troca o responsável: só passa a trabalhar no projeto (qualquer Analista pode
// agir). Se o projeto já está em análise mas sem responsável (voltou de uma complementação pedida antes de alguém assumir), o
// primeiro a assumir vira o responsável.
func (s *Service) StartAnalysis(ctx context.Context, actor Actor, id uuid.UUID, dec authz.Decision) (*Project, error) {
	var out *Project
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		cur, err := s.lockedForReview(ctx, q, id, dec)
		if err != nil {
			return err
		}
		row, changed := cur, false
		switch str(cur.Stage) {
		case StageAwaitingChecklist:
			if row, err = q.AssumeProjectAnalysis(ctx, db.AssumeProjectAnalysisParams{ID: id, AnalystID: &actor.ID}); err != nil {
				return notFound(err)
			}
			changed = true
		case StageInAnalysis:
			if cur.AnalystID != nil {
				break // já assumido: nada muda
			}
			if row, err = q.ClaimProjectAnalyst(ctx, db.ClaimProjectAnalystParams{ID: id, AnalystID: &actor.ID}); err != nil {
				return notFound(err)
			}
			changed = true
		default:
			return errStage(str(cur.Stage))
		}
		if changed {
			if err := s.audit.Record(ctx, q, actor.entry("project.start-analysis", id, map[string]any{"stage": StageInAnalysis})); err != nil {
				return err
			}
		}
		out, err = s.view(ctx, q, row, actor)
		return err
	})
	return out, err
}

// ReviewItem aprova ou desfaz a aprovação de um item da conferência. Exige o projeto em análise. Para aprovar um documento com
// arquivo, quem aprova precisa ter aberto cada arquivo confirmado atual do item (ver recordStaffOpening).
func (s *Service) ReviewItem(ctx context.Context, actor Actor, id uuid.UUID, in ReviewInput, dec authz.Decision) (*Project, error) {
	in.Target = strings.TrimSpace(in.Target)
	switch in.Kind {
	case ReviewSection, ReviewDocument, ReviewExtra:
	default:
		return nil, fieldIssue("kind", "Use section, document ou extra.")
	}
	var out *Project
	err := database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		cur, err := s.lockedForReview(ctx, q, id, dec)
		if err != nil {
			return err
		}
		if str(cur.Stage) != StageInAnalysis {
			return errStageForReview(str(cur.Stage))
		}
		if framingPending(cur) {
			return errFramingPending()
		}
		_, cl, err := s.viewWithChecklist(ctx, q, cur, actor)
		if err != nil {
			return err
		}
		t, ok := findTarget(reviewTargets(cl), in.Kind, in.Target)
		if !ok || t.Blocking {
			return &httpapi.APIError{Status: 422, Code: "item_not_reviewable", Message: "Este item não pode ser conferido agora.",
				Details: map[string]any{"fields": []httpapi.FieldError{{Field: "target", Message: "Item inexistente ou sem o que conferir (documento obrigatório sem arquivo nem justificativa)."}}}}
		}
		if in.Approved {
			opened, err := s.openedBy(ctx, q, id, actor.ID)
			if err != nil {
				return err
			}
			if missing := unopened(t, opened); len(missing) > 0 {
				return errFileNotOpened(missing)
			}
			err = q.ApproveItemReview(ctx, db.ApproveItemReviewParams{ProjectID: id, Kind: in.Kind, Target: in.Target, ReviewedByID: &actor.ID})
			if err != nil {
				return err
			}
		} else if err := q.ResetItemReview(ctx, db.ResetItemReviewParams{ProjectID: id, Kind: in.Kind, Target: in.Target}); err != nil {
			return err
		}
		if err := s.audit.Record(ctx, q, actor.entry("project.review-item", id, map[string]any{"kind": in.Kind, "target": in.Target, "approved": in.Approved})); err != nil {
			return err
		}
		out, err = s.view(ctx, q, cur, actor)
		return err
	})
	return out, err
}

func errFileNotOpened(missing []DocFile) error {
	type file struct {
		ID       uuid.UUID `json:"id"`
		Filename string    `json:"filename"`
	}
	listed := make([]file, 0, min(len(missing), maxUnopenedListed))
	for _, f := range missing[:min(len(missing), maxUnopenedListed)] {
		listed = append(listed, file{f.ID, f.Filename})
	}
	return &httpapi.APIError{Status: 422, Code: "file_not_opened", Message: "Abra todos os arquivos do documento antes de aprová-lo.",
		Details: map[string]any{"files": listed, "total": len(missing)}}
}

// openedBy devolve os arquivos do projeto que o administrador já abriu.
func (s *Service) openedBy(ctx context.Context, q *db.Queries, projectID, adminID uuid.UUID) (map[uuid.UUID]bool, error) {
	ids, err := q.ListFileViewsOfAdmin(ctx, db.ListFileViewsOfAdminParams{ProjectID: projectID, AdminID: adminID})
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]bool, len(ids))
	for _, id := range ids {
		out[id] = true
	}
	return out, nil
}

// recordStaffOpening registra que o administrador pediu o link do arquivo, se ele é da equipe (nem titular nem membro). Quem baixa o
// próprio arquivo não gera registro. Se o registro falhar o download falha: melhor do que permitir aprovar sem a prova.
func (s *Service) recordStaffOpening(ctx context.Context, projectID, fileID uuid.UUID, actor Actor) error {
	if actor.ID == uuid.Nil {
		return nil
	}
	proj, err := s.q.GetProject(ctx, projectID)
	if err != nil {
		return notFound(err)
	}
	access, err := s.accessOf(ctx, s.q, projectID, proj.CreatedByID, actor)
	if err != nil || access != "" {
		return err
	}
	return s.q.RecordFileView(ctx, db.RecordFileViewParams{FileID: fileID, AdminID: actor.ID})
}

// reviewState monta a conferência a partir do que está gravado. opened pode ser nulo quando só interessa o resultado (allApproved).
func (s *Service) reviewState(ctx context.Context, q *db.Queries, projectID uuid.UUID, cl *Checklist, opened map[uuid.UUID]bool) (Review, error) {
	rows, err := q.ListItemReviews(ctx, projectID)
	if err != nil {
		return Review{}, err
	}
	saved := make(map[string]reviewRow, len(rows))
	for _, x := range rows {
		saved[x.Kind+":"+x.Target] = reviewRow{Status: x.Status, By: x.ReviewedByName, At: x.ReviewedAt}
	}
	return buildReview(reviewTargets(cl), saved, opened), nil
}

// attachReview acrescenta à visão da equipe o analista responsável e a conferência. O titular e os membros não recebem (é interna),
// nem rascunhos (a equipe não os vê), nem as chamadas internas sem pessoa (Actor vazio).
func (s *Service) attachReview(ctx context.Context, q *db.Queries, r db.Project, p *Project, cl *Checklist, actor Actor) error {
	if p.Access != "" || actor.ID == uuid.Nil || r.Status == StatusDraft {
		return nil
	}
	opened, err := s.openedBy(ctx, q, r.ID, actor.ID)
	if err != nil {
		return err
	}
	rv, err := s.reviewState(ctx, q, r.ID, cl, opened)
	if err != nil {
		return err
	}
	p.Review = &rv
	if p.Framing, p.Framings, err = s.framingViews(ctx, q, r); err != nil {
		return err
	}
	if r.AnalystID != nil {
		a, err := q.GetAdminByID(ctx, *r.AnalystID)
		if err == nil {
			name := strings.TrimSpace(a.FirstName + " " + a.LastName)
			if name == "" {
				name = a.Email
			}
			p.Analyst = &AnalystInfo{ID: a.ID, Name: name}
		}
	}
	return nil
}
