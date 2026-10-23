package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/wagner/convenios-e-fomentos/backend/internal/authz"
	"github.com/wagner/convenios-e-fomentos/backend/internal/catalog"
	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
	"github.com/wagner/convenios-e-fomentos/backend/internal/rules"
)

// Tipos de recurso.
const (
	FundingAmendment = "amendment"
	FundingOwn       = "own"
	FundingMixed     = "mixed"

	CounterpartFinancial = "financial"
	CounterpartInKind    = "in_kind"

	TenderHeld          = "held"
	TenderWaived        = "waived"
	TenderUnfeasible    = "unfeasible"
	TenderNotApplicable = "not_applicable"
)

const (
	// maxCents limita qualquer valor (R$ 10 bilhões): protege contra erro de digitação e contra estouro nas contas.
	maxCents       = int64(1_000_000_000_000)
	maxAmendments  = 20
	maxAmendNumber = 30
	maxFreeName    = 120
	maxCounterpart = 1000
	maxTenderText  = 2000
	maxNotice      = 50
)

// tenderBasis é um fundamento legal de dispensa ou inexigibilidade do chamamento público.
type tenderBasis struct{ status, instrument string }

// tenderBases são os fundamentos aceitos. Fomento: arts. 19 (dispensa) e 20 (inexigibilidade) da IN Conjunta 001/2016.
// Convênio com entidade privada: exceções do art. 14 da IN Conjunta 001/2015 (o código começa por "conv14").
var tenderBases = map[string]tenderBasis{
	"art19_i":        {TenderWaived, InstrumentFomento},
	"art19_ii":       {TenderWaived, InstrumentFomento},
	"art19_iii":      {TenderWaived, InstrumentFomento},
	"art19_iv":       {TenderWaived, InstrumentFomento},
	"art20_i":        {TenderUnfeasible, InstrumentFomento},
	"art20_ii":       {TenderUnfeasible, InstrumentFomento},
	"art20_singular": {TenderUnfeasible, InstrumentFomento},
	"conv14_i":       {TenderWaived, InstrumentConvenio},
	"conv14_ii":      {TenderWaived, InstrumentConvenio},
	"conv14_iii":     {TenderWaived, InstrumentConvenio},
}

// municipalTypes são os tipos de entidade tratados como Município na Lei 12.082 (contrapartida mínima).
var municipalTypes = map[string]bool{"municipal_government": true, "municipal_fund": true}

func cents(v int64, field string, p *problems) int64 {
	if v < 0 || v > maxCents {
		p.add(field, "Informe um valor entre R$ 0,00 e R$ 10.000.000.000,00.")
		return 0
	}
	return v
}

// SaveFunding grava a seção Recursos. Valida o formato e a coerência entre os campos (422); os limites que dependem do
// valor total e das normas vigentes viram pendências (ver fundingIssues), para o rascunho poder ser salvo enquanto se ajusta.
func (s *Service) SaveFunding(ctx context.Context, actor Actor, id uuid.UUID, version int32, in FundingSection, upd authz.Decision) (*Project, error) {
	var orphans []string // arquivos das emendas removidas: apagados do armazenamento depois do commit
	out, err := s.mutate(ctx, actor, id, version, upd, "funding", func(q *db.Queries, cur db.Project) (db.Project, error) {
		var p problems
		arg := db.UpdateProjectFundingParams{ID: id, UpdatedByID: &actor.ID}

		switch in.Type {
		case "", FundingAmendment, FundingOwn, FundingMixed:
			arg.FundingType = nullText(in.Type)
		default:
			p.add("type", "Use emenda parlamentar, recurso próprio ou os dois.")
		}
		arg.OwnAmountCents = cents(in.OwnAmountCents, "ownAmountCents", &p)
		switch {
		case in.Type == FundingAmendment && arg.OwnAmountCents != 0:
			p.add("ownAmountCents", "No tipo \"Emenda parlamentar\" não há recurso próprio da SEDEC.")
		case in.Type == "" && arg.OwnAmountCents != 0:
			p.add("type", "Escolha o tipo de recurso antes de informar valores.")
		}

		amendments := s.amendments(ctx, q, id, in, &p)
		for _, a := range amendments {
			arg.AmendmentAmountCents += a.params.AmountCents
		}
		arg.TotalAmountCents = arg.OwnAmountCents + arg.AmendmentAmountCents

		s.counterpart(in, cur, &arg, &p)
		arg.IndirectCostsCents = cents(in.IndirectCostsCents, "indirectCostsCents", &p)
		if arg.IndirectCostsCents > 0 && cur.Instrument != nil && *cur.Instrument != InstrumentFomento {
			p.add("indirectCostsCents", "Custos indiretos são do Termo de Fomento.")
		}
		s.tender(in, cur, &arg, &p)
		s.artists(in, cur, &arg, &p)
		arg.PriceOldestCollectedOn = date(in.PriceOldestCollectedOn, "priceOldestCollectedOn", &p)
		if arg.PriceOldestCollectedOn.Valid && arg.PriceOldestCollectedOn.Time.After(s.today()) {
			p.add("priceOldestCollectedOn", "A data de coleta não pode ser futura.")
		}

		if err := p.err(); err != nil {
			return db.Project{}, err
		}
		// Emendas mantidas são atualizadas no lugar (os documentos anexados a elas continuam valendo); as que saíram são
		// apagadas, junto com os arquivos delas, e as novas são inseridas.
		keep := make([]uuid.UUID, 0, len(amendments))
		for _, a := range amendments {
			if a.id != uuid.Nil {
				keep = append(keep, a.id)
			}
		}
		if locked, err := q.CountLockedFilesOfRemovedAmendments(ctx, db.CountLockedFilesOfRemovedAmendmentsParams{ProjectID: id, Keep: keep}); err != nil {
			return db.Project{}, err
		} else if locked > 0 {
			return db.Project{}, fieldIssue("amendments", "Esta emenda já tem documentos enviados e não pode ser removida; corrija os dados dela.")
		}
		keys, err := q.ListProjectFileKeysOfRemovedAmendments(ctx, db.ListProjectFileKeysOfRemovedAmendmentsParams{ProjectID: id, Keep: keep})
		if err != nil {
			return db.Project{}, err
		}
		orphans = append(orphans, keys...)
		if err := q.DeleteProjectAmendmentsExcept(ctx, db.DeleteProjectAmendmentsExceptParams{ProjectID: id, Keep: keep}); err != nil {
			return db.Project{}, err
		}
		for _, a := range amendments {
			if a.id == uuid.Nil {
				if err := q.AddProjectAmendment(ctx, a.params); err != nil {
					return db.Project{}, err
				}
				continue
			}
			if err := q.UpdateProjectAmendment(ctx, db.UpdateProjectAmendmentParams{ID: a.id, ProjectID: id, Position: a.params.Position,
				ParliamentarianID: a.params.ParliamentarianID, ParliamentarianFreeName: a.params.ParliamentarianFreeName,
				AmendmentNumber: a.params.AmendmentNumber, Year: a.params.Year, AmountCents: a.params.AmountCents}); err != nil {
				return db.Project{}, err
			}
		}
		return q.UpdateProjectFunding(ctx, arg)
	})
	if err == nil {
		s.deleteObjects(ctx, orphans)
	}
	return out, err
}

// amendRow é uma emenda validada: id é o de uma emenda existente do projeto (zero = nova).
type amendRow struct {
	id     uuid.UUID
	params db.AddProjectAmendmentParams
}

// amendments valida as linhas de emenda e as converte para o banco.
func (s *Service) amendments(ctx context.Context, q *db.Queries, id uuid.UUID, in FundingSection, p *problems) []amendRow {
	if in.Type != FundingAmendment && in.Type != FundingMixed {
		if len(in.Amendments) > 0 {
			p.add("amendments", "Só os tipos \"Emenda parlamentar\" e \"Emenda e recursos próprios\" têm emendas.")
		}
		return nil
	}
	if len(in.Amendments) > maxAmendments {
		p.add("amendments", "Informe no máximo 20 emendas.")
		return nil
	}
	// Um parlamentar desativado depois da escolha segue valendo no projeto que já o usa.
	used := map[uuid.UUID]bool{}
	existingIDs := map[uuid.UUID]bool{}
	if existing, err := q.ListProjectAmendments(ctx, id); err == nil {
		for _, e := range existing {
			existingIDs[e.ID] = true
			if e.ParliamentarianID != nil {
				used[*e.ParliamentarianID] = true
			}
		}
	}
	year := s.today().Year()
	seen := map[string]bool{}
	out := make([]amendRow, 0, len(in.Amendments))
	kept := map[uuid.UUID]bool{}
	for i, a := range in.Amendments {
		f := fmt.Sprintf("amendments[%d]", i)
		row := db.AddProjectAmendmentParams{ProjectID: id, Position: int32(i), AmendmentNumber: text(a.Number)} //nolint:gosec // i < 20
		if tooLong(row.AmendmentNumber, maxAmendNumber) {
			p.add(f+".number", "Use até 30 caracteres.")
		}
		row.AmountCents = cents(a.AmountCents, f+".amountCents", p)
		if a.Year != 0 {
			if a.Year < 2000 || a.Year > year+1 {
				p.add(f+".year", "Informe o ano da emenda.")
			}
			y := int32(a.Year) //nolint:gosec // validado acima
			row.Year = &y
		}
		switch pid := strings.TrimSpace(a.ParliamentarianID); {
		case pid != "":
			u, err := uuid.Parse(pid)
			if err != nil {
				p.add(f+".parliamentarianId", "Parlamentar inválido.")
				break
			}
			par, err := q.GetParliamentarian(ctx, u)
			if err != nil {
				p.add(f+".parliamentarianId", "Parlamentar não encontrado.")
				break
			}
			if !par.Active && !used[u] {
				p.add(f+".parliamentarianId", "Este parlamentar não está mais disponível.")
				break
			}
			row.ParliamentarianID = &u
		default:
			row.ParliamentarianFreeName = text(a.ParliamentarianName)
			if tooLong(row.ParliamentarianFreeName, maxFreeName) {
				p.add(f+".parliamentarianName", "Use até 120 caracteres.")
			}
		}
		key := fmt.Sprintf("%v|%s|%s|%s|%d", row.ParliamentarianID, strings.ToLower(row.ParliamentarianFreeName), row.AmendmentNumber, "", a.Year)
		if row.AmendmentNumber != "" && seen[key] {
			p.add(f+".number", "Emenda repetida.")
		}
		seen[key] = true
		var rowID uuid.UUID
		if u, err := uuid.Parse(strings.TrimSpace(a.ID)); err == nil && existingIDs[u] && !kept[u] {
			rowID, kept[u] = u, true
		}
		out = append(out, amendRow{id: rowID, params: row})
	}
	return out
}

// counterpart valida a contrapartida do proponente. No Termo de Fomento ela não pode ser em dinheiro (art. 28 §1º da IN 001/2016).
func (s *Service) counterpart(in FundingSection, cur db.Project, arg *db.UpdateProjectFundingParams, p *problems) {
	switch in.CounterpartKind {
	case "", CounterpartFinancial, CounterpartInKind:
		arg.CounterpartKind = nullText(in.CounterpartKind)
	default:
		p.add("counterpartKind", "Use financeira ou em bens e serviços.")
	}
	arg.CounterpartCents = cents(in.CounterpartCents, "counterpartCents", p)
	arg.CounterpartDescription = multiline(in.CounterpartDescription)
	if tooLong(arg.CounterpartDescription, maxCounterpart) {
		p.add("counterpartDescription", "Use até 1.000 caracteres.")
	}
	if in.CounterpartKind == "" && (arg.CounterpartCents != 0 || arg.CounterpartDescription != "") {
		p.add("counterpartKind", "Escolha o tipo da contrapartida.")
	}
	if in.CounterpartKind == CounterpartFinancial && cur.Instrument != nil && *cur.Instrument == InstrumentFomento {
		p.add("counterpartKind", "No Termo de Fomento a contrapartida não pode ser em dinheiro: use bens e serviços, com o valor em reais.")
	}
}

// tender valida os dados do chamamento público.
func (s *Service) tender(in FundingSection, cur db.Project, arg *db.UpdateProjectFundingParams, p *problems) {
	switch in.TenderStatus {
	case "", TenderHeld, TenderWaived, TenderUnfeasible, TenderNotApplicable:
		arg.TenderStatus = nullText(in.TenderStatus)
	default:
		p.add("tenderStatus", "Situação do chamamento inválida.")
	}
	arg.TenderNoticeNumber = text(in.TenderNoticeNumber)
	arg.TenderJustification = multiline(in.TenderJustification)
	arg.TenderApprovedOn = date(in.TenderApprovedOn, "tenderApprovedOn", p)
	if tooLong(arg.TenderNoticeNumber, maxNotice) {
		p.add("tenderNoticeNumber", "Use até 50 caracteres.")
	}
	if tooLong(arg.TenderJustification, maxTenderText) {
		p.add("tenderJustification", "Use até 2.000 caracteres.")
	}
	if arg.TenderApprovedOn.Valid && arg.TenderApprovedOn.Time.After(s.today()) {
		p.add("tenderApprovedOn", "A data de homologação não pode ser futura.")
	}
	if b := strings.TrimSpace(in.TenderLegalBasis); b != "" {
		basis, ok := tenderBases[b]
		switch {
		case !ok:
			p.add("tenderLegalBasis", "Fundamento legal inválido.")
		case basis.status != in.TenderStatus:
			p.add("tenderLegalBasis", "O fundamento não corresponde à situação do chamamento.")
		case cur.Instrument != nil && basis.instrument != *cur.Instrument:
			p.add("tenderLegalBasis", "O fundamento não se aplica a este instrumento.")
		default:
			arg.TenderLegalBasis = b
		}
	}
}

// artists valida os valores de cachês (Lei 12.082), que só existem se o projeto prevê contratação de artistas.
func (s *Service) artists(in FundingSection, cur db.Project, arg *db.UpdateProjectFundingParams, p *problems) {
	arg.ArtistFeesCents = cents(in.ArtistFeesCents, "artistFeesCents", p)
	arg.RegionalArtistFeesCents = cents(in.RegionalArtistFeesCents, "regionalArtistFeesCents", p)
	arg.GovernorWaiver = in.GovernorWaiver
	if !cur.HasArtisticHiring && (arg.ArtistFeesCents > 0 || arg.RegionalArtistFeesCents > 0 || in.GovernorWaiver) {
		p.add("artistFeesCents", "Marque \"contratação de artistas\" na etapa Proposta antes de informar cachês.")
	}
	if arg.RegionalArtistFeesCents > arg.ArtistFeesCents {
		p.add("regionalArtistFeesCents", "Os cachês de artistas regionais não podem passar do total de cachês.")
	}
}

// ---- Leitura ----

func (s *Service) fundingView(ctx context.Context, q *db.Queries, r db.Project) (FundingSection, error) {
	rows, err := q.ListProjectAmendments(ctx, r.ID)
	if err != nil {
		return FundingSection{}, err
	}
	f := FundingSection{
		Type: str(r.FundingType), OwnAmountCents: r.OwnAmountCents, AmendmentTotalCents: r.AmendmentAmountCents, TotalCents: r.TotalAmountCents,
		Amendments: make([]Amendment, len(rows)), CounterpartKind: str(r.CounterpartKind), CounterpartCents: r.CounterpartCents,
		CounterpartDescription: r.CounterpartDescription, IndirectCostsCents: r.IndirectCostsCents, TenderStatus: str(r.TenderStatus),
		TenderNoticeNumber: r.TenderNoticeNumber, TenderApprovedOn: dateString(r.TenderApprovedOn), TenderLegalBasis: r.TenderLegalBasis,
		TenderJustification: r.TenderJustification, ArtistFeesCents: r.ArtistFeesCents, RegionalArtistFeesCents: r.RegionalArtistFeesCents,
		GovernorWaiver: r.GovernorWaiver, PriceOldestCollectedOn: dateString(r.PriceOldestCollectedOn),
	}
	for i, a := range rows {
		name := a.ParliamentarianName
		if a.ParliamentarianID == nil {
			name = a.ParliamentarianFreeName
		}
		f.Amendments[i] = Amendment{ID: a.ID.String(), ParliamentarianID: uuidString(a.ParliamentarianID), ParliamentarianName: name, Number: a.AmendmentNumber, AmountCents: a.AmountCents}
		if a.Year != nil {
			f.Amendments[i].Year = int(*a.Year)
		}
	}
	return f, nil
}

// ---- Pendências ----

func brl(c int64) string {
	reais, cent := c/100, c%100
	s := fmt.Sprintf("%d", reais)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return fmt.Sprintf("R$ %s,%02d", s, cent)
}

// indirectLimit devolve o máximo de custos indiretos para o valor total (IN Conjunta 002/2026, art. 36-A §1º):
// o menor entre o percentual da faixa e o teto em reais.
func indirectLimit(tiers []rules.IndirectCostTier, total int64) (int64, bool) {
	for _, t := range tiers {
		if t.UpToCents == nil || total <= *t.UpToCents {
			return min(total*int64(t.MaxPct)/100, t.CapCents), true
		}
	}
	return 0, false
}

// tenderRequired informa se o chamamento público é exigido: Fomento com recurso próprio da SEDEC (puro ou misto) e
// Convênio com entidade privada. Fomento só com emenda não exige (art. 14 da IN 001/2016).
func tenderRequired(instrument, nature, fundingType string) bool {
	switch instrument {
	case InstrumentFomento:
		return fundingType == FundingOwn || fundingType == FundingMixed
	case InstrumentConvenio:
		return nature == catalog.NaturePrivate
	}
	return false
}

func fundingIssues(l *issueList, p Project, snap *rules.Snapshot, today time.Time) {
	const sec = "funding"
	f, inst := p.Funding, p.Institution
	l.required(sec, "type", "Escolha o tipo de recurso.", f.Type == "")

	if f.Type == FundingAmendment || f.Type == FundingMixed {
		if len(f.Amendments) == 0 {
			l.add(sec, "amendments", "required", SeverityBlock, "Informe pelo menos uma emenda parlamentar.")
		}
		for i, a := range f.Amendments {
			n := i + 1
			if a.ParliamentarianName == "" {
				l.add(sec, fmt.Sprintf("amendments[%d].parliamentarianId", i), "required", SeverityBlock, fmt.Sprintf("Emenda %d: informe o parlamentar.", n))
			}
			if a.Number == "" {
				l.add(sec, fmt.Sprintf("amendments[%d].number", i), "required", SeverityBlock, fmt.Sprintf("Emenda %d: informe o número da emenda.", n))
			}
			if a.Year == 0 {
				l.add(sec, fmt.Sprintf("amendments[%d].year", i), "required", SeverityBlock, fmt.Sprintf("Emenda %d: informe o ano.", n))
			}
			if a.AmountCents == 0 {
				l.add(sec, fmt.Sprintf("amendments[%d].amountCents", i), "required", SeverityBlock, fmt.Sprintf("Emenda %d: informe o valor.", n))
			}
		}
	}
	if (f.Type == FundingOwn || f.Type == FundingMixed) && f.OwnAmountCents == 0 {
		l.add(sec, "ownAmountCents", "required", SeverityBlock, "Informe o valor do recurso próprio da SEDEC.")
	}
	if f.Type != "" && f.TotalCents == 0 {
		l.add(sec, "totalCents", "required", SeverityBlock, "O valor total solicitado precisa ser maior que zero.")
	}

	counterpartIssues(l, p)
	if inst.Instrument == InstrumentFomento {
		if f.IndirectCostsCents > 0 {
			if limit, ok := indirectLimit(snap.IndirectTiers, f.TotalCents); ok && f.IndirectCostsCents > limit {
				l.add(sec, "indirectCostsCents", "indirect_costs_over_limit", SeverityBlock,
					fmt.Sprintf("Custos indiretos acima do limite para este valor: no máximo %s (IN Conjunta 002/2026, art. 36-A).", brl(limit)))
			}
		}
		oldest, err := time.Parse(DateLayout, f.PriceOldestCollectedOn)
		if snap.PriceMaxAgeDays > 0 && err == nil && today.Sub(oldest) > time.Duration(snap.PriceMaxAgeDays)*24*time.Hour {
			l.add(sec, "priceOldestCollectedOn", "price_reference_old", SeverityWarn,
				fmt.Sprintf("A referência de preço mais antiga tem mais de %d dias; o ideal é refazer a pesquisa (IN Conjunta 002/2026).", snap.PriceMaxAgeDays))
		}
	}
	tenderIssues(l, p)
	artistIssues(l, p, snap)
}

func counterpartIssues(l *issueList, p Project) {
	const sec = "funding"
	f, inst := p.Funding, p.Institution
	if f.CounterpartKind == CounterpartFinancial && inst.Instrument == InstrumentFomento {
		l.add(sec, "counterpartKind", "counterpart_not_financial", SeverityBlock, "No Termo de Fomento a contrapartida não pode ser em dinheiro: use bens e serviços.")
	}
	if f.CounterpartKind != "" && f.CounterpartCents == 0 {
		l.add(sec, "counterpartCents", "required", SeverityBlock, "Informe o valor da contrapartida.")
	}
	if f.CounterpartKind == CounterpartInKind && f.CounterpartDescription == "" {
		l.add(sec, "counterpartDescription", "required", SeverityBlock, "Descreva os bens e serviços da contrapartida.")
	}
	if inst.Instrument == InstrumentConvenio && f.CounterpartKind == "" && !p.Proposal.Calamity {
		l.add(sec, "counterpartKind", "counterpart_missing", SeverityWarn, "Convênios costumam exigir contrapartida (limites da LDO). Confirme com a concedente; em calamidade ou emergência ela é dispensada.")
	}
}

func tenderIssues(l *issueList, p Project) {
	const sec = "funding"
	f, inst := p.Funding, p.Institution
	if !tenderRequired(inst.Instrument, inst.Nature, f.Type) {
		return
	}
	switch f.TenderStatus {
	case "", TenderNotApplicable:
		l.add(sec, "tenderStatus", "tender_required", SeverityBlock, "Este projeto exige chamamento público: informe o chamamento realizado ou o fundamento da dispensa ou inexigibilidade.")
	case TenderHeld:
		l.required(sec, "tenderNoticeNumber", "Informe o número do edital do chamamento.", f.TenderNoticeNumber == "")
		l.required(sec, "tenderApprovedOn", "Informe a data da homologação do chamamento.", f.TenderApprovedOn == "")
	default:
		l.required(sec, "tenderLegalBasis", "Informe o fundamento legal.", f.TenderLegalBasis == "")
		l.required(sec, "tenderJustification", "Justifique a ausência do chamamento.", f.TenderJustification == "")
	}
}

// artistIssues aplica a Lei Estadual 12.082/2023 a projetos com contratação de artistas. Os limites vêm dos parâmetros
// vigentes. Com autorização expressa do Governador (art. 5º), os limites podem ser afastados.
func artistIssues(l *issueList, p Project, snap *rules.Snapshot) {
	const sec = "funding"
	f := p.Funding
	if !p.Proposal.HasArtisticHiring {
		return
	}
	l.required(sec, "artistFeesCents", "Informe o valor total de cachês artísticos.", f.ArtistFeesCents == 0)
	if f.GovernorWaiver || snap.Lei12082 == nil {
		return
	}
	lei, total := snap.Lei12082, f.TotalCents
	if total > lei.CeilingCents {
		l.add(sec, "totalCents", "lei12082_ceiling", SeverityBlock, fmt.Sprintf("Lei 12.082/2023: o repasse para shows e eventos musicais é limitado a %s.", brl(lei.CeilingCents)))
	}
	if municipalTypes[p.Institution.EntityTypeCode] && f.CounterpartCents*100 < total*int64(lei.MinMunicipalCounterpartPct) {
		l.add(sec, "counterpartCents", "lei12082_counterpart", SeverityBlock,
			fmt.Sprintf("Lei 12.082/2023: convênio com município exige contrapartida mínima de %d%% do repasse (%s).", lei.MinMunicipalCounterpartPct, brl((total*int64(lei.MinMunicipalCounterpartPct)+99)/100)))
	}
	if f.ArtistFeesCents*100 > total*int64(lei.MaxFeesPct) {
		l.add(sec, "artistFeesCents", "lei12082_fees", SeverityBlock,
			fmt.Sprintf("Lei 12.082/2023: os cachês podem usar no máximo %d%% do repasse (%s).", lei.MaxFeesPct, brl(total*int64(lei.MaxFeesPct)/100)))
	}
	if f.ArtistFeesCents > 0 && f.RegionalArtistFeesCents*100 < f.ArtistFeesCents*int64(lei.MinRegionalFeesPct) {
		l.add(sec, "regionalArtistFeesCents", "lei12082_regional", SeverityBlock,
			fmt.Sprintf("Lei 12.082/2023: pelo menos %d%% dos cachês devem ir para artistas regionais (%s).", lei.MinRegionalFeesPct, brl((f.ArtistFeesCents*int64(lei.MinRegionalFeesPct)+99)/100)))
	}
}
