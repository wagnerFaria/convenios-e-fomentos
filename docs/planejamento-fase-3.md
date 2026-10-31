# Planejamento detalhado — Fase 3: conferência da análise

Parte de [planejamento-fluxo-analise-enquadramento.md](planejamento-fluxo-analise-enquadramento.md) (seções 3, 5, 6 e 7) e do [BPMN](bpmn/convenios-e-fomento.bpmn). Cobre os passos "Assumir a análise" e "Analisar informações e documentos" do Analista. A indicação do departamento e o enquadramento são da Fase 4. Sem código aqui: decisões, modelo, regras, contratos, testes e ordem de trabalho.

**Estado: implementada** (commit da fase 3). D12 e D13 foram aceitas como recomendadas.

**Resultado da fase:** o Analista assume a análise, abre cada documento, aprova (ou desfaz) seção por seção e documento por documento, e o sistema sabe dizer quando "tudo está aprovado". Devolver para complementação e reenviar passam a mexer na conferência certa. Tudo por API, com OpenAPI e testes; as telas são da Fase 5.

---

## 1. O que existe e o que muda

| Peça | Hoje | Nesta fase |
|---|---|---|
| Etapa "Em análise" | Só o **reenvio** grava `in_analysis` (`Resubmit`); nenhuma rota faz "Aguardando análise de checklist" → "Em análise" | `POST /start-analysis` faz a transição e grava o analista responsável |
| Analista responsável | Não existe | Coluna `projects.analyst_id`; usada pelo e-mail de arquivamento (decisão D2) |
| Situação de conferência por item | Não existe (o planejamento original deixou para "fase seguinte") | Tabela `project_item_reviews` (pendente / aprovado) |
| "Abrir antes de aprovar" | O download da equipe só é auditado (`project.file-download`) | Tabela `project_file_views`: quem abriu qual arquivo; aprovar exige a abertura |
| Devolver para complementação | `RequestComplement` cria o pedido e muda a etapa | Também **zera** a aprovação dos itens devolvidos |
| Reenvio | `Resubmit` volta para "Em análise" e só os itens devolvidos mudam | Sem mudança de código: os itens devolvidos já estão pendentes e os demais seguem aprovados |
| Visão do projeto (`GET /projects/{id}`) | Mesma para todos | A equipe recebe também `analyst` e `review` (o titular e os membros não) |
| Ações de permissão | `read`, `request-complement`, `archive`, `deny`, ... | Novas: `start-analysis` e `review-item`, só do Analista |

---

## 2. Decisões

### 2.1 Já tomadas
1. O Analista analisa **cada documento separadamente** e **abre o documento antes de aprová-lo**; analisa também todos os grupos de informações.
2. Reenvio: itens já aprovados e não alterados continuam aprovados.
3. O e-mail de arquivamento vai ao "Analista" responsável quando houver; sem responsável, a todos (D2).
4. O Analista só segue para a indicação do departamento com tudo aprovado.

### 2.2 Novas, com recomendação adotada

| # | Decisão | Recomendação | Alternativa |
|---|---|---|---|
| D12 | **Documentos adicionais** (até 5 por projeto, título + descrição + um arquivo) entram na conferência? | **Sim**, como itens conferíveis (`kind=extra`, alvo = id do documento), com a mesma regra de abrir antes de aprovar. Como o arquivo de um adicional já enviado fica travado, ele **não pode ser devolvido individualmente**; se houver problema, o Analista usa a mensagem do pedido de complementação | Deixar de fora: o Analista não os conferiria |
| D13 | **Quem pode agir**: só o analista responsável, ou qualquer Analista? | **Qualquer Analista** com a permissão. O responsável é informativo (e-mails, tela); quem assume primeiro vira o responsável. Evita projeto parado se o responsável estiver ausente | Só o responsável (exigiria função de reatribuir, que não existe) |
| D14 | O que vale como "abriu o documento"? | Ter **pedido o link de download** pelo sistema (o servidor registra). Não dá para saber se o navegador exibiu o conteúdo | Exigir confirmação manual |
| D15 | Quais itens são conferíveis? | As **5 seções** devolvíveis; os documentos do checklist com **arquivo confirmado ou justificativa aceita**; os adicionais com arquivo. Documento opcional vazio é ignorado. **Obrigatório sem arquivo nem justificativa** é *impedimento*: não há o que aprovar, e o Analista devolve ou arquiva | Exigir aprovação até do opcional vazio |
| D16 | Reatribuir a análise a outro Analista | **Fora desta fase** (limitação conhecida). Qualquer Analista continua podendo agir (D13) | Criar a função agora |
| D17 | Itens fora de ordem: aprovar antes de assumir? | **Não.** Aprovar e desfazer exigem a etapa "Em análise" (409 `analysis_not_started` antes disso). Devolver e arquivar continuam valendo também em "Aguardando análise de checklist", como hoje | Assumir implicitamente ao aprovar o primeiro item |

D12 e D13 mudam o comportamento visível; peço sua confirmação (seção 12).

---

## 3. Modelo de dados — migração `00026_project_review.sql`

(A caixa de saída ficou com a 00025. A migração do enquadramento passa a ser a **00027**.)

- `projects.analyst_id uuid REFERENCES admins (id) ON DELETE SET NULL`. Nulo até alguém assumir.
- `project_item_reviews`:
  - `project_id` (FK, `ON DELETE CASCADE`), `kind` (`section` | `document` | `extra`), `target` (nome da seção; chave do item `codigo` ou `codigo@id-da-emenda`; id do documento adicional), `status` (`pending` | `approved`), `reviewed_by_id` (FK `admins`, `SET NULL`), `reviewed_at`.
  - Única por `(project_id, kind, target)`. **Ausência de linha = pendente**: nada é criado no envio; só aprovar e devolver escrevem.
- `project_file_views`:
  - `file_id` (FK `project_files`, `CASCADE`), `admin_id` (FK `admins`, `CASCADE`), `opened_at`. Chave `(file_id, admin_id)`; repetir a abertura não muda nada (`ON CONFLICT DO NOTHING`).
- Índice em `project_item_reviews (project_id)`.
- `Down` remove as duas tabelas e a coluna.

A etapa nova (`awaiting_opinion`) e o `framing_status` **não** entram aqui; são da Fase 4.

---

## 4. Regras

### 4.1 Assumir a análise (`start-analysis`)
- Exige: `submitted` e etapa `awaiting_checklist`. Efeito: etapa → `in_analysis`, `analyst_id` = quem assumiu.
- Chamar de novo, em `in_analysis`, pelo mesmo Analista → 200 sem mudar nada (idempotente). Por outro Analista → 200 também, **sem trocar o responsável** (D13/D16): quem entra depois só passa a trabalhar. Em `awaiting_proponent` → 409 `project_stage_invalid`.
- Trava a linha do projeto (`GetProjectForUpdate`) como as demais transições; duas pessoas assumindo ao mesmo tempo: a primeira vence e a outra recebe o projeto já assumido.

### 4.2 Itens conferíveis
Calculados **no servidor** a cada leitura e a cada aprovação, a partir do estado atual (seções + checklist + adicionais), nunca de uma lista gravada:

| Item | Alvo | Conferível quando | Impedimento |
|---|---|---|---|
| Seção | `proponent`, `institution`, `proposal`, `funding`, `document-options` | sempre | — |
| Documento do checklist | chave do item | situação "feito" (arquivo confirmado) ou "justificado" | obrigatório sem arquivo nem justificativa |
| Documento adicional | id | tem arquivo confirmado | — |

Itens opcionais vazios não aparecem. `allApproved` = há pelo menos um item, todos aprovados e **nenhum impedimento**.

### 4.3 Aprovar e desfazer (`review-item`)
- Exige `submitted`, etapa `in_analysis`. O item precisa existir e ser conferível agora (senão 422).
- **Aprovar documento com arquivo** (checklist ou adicional) exige que **o próprio Analista** tenha aberto **cada arquivo confirmado atual** do item. Faltando, 422 `file_not_opened` com a lista dos arquivos que faltam. Item só com justificativa e seções não exigem abertura.
- Desfazer (`approved=false`) vale a qualquer momento enquanto em `in_analysis`.
- Aprovar de novo um item já aprovado é idempotente.
- Grava `reviewed_by_id` e `reviewed_at` e a auditoria `project.review-item` (`kind`, `target`, `approved`; **sem** texto livre).

### 4.4 Abrir o documento
- No `Download` (rota que já existe), se quem baixa é da **equipe** (não é titular nem membro: `accessOf` vazio), grava em `project_file_views`. O titular e os membros baixando os próprios arquivos não geram registro.
- O registro é feito **antes** de devolver o link, na mesma operação; se falhar, o download falha (melhor do que aprovar sem prova).
- Arquivo novo (id novo) exige nova abertura. O mesmo arquivo já aberto não exige de novo.

### 4.5 Devolver e reenviar
- `RequestComplement`: na mesma transação, os itens devolvidos viram **pendentes** (linha com `status=pending`, sem quem aprovou). Os demais mantêm a situação.
- `Resubmit`: nada novo. Os devolvidos seguem pendentes até o Analista reaprovar; os demais continuam aprovados porque o proponente só consegue mexer no que foi devolvido.
- Documento adicional **novo** criado depois do envio aparece como pendente e impede `allApproved` até ser conferido. Isso é desejado.

### 4.6 O que fica bloqueado
Depois da Fase 4, com enquadramento pendente, não se aprova nem se desfaz (409 `framing_pending`). Nesta fase essa regra ainda não existe, mas o ponto de checagem fica no mesmo lugar para ser ligado.

---

## 5. API (prefixo `/api/admin/projects/{id}`)

| Rota | Permissão | Corpo | Resposta |
|---|---|---|---|
| `POST /start-analysis` | `admin::projects.start-analysis` | — | Projeto (visão da equipe) |
| `PUT /review/items` | `admin::projects.review-item` | `{kind, target, approved}` | Projeto (visão da equipe, com `review` atualizado) |

Erros novos (código estável em inglês, mensagem em português): `analysis_not_started` (409), `project_stage_invalid` (409, já existe), `item_not_reviewable` (422), `file_not_opened` (422, `details.files = [{id, filename}]`), `project_not_submitted` (409, já existe).

**Visão do projeto para a equipe** (`access` vazio) ganha:

```
analyst: { id, name } | null
review: {
  items: [ { kind, target, label, status: "pending"|"approved", blocking: bool, reason?,
             reviewedBy?: {name}, reviewedAt?,
             files?: [ { id, filename, opened: bool } ] } ],
  summary: { total, approved, pending, blockers },
  allApproved: bool
}
```

`opened` é **do administrador que consulta** (cada Analista vê o que ele mesmo abriu), porque a regra de aprovar é por pessoa. `reason` explica o impedimento ("Documento obrigatório sem arquivo nem justificativa"). Titular e membros não recebem `review` nem `analyst` (a conferência é interna).

OpenAPI: as duas rotas, os erros e os campos novos, na mesma entrega (existe teste que falha se uma rota não estiver documentada).

---

## 6. Permissões e perfis

- `project.Register`: `ActionStartAnalysis = "admin::projects.start-analysis"` e `ActionReviewItem = "admin::projects.review-item"`, ambas com suporte a condições.
- Migração 00026 concede as duas ao **Analista**, com a condição `admin::project-submitted` (como as demais ações dele). Nenhum outro perfil as recebe.
- Seed espelha; `TestPerfisDeEquipeDaMigracaoBatemComOSeed` e `TestQuemPodeDecidirEmProjetos` cobrem. A contagem do seed sobe de 54 para 56 permissões.
- A descrição do Analista já fala em "assume a análise ... confere documentos e informações".

---

## 7. E-mail de arquivamento ao Analista responsável (D2)

Na fase 2 o aviso ia a todos os Analistas. Agora: se `analyst_id` está preenchido e o administrador está ativo e não bloqueado, **só ele** recebe; sem responsável (ou responsável inativo), todos os Analistas ativos, como hoje. `notifyArchived` já recebe o departamento; passa a receber também o responsável (a linha do projeto traz `analyst_id`).

---

## 8. Passos de trabalho (um commit por passo; `make lint` e testes verdes antes de cada um)

1. **P3.1 Migração e consultas.** `00026_project_review.sql`; `queries/project_review.sql` (assumir analista, listar/gravar/zerar revisões de itens, registrar e listar aberturas); `make sqlc`.
2. **P3.2 Regras puras.** `internal/project/review.go`: monta os itens conferíveis a partir da seção + checklist + adicionais, calcula impedimentos e `allApproved`. Sem banco, testável por unidade.
3. **P3.3 Ações.** `StartAnalysis` e `ReviewItem` no serviço (transação, trava de linha, auditoria); `ActionStartAnalysis`/`ActionReviewItem` registradas; handlers e rotas.
4. **P3.4 Abertura.** `Download` registra a abertura da equipe.
5. **P3.5 Visão.** `analyst` e `review` na visão da equipe, com `opened` por administrador.
6. **P3.6 Devolução.** `RequestComplement` zera os itens devolvidos na mesma transação.
7. **P3.7 Perfis.** Concessão ao Analista na migração e no seed; testes de perfis ajustados; contagem do seed.
8. **P3.8 E-mail ao responsável** (seção 7).
9. **P3.9 OpenAPI e documentação** (`backend/README.md`, planejamentos, BPMN sem mudança).

P3.2 não depende de nada e pode ser feito primeiro, com testes de unidade rápidos.

---

## 9. Testes

**Unidade (`review_test.go`, sem Docker):** itens conferíveis para combinações do checklist (arquivo, justificativa, obrigatório vazio, opcional vazio, emenda, adicional); impedimentos; `allApproved` com zero itens (falso), com impedimento (falso), com tudo aprovado (verdadeiro).

**Integração (`internal/app`, perfis semeados):**
- *Assumir:* `awaiting_checklist` → `in_analysis` com o responsável; repetir é idempotente; outro Analista não troca o responsável; duas chamadas simultâneas, uma assume; projeto em `awaiting_proponent` ou rascunho → erro; outros perfis → 403.
- *Aprovar:* sem assumir → 409; aprovar documento sem abrir → 422 `file_not_opened` com a lista; abrir e aprovar → 200; desfazer; aprovar item inexistente → 422; **abertura é por pessoa** (o que o Analista A abriu não vale para o B); arquivo novo exige nova abertura; justificativa e seção aprovam sem abertura; documento adicional segue a mesma regra.
- *Impedimento:* obrigatório sem arquivo nem justificativa não é aprovável e mantém `allApproved=false`.
- *Abertura:* download da equipe registra; download do titular não; falha ao registrar impede o download; segunda abertura não duplica.
- *Devolver e reenviar:* devolver zera só os itens devolvidos (os outros ficam aprovados); reenvio mantém os não alterados aprovados; adicional criado depois do envio fica pendente.
- *Visão:* a equipe vê `review` e `analyst`; titular e membro não vêem; `opened` muda conforme quem consulta.
- *Matriz de acesso:* as duas ações novas só funcionam para o Analista (403 para Supervisor, Secretário e Revisor) e sobre projeto enviado (403 em rascunho).
- *E-mail ao responsável:* com responsável, só ele recebe o arquivamento; sem, todos; responsável bloqueado, todos.
- *Auditoria:* cada ação gera sua entrada, sem texto livre.

**Já existentes que mudam:** contagem do seed e a tabela de `TestQuemPodeDecidirEmProjetos`; o teste de e-mail de arquivamento que conta analistas.

---

## 10. Riscos e cuidados

| Risco | Cuidado |
|---|---|
| "Aberto" só prova que o link foi pedido, não que alguém leu | D14 documentado; a auditoria de download continua |
| Aprovações antigas de um item cujo conteúdo mudou | O proponente só altera o que foi devolvido, e devolver zera; arquivos novos têm id novo e exigem nova abertura |
| Adicional não pode ser devolvido (arquivo travado) | D12 explícito; o Analista usa a mensagem do pedido |
| Dois Analistas aprovando/desfazendo o mesmo item ao mesmo tempo | Trava de linha do projeto; última gravação vence, ambas auditadas |
| Visão pesada: calcular itens e consultar aberturas a cada leitura | Só na visão da equipe; uma consulta por projeto para revisões e outra para aberturas do administrador |
| Responsável fica sem poder reatribuir | D16: limitação anotada; qualquer Analista continua podendo agir |
| Projeto com muitos arquivos gerando lista longa em `file_not_opened` | Limitar a 20 na resposta e dizer quantos faltam |

---

## 11. Definição de pronto

- [ ] Migração sobe e desce (o teste de migração existente cobre); sem diferença no `sqlc`.
- [ ] Rotas novas na API e no OpenAPI.
- [ ] Todos os testes da seção 9 verdes; suíte completa, `make lint` e `make cover-authz` ≥ 80%.
- [ ] Seed, migração e testes de perfis alinhados (56 permissões confirmadas ao rodar).
- [ ] Documentação atualizada.
- [ ] Banco de desenvolvimento migrado.

---

## 12. Perguntas

1. **D12:** os documentos adicionais entram na conferência (e portanto impedem "tudo aprovado" enquanto não forem abertos e aprovados)? Recomendo que sim.
2. **D13:** qualquer Analista pode aprovar, devolver e arquivar, ou só o responsável? Recomendo qualquer um.
