package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/cpf"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/httpapi"
	"github.com/wagner/convenios-e-fomentos/backend/internal/mailer"
)

// ActionShare é a gestão do acesso ao projeto (convidar, alterar o nível, revogar). O proponente só a tem sobre os projetos que criou.
const ActionShare = "admin::projects.share"

// Condições de permissão do compartilhamento: somam-se à "é o criador" (condições de uma permissão se combinam com OR).
const (
	// ConditionMember: o projeto em que a pessoa é membro, com qualquer nível (leitura).
	ConditionMember = "admin::project-member"
	// ConditionEditor: o projeto em que a pessoa é membro com edição (alterar e enviar).
	ConditionEditor = "admin::project-editor"
)

// Níveis de acesso de um membro e valores de Project.Access.
const (
	LevelRead = "read"
	LevelEdit = "edit"
	// AccessOwner é o titular. Quem não é titular nem membro (a equipe) fica com Access vazio.
	AccessOwner = "owner"

	// maxMembers limita os convidados ativos por projeto.
	maxMembers = 10
)

// MemberLookup é o que as condições precisam do banco.
type MemberLookup interface {
	ListMembershipsOfAdmin(ctx context.Context, id uuid.UUID) ([]db.ListMembershipsOfAdminRow, error)
}

// RegisterConditions registra as condições do compartilhamento e as de visibilidade da equipe (visibility.go). A lista de projetos da pessoa é resolvida a cada requisição, antes
// da consulta (um proponente tem poucos projetos), o que mantém o filtro simples e faz a revogação valer na hora.
func RegisterConditions(reg *authz.ConditionRegistry, members MemberLookup) {
	handler := func(editOnly bool) authz.Handler {
		return func(ctx context.Context, s authz.Subject) (authz.Result, error) {
			rows, err := members.ListMembershipsOfAdmin(ctx, s.ID)
			if err != nil {
				return authz.Deny(), err
			}
			ids := make([]any, 0, len(rows))
			for _, m := range rows {
				if !editOnly || m.Level == LevelEdit {
					ids = append(ids, m.ProjectID)
				}
			}
			return authz.Restrict(authz.In("id", ids...)), nil
		}
	}
	reg.MustRegister(
		authz.Condition{ID: ConditionMember, Name: "É membro do projeto (leitura ou edição)", Category: "default", Handler: handler(false)},
		authz.Condition{ID: ConditionEditor, Name: "É membro do projeto com edição", Category: "default", Handler: handler(true)},
	)
	registerVisibilityConditions(reg)
}

// Member é uma pessoa com acesso a um projeto. O CPF sai mascarado, e a API não informa se a pessoa tem conta (nem o nome dela).
type Member struct {
	ID    uuid.UUID `json:"id"`
	CPF   string    `json:"cpf"`
	Label string    `json:"label"`
	Level string    `json:"level"`
	// Email é o e-mail que o titular informou (mascarado); vazio se não informou. Quem já tem conta é avisado no e-mail da conta, que não aparece aqui.
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
	// AcceptedAt é a primeira vez em que a pessoa abriu o projeto; vazio enquanto o convite está pendente.
	AcceptedAt *time.Time `json:"acceptedAt"`
}

// MemberInput é o convite (CPF e nível) ou a alteração (só nível e identificação).
type MemberInput struct {
	CPF   string `json:"cpf"`
	Label string `json:"label" validate:"max=100"`
	Level string `json:"level" validate:"required,oneof=read edit"`
	// Email é opcional: serve para avisar quem ainda não tem conta. Se o CPF já tem conta, o aviso vai para o e-mail dela.
	Email string `json:"email" validate:"max=254"`
}

func memberView(m db.ProjectMember) Member {
	return Member{ID: m.ID, CPF: maskCPF(m.Cpf), Label: m.Label, Level: m.Level, Email: maskEmail(str(m.Email)), CreatedAt: m.CreatedAt, AcceptedAt: m.AcceptedAt}
}

// maskEmail deixa só a primeira letra do usuário: m***@exemplo.com.
func maskEmail(e string) string {
	at := strings.LastIndex(e, "@")
	if at < 1 {
		return ""
	}
	return e[:1] + "***" + e[at:]
}

// accessOf diz como a pessoa se relaciona com o projeto: titular, membro (edit ou read) ou nenhum (vazio, caso da equipe).
func (s *Service) accessOf(ctx context.Context, q *db.Queries, projectID, createdBy uuid.UUID, actor Actor) (string, error) {
	if createdBy == actor.ID {
		return AccessOwner, nil
	}
	level, err := q.GetMemberLevel(ctx, db.GetMemberLevelParams{ID: actor.ID, ProjectID: projectID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return level, err
}

// memberships devolve o nível do administrador em cada projeto em que é membro.
func (s *Service) memberships(ctx context.Context, actor Actor) (map[uuid.UUID]string, error) {
	rows, err := s.q.ListMembershipsOfAdmin(ctx, actor.ID)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]string, len(rows))
	for _, r := range rows {
		out[r.ProjectID] = r.Level
	}
	return out, nil
}

// lockedForSharing trava o projeto e confere que a pessoa pode gerir o acesso dele.
func (s *Service) lockedForSharing(ctx context.Context, q *db.Queries, id uuid.UUID, dec authz.Decision) (db.Project, error) {
	cur, err := q.GetProjectForUpdate(ctx, id)
	if err != nil {
		return db.Project{}, notFound(err)
	}
	if !dec.AllowsRecord(refOf(cur).Getter()) {
		return db.Project{}, httpapi.ErrForbidden()
	}
	return cur, nil
}

func cleanMember(in MemberInput) (MemberInput, error) {
	in.Label = text(in.Label)
	if tooLong(in.Label, 100) {
		return in, fieldIssue("label", "Use até 100 caracteres.")
	}
	if in.Level != LevelRead && in.Level != LevelEdit {
		return in, fieldIssue("level", "Use read (leitura) ou edit (edição).")
	}
	if strings.TrimSpace(in.Email) != "" {
		e, ok := emailAddress(in.Email)
		if !ok {
			return in, fieldIssue("email", "E-mail inválido.")
		}
		in.Email = e
	} else {
		in.Email = ""
	}
	return in, nil
}

// Members lista as pessoas com acesso ao projeto (só o titular).
func (s *Service) Members(ctx context.Context, id uuid.UUID, dec authz.Decision) ([]Member, error) {
	cur, err := s.q.GetProject(ctx, id)
	if err != nil {
		return nil, notFound(err)
	}
	if !dec.AllowsRecord(refOf(cur).Getter()) {
		return nil, httpapi.ErrForbidden()
	}
	rows, err := s.q.ListProjectMembers(ctx, id)
	if err != nil {
		return nil, err
	}
	out := make([]Member, len(rows))
	for i, r := range rows {
		out[i] = memberView(r)
	}
	return out, nil
}

// Invite dá acesso ao CPF informado. Quem ainda não tem conta passa a ter o acesso quando entrar pelo MT Login.
func (s *Service) Invite(ctx context.Context, actor Actor, id uuid.UUID, in MemberInput, dec authz.Decision) (*Member, error) {
	in, err := cleanMember(in)
	if err != nil {
		return nil, err
	}
	digits, err := cpf.Normalize(in.CPF)
	if err != nil {
		return nil, fieldIssue("cpf", "CPF inválido.")
	}
	var (
		out   db.ProjectMember
		title string
	)
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		cur, err := s.lockedForSharing(ctx, q, id, dec)
		if err != nil {
			return err
		}
		// O titular já tem acesso total: convidar o CPF dele só criaria uma linha sem efeito.
		if cur.ProponentCpf == digits {
			return fieldIssue("cpf", "Este CPF é do titular do projeto, que já tem acesso total.")
		}
		prev, err := q.GetProjectMemberByCPF(ctx, db.GetProjectMemberByCPFParams{ProjectID: id, Cpf: digits})
		revoked := err == nil // já houve convite para este CPF (e, se chegou aqui, foi revogado)
		switch {
		case err == nil && prev.Status == "active":
			return httpapi.ErrConflict("member_exists", "Este CPF já tem acesso ao projeto. Altere o nível na lista de acessos.")
		case err != nil && !errors.Is(err, pgx.ErrNoRows):
			return err
		}
		n, err := q.CountActiveProjectMembers(ctx, id)
		if err != nil {
			return err
		}
		if n >= maxMembers {
			return httpapi.ErrConflict("members_limit", "O projeto atingiu o limite de pessoas com acesso. Revogue algum acesso antes de convidar outra pessoa.")
		}
		title = cur.Title
		if revoked { // acesso revogado antes: reativa a mesma linha
			out, err = q.ReactivateProjectMember(ctx, db.ReactivateProjectMemberParams{ID: prev.ID, Level: in.Level, Label: in.Label, InvitedByID: actor.ID, Email: nullText(in.Email)})
		} else {
			out, err = q.InsertProjectMember(ctx, db.InsertProjectMemberParams{ProjectID: id, Cpf: digits, Label: in.Label, Level: in.Level, InvitedByID: actor.ID, Email: nullText(in.Email)})
		}
		if err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("project.member-invited", id, map[string]any{"member": out.ID, "cpf": maskCPF(digits), "level": in.Level}))
	})
	if err != nil {
		return nil, err
	}
	s.notifyInvited(ctx, actor, id, title, out)
	v := memberView(out)
	return &v, nil
}

// notifyInvited avisa o convidado por e-mail (melhor esforço: falha vira log e não desfaz o convite). O e-mail da conta que tem o CPF
// tem prioridade e nunca é revelado ao titular; sem conta, vale o e-mail que o titular informou; sem nenhum dos dois, não há aviso.
func (s *Service) notifyInvited(ctx context.Context, actor Actor, projectID uuid.UUID, title string, m db.ProjectMember) {
	if s.mail == nil {
		return
	}
	to := str(m.Email)
	if acct, err := s.q.GetNotifiableAdminEmail(ctx, &m.Cpf); err == nil && acct != "" {
		to = acct
	}
	if to == "" {
		return
	}
	if title == "" {
		title = "(projeto sem título)"
	}
	what := "para consultar (somente leitura)"
	if m.Level == LevelEdit {
		what = "para consultar e editar"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Olá.\n\n%s deu a você acesso %s ao projeto \"%s\" no sistema de convênios e termos de fomento da SEDEC-MT.\n\n", actor.Name, what, title)
	fmt.Fprintf(&b, "Para acessar, entre no sistema pelo MT Login com o CPF %s:\n%s/dashboard/projetos/%s\n\n", maskCPF(m.Cpf), s.publicURL, projectID)
	b.WriteString("Se você não esperava este convite, ignore esta mensagem: o titular do projeto pode revogar o acesso a qualquer momento.\n")
	if err := s.mail.Send(ctx, mailer.Message{To: to, Subject: "Você recebeu acesso a um projeto na SEDEC-MT", Body: b.String()}); err != nil && s.log != nil {
		s.log.ErrorContext(ctx, "enviando o aviso de convite do projeto", "project", projectID, "error", err)
	}
}

// UpdateMember altera o nível ou a identificação de quem já tem acesso.
func (s *Service) UpdateMember(ctx context.Context, actor Actor, id, memberID uuid.UUID, in MemberInput, dec authz.Decision) (*Member, error) {
	in, err := cleanMember(in)
	if err != nil {
		return nil, err
	}
	var out db.ProjectMember
	err = database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := s.lockedForSharing(ctx, q, id, dec); err != nil {
			return err
		}
		prev, err := q.GetProjectMember(ctx, db.GetProjectMemberParams{ID: memberID, ProjectID: id})
		if err != nil || prev.Status != "active" {
			return httpapi.ErrNotFound()
		}
		if out, err = q.UpdateProjectMember(ctx, db.UpdateProjectMemberParams{ID: memberID, Level: in.Level, Label: in.Label, Email: nullText(in.Email)}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("project.member-updated", id, map[string]any{"member": memberID, "cpf": maskCPF(out.Cpf), "from": prev.Level, "level": in.Level}))
	})
	if err != nil {
		return nil, err
	}
	v := memberView(out)
	return &v, nil
}

// RevokeMember tira o acesso. Vale na hora: o acesso é conferido a cada requisição.
func (s *Service) RevokeMember(ctx context.Context, actor Actor, id, memberID uuid.UUID, dec authz.Decision) error {
	return database.WithTx(ctx, s.pool, func(tx pgx.Tx) error {
		q := s.q.WithTx(tx)
		if _, err := s.lockedForSharing(ctx, q, id, dec); err != nil {
			return err
		}
		prev, err := q.GetProjectMember(ctx, db.GetProjectMemberParams{ID: memberID, ProjectID: id})
		if err != nil || prev.Status != "active" {
			return httpapi.ErrNotFound()
		}
		if _, err := q.RevokeProjectMember(ctx, db.RevokeProjectMemberParams{ID: memberID, RevokedByID: &actor.ID}); err != nil {
			return err
		}
		return s.audit.Record(ctx, q, actor.entry("project.member-revoked", id, map[string]any{"member": memberID, "cpf": maskCPF(prev.Cpf), "level": prev.Level}))
	})
}

// markAccepted registra a primeira vez que um membro abriu o projeto (silencioso: falha aqui nunca atrapalha a leitura).
func (s *Service) markAccepted(ctx context.Context, projectID uuid.UUID, actor Actor) {
	if err := s.q.MarkMemberAccepted(ctx, db.MarkMemberAcceptedParams{ID: actor.ID, ProjectID: projectID}); err != nil && s.log != nil {
		s.log.WarnContext(ctx, "registrando o primeiro acesso do membro", "error", err)
	}
}
