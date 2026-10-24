package project

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/wagner/convenios-e-fomentos/backend/internal/database/db"
)

// Clause é uma condição de aplicação de uma exigência: o fato do projeto precisa estar entre os valores. Todas as cláusulas
// de uma exigência precisam valer (E lógico).
type Clause struct {
	Field string   `json:"field"`
	In    []string `json:"in"`
}

// Template é o modelo de documento que a SEDEC publica para a exigência.
type Template struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// Escopo de uma exigência.
const (
	ScopeProject   = "project"
	ScopeAmendment = "amendment"
)

// Campos que as condições podem usar (ver facts).
var factFields = map[string]bool{
	"instrument": true, "entityNature": true, "entityTypeCode": true, "fundingType": true, "hasAmendments": true,
	"objectNature": true, "hasArtisticHiring": true, "calamity": true, "tenderStatus": true, "governorWaiver": true,
	"exclusiveSupplier": true, "publicSpaceUse": true, "reducedDeadline": true,
}

// facts são as respostas do projeto que decidem quais exigências se aplicam.
func facts(p Project) map[string]string {
	b := func(v bool) string {
		if v {
			return "true"
		}
		return "false"
	}
	return map[string]string{
		"instrument": p.Institution.Instrument, "entityNature": p.Institution.Nature, "entityTypeCode": p.Institution.EntityTypeCode,
		"fundingType": p.Funding.Type, "hasAmendments": b(p.Funding.Type == FundingAmendment || p.Funding.Type == FundingMixed),
		"objectNature": p.Proposal.ObjectNature, "hasArtisticHiring": b(p.Proposal.HasArtisticHiring), "calamity": b(p.Proposal.Calamity),
		"tenderStatus": p.Funding.TenderStatus, "governorWaiver": b(p.Funding.GovernorWaiver),
		"exclusiveSupplier": b(p.DocumentOptions.ExclusiveSupplier), "publicSpaceUse": b(p.DocumentOptions.PublicSpaceUse),
		"reducedDeadline": b(p.DocumentOptions.ReducedDeadline),
	}
}

func holds(conds []Clause, f map[string]string) bool {
	for _, c := range conds {
		v, known := f[c.Field]
		if !known {
			return false
		}
		match := false
		for _, want := range c.In {
			match = match || want == v
		}
		if !match {
			return false
		}
	}
	return true
}

// reqDef é a definição de uma exigência na carga inicial.
type reqDef struct {
	code, name, desc, grp string
	instruments           []string
	conds                 []Clause
	scope                 string
	noJustification       bool
	optional              bool // não bloqueia a continuidade nem o envio; o padrão é obrigatório
	sources               map[string]string
	templates             map[string]Template
}

const sedecDocs = "https://www.sedec.mt.gov.br/documents/d/sedec/"

func tpl(name, slug string) Template { return Template{Name: name, URL: sedecDocs + slug} }

func is(field string, in ...string) Clause { return Clause{Field: field, In: in} }

var (
	both = []string{InstrumentConvenio, InstrumentFomento}
	conv = []string{InstrumentConvenio}
	fom  = []string{InstrumentFomento}
	yes  = "true"
	no   = "false"
	src  = func(c, f string) map[string]string {
		return map[string]string{InstrumentConvenio: c, InstrumentFomento: f}
	}
	srcC  = func(c string) map[string]string { return map[string]string{InstrumentConvenio: c} }
	srcF  = func(f string) map[string]string { return map[string]string{InstrumentFomento: f} }
	tplCF = func(c, f Template) map[string]Template {
		return map[string]Template{InstrumentConvenio: c, InstrumentFomento: f}
	}
	tplC   = func(c Template) map[string]Template { return map[string]Template{InstrumentConvenio: c} }
	tplF   = func(f Template) map[string]Template { return map[string]Template{InstrumentFomento: f} }
	gOffic = "Ofícios e requerimento"
	gPlan  = "Plano de trabalho e habilitação"
	gPrice = "Pesquisa de preços"
	gDecl  = "Declarações"
	gInst  = "Instituição e dirigente"
	gWork  = "Obras e serviços de engenharia"
	gSpec  = "Situações específicas"
)

// defaultRequirements segue os checklists da SEDEC (itens "C#" do Convênio/prefeituras e "F#" do Termo de Fomento/OSC), a
// IN Conjunta 001/2015 (art. 4º e 8º), a IN Conjunta 001/2016 (art. 7º, 28 e 29) e as regras desta aplicação (ofício da Casa
// Civil para os dois instrumentos). Os itens posteriores à análise técnica (conta bancária) ficam fora do escopo.
var defaultRequirements = []reqDef{
	{code: "oficio_solicitacao", name: "Ofício de solicitação ao Secretário da SEDEC", grp: gOffic, instruments: both, sources: src("C1", "F1"),
		desc: "Ofício da instituição propondo o projeto ao Secretário de Estado de Desenvolvimento Econômico."},
	{code: "requerimento_apoio", name: "Requerimento de apoio financeiro", grp: gOffic, instruments: both, sources: src("C2", "F2"),
		templates: tplCF(tpl("Anexo II", "anexo-ii-requerimento-docx-1"), tpl("Anexo II", "anexo-ii-requerimento-docx"))},
	{code: "oficio_deputado", name: "Ofício do deputado com o número e o valor da emenda", grp: gOffic, instruments: both, scope: ScopeAmendment,
		conds: []Clause{is("hasAmendments", yes)}, sources: src("C2.1", "F2.1")},
	{code: "oficio_casa_civil", name: "Ofício da Casa Civil com o número e o valor da emenda", grp: gOffic, instruments: both, scope: ScopeAmendment,
		conds: []Clause{is("hasAmendments", yes)}, sources: srcC("C2.2"), desc: "Assinado pelo Secretário-Chefe da Casa Civil."},
	{code: "chamamento_edital", name: "Edital e homologação do chamamento público", grp: gOffic, instruments: both, conds: []Clause{is("tenderStatus", TenderHeld)}},
	{code: "chamamento_justificativa", name: "Justificativa de dispensa ou inexigibilidade do chamamento público", grp: gOffic, instruments: both,
		conds: []Clause{is("tenderStatus", TenderWaived, TenderUnfeasible)}, sources: srcF("F2.2")},

	{code: "plano_trabalho_sigcon", name: "Plano de Trabalho e anexos do SIGCon", grp: gPlan, instruments: both, sources: src("C3, 3.1 a 3.6", "F3"),
		desc: "Anexos I a V impressos do SIGCon: cadastro de órgãos e dirigentes, dados do projeto, cronogramas de execução física e de desembolso, plano de aplicação, relação de equipamentos e memória de cálculo."},
	{code: "comprovante_envio_sigcon", name: "Comprovante de envio do Plano de Trabalho no SIGCon", grp: gPlan, instruments: conv, sources: srcC("C3.7")},
	{code: "certidao_habilitacao", name: "Certidão de Habilitação Plena do SIGCon", grp: gPlan, instruments: both, sources: src("C4", "F4"),
		desc: "Emitida pelo sistema depois da habilitação na SEFAZ (SIFI/MT). Precisa estar válida."},
	{code: "termo_referencia", name: "Termo de Referência", grp: gPlan, instruments: both, sources: src("C5", "F5"),
		templates: tplCF(tpl("Anexo III", "anexo-iii-termo-de-referencia-docx-1"), tpl("Anexo III", "anexo-iii-termo-de-referencia-docx"))},

	{code: "mapa_comparativo", name: "Mapa comparativo dos orçamentos", grp: gPrice, instruments: both, sources: src("C18", "F6")},
	{code: "analise_critica", name: "Análise crítica dos orçamentos", grp: gPrice, instruments: both, sources: src("C19", "F6.1")},
	{code: "pesquisa_precos", name: "Avaliação de mercado, três orçamentos ou Ata de Registro de Preços", grp: gPrice, instruments: conv, sources: srcC("C7.1"),
		desc: "As referências de preço do Termo de Fomento devem ter no máximo 60 dias."},
	{code: "declaracao_compatibilidade_precos", name: "Declaração de compatibilidade de preços", grp: gPrice, instruments: both, sources: src("C7", "F7"),
		templates: tplCF(tpl("Anexo V", "anexo-v-declaracao-de-compatibilidade-de-precos-prefeitura-docx"), tpl("Anexo IV", "anexo-iv-declaracao-de-compatibilidade-de-precos-osc-docx"))},
	{code: "exclusividade", name: "Documentação de exclusividade", grp: gPrice, instruments: both, conds: []Clause{is("exclusiveSupplier", yes)}, sources: src("C7.2", "F8"),
		desc: "Convênio: três notas fiscais e contrato de exclusividade. Termo de Fomento: documentação para contratação por exclusividade artística."},

	{code: "declaracao_nao_duplicidade", name: "Declaração de não duplicidade de objeto", grp: gDecl, instruments: both, sources: src("C8", "F10"),
		templates: tplCF(tpl("Anexo VI", "anexo-vi-declaracao-de-nao-duplicidade-de-objeto-prefeitura-docx"), tpl("Anexo V", "anexo-v-declaracao-de-nao-duplicidade-de-objeto-osc-docx"))},
	{code: "declaracao_gratuidade", name: "Declaração de gratuidade", grp: gDecl, instruments: both, sources: src("C9", "F11"),
		templates: tplCF(tpl("Anexo VII", "anexo-vii-declaracao-de-gratuidade-prefeitura-docx"), tpl("Anexo VI", "anexo-vi-declaracao-de-gratuidade-osc-docx"))},
	{code: "declaracao_publicidade", name: "Declaração de publicidade da parceria", grp: gDecl, instruments: both, sources: src("C10", "F12"),
		templates: tplCF(tpl("Anexo VIII", "anexo-viii-declaracao-de-publicidade-da-parceria-prefeitura-docx"), tpl("Anexo VII", "anexo-vii-declaracao-de-publicidade-da-parceria-osc-docx"))},
	{code: "declaracao_contrapartida", name: "Declaração de contrapartida", grp: gDecl, instruments: conv, conds: []Clause{is("calamity", no)}, sources: srcC("C6"),
		templates: tplC(tpl("Anexo IV", "anexo-iv-declaracao-de-contrapartida-prefeitura-docx")), desc: "Dispensada em calamidade pública ou situação de emergência."},
	{code: "quadro_detalhamento_despesa", name: "Quadro de Detalhamento de Despesa", grp: gDecl, instruments: conv, conds: []Clause{is("entityNature", "public")}, sources: srcC("C6.1")},
	{code: "loa_publicada", name: "Publicação da Lei Orçamentária Anual no Diário Oficial", grp: gDecl, instruments: conv, conds: []Clause{is("entityNature", "public")}, sources: srcC("C6.2")},
	{code: "nao_impedimento", name: "Declaração de não impedimento (art. 39 da Lei 13.019/2014)", grp: gDecl, instruments: fom, sources: srcF("F18"),
		templates: tplF(tpl("Anexo X", "anexo-x-declaracao-de-nao-impedimento-atendimento-aos-requisitos-art-39-da-a-lei-n-13-019-14-osc-docx-1"))},
	{code: "execucao_direta", name: "Atesto de execução direta do objeto", grp: gDecl, instruments: fom, sources: srcF("F19")},
	{code: "compatibilidade_institucional", name: "Demonstração de compatibilidade institucional da OSC", grp: gDecl, instruments: fom, sources: srcF("F20")},
	{code: "ausencia_conflito_interesses", name: "Declaração de ausência de conflito de interesses", grp: gDecl, instruments: fom, sources: srcF("F21"),
		templates: tplF(tpl("Anexo XI", "anexo-xi-declaracao-de-ausencia-de-conflito-de-interesses-osc-docx-1"))},
	{code: "declaracao_despesas_indiretas", name: "Declaração sobre despesas indiretas", grp: gDecl, instruments: fom, sources: srcF("F22"),
		templates: tplF(tpl("Anexo XII", "anexo-xii-declaracao-de-despesas-indireta-osc-docx"))},
	{code: "declaracao_remuneracao_dirigentes", name: "Declaração sobre remuneração de dirigentes", grp: gDecl, instruments: fom, sources: srcF("F23"),
		templates: tplF(tpl("Anexo XIII", "anexo-xiii-declaracao-sobre-remuneracao-de-dirigentes-osc-docx"))},
	{code: "termo_ciencia_dirigentes", name: "Termo de ciência e responsabilidade dos dirigentes", grp: gDecl, instruments: fom, sources: srcF("F24"),
		templates: tplF(tpl("Anexo XIV", "anexo-xiv-termo-de-ciencia-e-responsabilidade-dos-dirigentes-osc-docx"))},

	{code: "capacidade_tecnica", name: "Comprovante de equipe técnica e de capacidade técnica e operacional", grp: gInst, instruments: both, sources: src("C11", "F14"),
		templates: tplF(tpl("Anexo VIII (instalações e condições)", "anexo-viii-declaracao-de-que-possui-instalacoes-e-condicoes-para-celebrar-executar-e-prestar-contas-osc-docx-1")),
		desc:      "Convênio: equipe técnica com capacidade para executar e prestar contas. Termo de Fomento: capacidade técnica e operacional, com a declaração de instalações e condições."},
	{code: "experiencia_previa", name: "Comprovação de experiência prévia", grp: gInst, instruments: fom, sources: srcF("F15"),
		templates: tplF(tpl("Anexo IX", "anexo-ix-declaracao-de-experiencia-previa-osc-docx"))},
	{code: "comprovante_endereco_instituicao", name: "Comprovante de endereço da instituição", grp: gInst, instruments: both, sources: src("C12", "F17")},
	{code: "cartao_cnpj", name: "Cartão do CNPJ", grp: gInst, instruments: conv, sources: srcC("C13")},
	{code: "documento_dirigente", name: "RG e CPF do dirigente", grp: gInst, instruments: conv, sources: srcC("C14")},
	{code: "ato_nomeacao_dirigente", name: "Ato de nomeação ou posse do dirigente", grp: gInst, instruments: both, sources: src("C15", "F16")},
	{code: "diploma_prefeito", name: "Diploma de nomeação do prefeito", grp: gInst, instruments: conv, conds: []Clause{is("entityTypeCode", "municipal_government")}, sources: srcC("C16")},
	{code: "comprovante_endereco_dirigente", name: "Comprovante de endereço do dirigente", grp: gInst, instruments: conv, sources: srcC("C17")},
	{code: "estatuto_atas", name: "Estatuto, atas e documentação da diretoria", grp: gInst, instruments: fom, sources: srcF("F16")},
	{code: "comprovacao_domicilio", name: "Comprovação de domicílio", grp: gInst, instruments: fom, sources: srcF("F9")},

	{code: "projeto_basico_obra", name: "Projeto básico da obra ou serviço de engenharia, com ART/RRT", grp: gWork, instruments: both, conds: []Clause{is("objectNature", "works")}, sources: srcC("C21 a 21.12"),
		desc: "Projeto arquitetônico executivo, instalações elétricas e hidrossanitárias, estrutural, climatização (se houver), prevenção e combate a incêndio (quando exigido), SPDA ou dispensa, mapa de danos (restauração), memorial descritivo, planilha orçamentária, cronograma físico-financeiro e Anotação de Responsabilidade Técnica (CREA/CAU)."},
	{code: "licencas_ambientais", name: "Licenças ambientais ou dispensa", grp: gWork, instruments: both, conds: []Clause{is("objectNature", "works")}, sources: srcC("C23")},
	{code: "propriedade_imovel", name: "Certificado de propriedade do imóvel", grp: gWork, instruments: both, conds: []Clause{is("objectNature", "works")}, sources: srcC("C24"),
		desc: "Em área pública, a regularização pode ser concluída até o fim da execução, com declaração do chefe do Poder Executivo sobre a posse."},

	{code: "uso_espaco_publico", name: "Autorização para uso de espaço público", grp: gSpec, instruments: fom, conds: []Clause{is("publicSpaceUse", yes)}, sources: srcF("F13")},
	{code: "reducao_prazo", name: "Justificativa para redução do prazo de tramitação", grp: gSpec, instruments: both, conds: []Clause{is("reducedDeadline", yes)}, sources: srcC("C20"),
		templates: tplCF(tpl("Anexo XI", "anexo-xi-justificativa-para-reducao-do-prazo-de-tramitacao-da-proposta-osc-docx"), tpl("Anexo XVI", "anexo-xvi-justificativa-para-reducao-do-prazo-de-tramitacao-da-proposta-osc-1-docx"))},
	{code: "ato_calamidade", name: "Ato de calamidade pública ou situação de emergência", grp: gSpec, instruments: both, conds: []Clause{is("calamity", yes)}},
	{code: "autorizacao_governador", name: "Autorização expressa do Governador (Lei 12.082/2023, art. 5º)", grp: gSpec, instruments: both, conds: []Clause{is("governorWaiver", yes)}},
}

// EnsureDefaults carrega as exigências que ainda não existem no catálogo.
func (s *Service) EnsureDefaults(ctx context.Context) error {
	for i, d := range defaultRequirements {
		for _, c := range d.conds {
			if !factFields[c.Field] {
				return fmt.Errorf("exigência %q: campo de condição desconhecido %q", d.code, c.Field)
			}
		}
		scope := d.scope
		if scope == "" {
			scope = ScopeProject
		}
		conds, _ := json.Marshal(append([]Clause{}, d.conds...))
		sources, _ := json.Marshal(nonNil(d.sources))
		templates, _ := json.Marshal(d.templates)
		if d.templates == nil {
			templates = []byte("{}")
		}
		if _, err := s.q.InsertDefaultDocumentRequirement(ctx, db.InsertDefaultDocumentRequirementParams{
			Code: d.code, Name: d.name, Description: d.desc, Grp: d.grp, SortOrder: int32((i + 1) * 10), //nolint:gosec // poucas dezenas
			Instruments: d.instruments, Conditions: conds, Scope: scope, AcceptsJustification: !d.noJustification, Sources: sources, Templates: templates, Mandatory: !d.optional,
		}); err != nil {
			return fmt.Errorf("exigência %q: %w", d.code, err)
		}
	}
	return nil
}

func nonNil(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}
