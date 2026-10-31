# Planejamento detalhado — Fase 4: indicação do departamento e enquadramento

Parte de [planejamento-fluxo-analise-enquadramento.md](planejamento-fluxo-analise-enquadramento.md) (seções 3, 4, 6, 7 e 8) e do [BPMN](bpmn/convenios-e-fomento.bpmn). Cobre, do lado do Analista, "Indicar departamento de enquadramento" e, do lado do Revisor do Departamento, "Analisar enquadramento" com os caminhos "Enquadramento correto? Sim / Não". Sem código aqui: decisões, modelo, regras, contratos, testes e ordem de trabalho. Depende das fases 1 (perfis e marcador de departamento), 2 (e-mails) e 3 (conferência, `allApproved`), todas concluídas.

**Estado: implementada.** D18 e D19 aceitas como recomendadas (a justificativa vai no e-mail de recusa; o Analista não retira indicação).

**Resultado da fase:** com tudo aprovado, o Analista indica um departamento finalístico; o projeto aparece só para os Revisores daquele departamento; o primeiro Revisor que decidir **confirma** (o projeto vai para "Aguardando parecer") ou **recusa com justificativa** (o departamento é limpo e o projeto volta ao Analista, sem limite de recusas). Os Supervisores recebem os dois e-mails do BPMN. As filas para as telas saem por filtros da listagem. Tudo por API, com OpenAPI e testes; as telas são da Fase 5.

---

## 1. O que existe e o que muda

| Peça | Hoje | Nesta fase |
|---|---|---|
| `projects.department_id` | Existe desde a migração 00010, mas nenhuma rota o grava | Gravado ao indicar, limpo ao recusar |
| Condição `admin::project-submitted-same-department` | Existe (fase 1); o Revisor lê os projetos do próprio departamento | Sem mudança; passa a ter o que ler quando houver `department_id` |
| `allApproved` | Calculado na visão da equipe (fase 3) | Passa a ser **exigido** para indicar o departamento |
| Etapas | `awaiting_checklist`, `in_analysis`, `awaiting_proponent` | Nova: `awaiting_opinion` ("Aguardando parecer"), quando o Revisor confirma |
| Enquadramento | Não existe | Tabela de rodadas `project_framings` e coluna `projects.framing_status` |
| Ações de permissão | Sem `assign-department` nem `review-framing` | Novas: a primeira do Analista, a segunda do Revisor (só do próprio departamento) |
| E-mails à equipe | Só o de arquivamento | Dois novos para os Supervisores: "documentação regular e encaminhada" e "enquadramento recusado" |
| Listagem | Filtros `q` e `status` | Novos: `stage`, `framing`, `department`, `analyst`; a linha ganha `department`, `framingStatus` e `analyst` |
| `GET /admin/departments` | `userCount`, `documentCount`, `isFinalistic` | Ganha `reviewerCount` e `projectCount` |
| Exclusão de departamento | Barrada por administradores e documentos | Também por projetos que o referenciam |
| Devolver / arquivar pelo Analista, aprovar / desfazer | Livres em `in_analysis` | Bloqueados com enquadramento pendente (409 `framing_pending`) |

---

## 2. Decisões

### 2.1 Já tomadas
1. Só departamento marcado como área finalística pode ser indicado, e só se tiver ao menos um Revisor ativo (422 `department_not_finalistic` e `department_without_reviewer`).
2. A indicação exige tudo aprovado.
3. A primeira decisão de qualquer Revisor do departamento vale. Não há limite de recusas.
4. A recusa exige justificativa, limpa o departamento (a história fica nas rodadas) e devolve o projeto ao Analista.
5. Confirmar leva o projeto a `awaiting_opinion` (D1).
6. Com enquadramento pendente, o Analista não devolve nem arquiva, e ninguém aprova nem desfaz item (D11). O Proponente (desistência) e o Secretário (negação) continuam podendo agir.
7. Os dois e-mails aos Supervisores seguem a lista permitida de campos (D9): sem dados pessoais.
8. Mudar o marcador `is_finalistic` de um departamento só afeta indicações novas, nunca rodadas abertas.

### 2.2 Novas, com recomendação adotada

| # | Decisão | Recomendação | Alternativa |
|---|---|---|---|
| D18 | O e-mail de **recusa** leva a justificativa do Revisor? | **Sim**, limitada a 500 caracteres, porque o BPMN diz que o Supervisor recebe "o departamento recusado e a justificativa". É texto de servidor da casa, não do proponente; o e-mail de arquivamento omite o motivo porque ele pode ter sido escrito pelo proponente (desistência) | Só o departamento, a rodada e o link, como no arquivamento |
| D19 | **Retirar a indicação** pelo Analista enquanto o enquadramento está pendente | **Não nesta fase.** O BPMN prevê só o ciclo recusa → nova indicação. Limitação conhecida: se a indicação foi errada e ninguém do departamento decide, o projeto fica esperando (o Secretário ainda pode negar) | Criar `DELETE /department` que encerra a rodada como "retirada" |
| D20 | O Analista é avisado por e-mail quando o Revisor recusa? | **Não.** O BPMN envia só ao Supervisor; o Analista vê o projeto de volta na fila dele, com o histórico e a justificativa. Fácil de acrescentar depois | Avisar o Analista responsável |
| D21 | Confirmar o enquadramento envia e-mail? | **Não** (o BPMN termina no evento "Projeto enquadrado") | Avisar Analista e Supervisor |
| D22 | Segunda decisão depois que outro Revisor já decidiu | A permissão é avaliada contra o departamento da **última rodada** quando o projeto já não tem departamento. Assim a segunda pessoa recebe **409 `framing_already_decided`** em qualquer dos dois casos (confirmado ou recusado), em vez de um 403 confuso | Aceitar o 403 na recusa |
| D23 | `version` no corpo de `PUT /department` | **Não.** O servidor revalida tudo sob a trava da linha (etapa, `allApproved`, enquadramento já pendente); o `round` nas decisões do Revisor cumpre o papel de evitar agir sobre rodada antiga | Exigir `version` como nas edições do proponente |

D18 e D19 são as que peço confirmação (seção 11).

---

## 3. Modelo de dados — migração `00027_project_framing.sql`

- `projects.framing_status text` com `CHECK` (`pending` | `confirmed`), nulo por padrão. Nulo = sem enquadramento em andamento (inclusive depois de uma recusa).
- Troca do `CHECK` de `stage` para incluir `awaiting_opinion` (mesmo padrão da 00014: derrubar e recriar a constraint).
- `CHECK` de consistência: `framing_status` preenchido exige `department_id` preenchido. Defesa em profundidade: se um departamento fosse excluído por fora do serviço, o banco recusaria em vez de deixar o projeto sem área.
- `project_framings` (uma linha por rodada):
  - `id`, `project_id` (FK, `CASCADE`), `round` (inteiro, único por projeto), `department_id` (FK `SET NULL`) e `department_name` (cópia do nome na data, para o histórico sobreviver a renomeações e exclusões);
  - `indicated_by_id`, `indicated_at`;
  - `decision` (`pending` | `confirmed` | `refused`, padrão `pending`), `decided_by_id`, `decided_at`, `justification`;
  - `CHECK`: `decision = 'pending'` se e somente se `decided_at` é nulo; recusa exige `justification` não vazia;
  - índice único parcial: **no máximo uma rodada `pending` por projeto**.
- Concede as duas ações novas (seção 5) e atualiza a descrição do Revisor onde ainda é a da fase 1.
- `Down`: remove tabela, colunas e constraints; devolve o `CHECK` de `stage` ao conjunto anterior (falha se existir projeto em `awaiting_opinion`, o que é o comportamento correto: não descer com dado que não cabe); retira as permissões; avisa o cache de permissões (`pg_notify`).

---

## 4. Regras

### 4.1 Indicar o departamento (`PUT /department`)
Sob a trava da linha do projeto:
1. Permissão e condição de visibilidade (como nas demais ações da equipe).
2. `submitted`, etapa `in_analysis` (senão 409 `analysis_not_started` / `project_stage_invalid`).
3. Sem enquadramento pendente (409 `framing_pending`).
4. **Tudo aprovado**: a conferência é recalculada no servidor naquele instante; se não, 409 `review_incomplete`, com o número de pendências e impedimentos.
5. Departamento existe, é finalístico (422 `department_not_finalistic`) e tem ao menos um Revisor **ativo e não bloqueado** (422 `department_without_reviewer`). A linha do departamento é lida com trava compartilhada, para uma exclusão simultânea esperar.
6. Efeitos, na mesma transação: abre a rodada `n+1` (`pending`, com o nome do departamento), grava `department_id` e `framing_status = pending`, auditoria `project.assign-department`, e grava na caixa de saída o e-mail "documentação regular" aos Supervisores (seção 7). A etapa continua `in_analysis`.

### 4.2 Confirmar (`POST /framing/confirm`) e recusar (`POST /framing/refuse`)
Corpo: `{round}` (e `justification` na recusa, de 3 a 1.000 caracteres). Sob a trava da linha:
1. Permissão `review-framing`, avaliada contra o departamento do projeto ou, se já foi limpo, o da última rodada (D22).
2. `submitted`. Sem rodada `pending`: 409 `framing_already_decided`. Rodada `pending` com outro número: 409 `framing_round_stale`.
3. **Confirmar:** a rodada vira `confirmed`; `framing_status = confirmed`; etapa `awaiting_opinion`; o departamento fica gravado; auditoria `project.framing-confirm`. Sem e-mail (D21).
4. **Recusar:** a rodada vira `refused` com a justificativa; `department_id = NULL`; `framing_status = NULL`; a etapa continua `in_analysis`; a conferência **não** é tocada (tudo continua aprovado, então o Analista pode indicar outro departamento direto, ou desfazer/devolver); auditoria `project.framing-refuse`; e-mail "enquadramento recusado" aos Supervisores.

Duas decisões simultâneas: a trava da linha serializa; a segunda vê a rodada já decidida e recebe 409.

### 4.3 O que fica bloqueado com enquadramento pendente (D11)
- `RequestComplement` e o arquivamento pelo Analista (`Archive`): 409 `framing_pending`.
- `ReviewItem` (aprovar/desfazer): 409 `framing_pending`.
- Continuam valendo: desistência (`withdraw`), negação (`deny`), prazo vencido (não se aplica: só existe em "Aguardando proponente"), e as leituras.
- Em `awaiting_opinion`: devolver, aprovar/desfazer e indicar dão 409 `project_stage_invalid` com mensagem própria ("O enquadramento já foi confirmado."); desistir, negar e arquivar continuam possíveis.

### 4.4 Arquivar durante o enquadramento
Os e-mails de arquivamento (fase 2) já avisam os Revisores do departamento do projeto. Com a rodada pendente, o projeto arquivado mantém `department_id`, então os Revisores continuam sendo avisados; a rodada pendente fica como está (o projeto está arquivado e somente leitura).

### 4.5 Excluir departamento
`department_in_use` passa a contar também projetos com `department_id` igual ao departamento (`projectCount`). Rodadas antigas não bloqueiam (têm o nome copiado).

---

## 5. Permissões e perfis

- Ações novas: `admin::projects.assign-department` e `admin::projects.review-framing`, registradas em `project.Register`.
- A migração 00027 concede: **Analista** → `assign-department` com `admin::project-submitted`; **Revisor do Departamento** → `review-framing` com `admin::project-submitted-same-department`. Ninguém mais.
- Descrição do Revisor: "Vê os projetos enviados do próprio departamento e confirma ou recusa o enquadramento." (atualizada só onde ainda é a da fase 1).
- Seed espelha; os testes de perfis e de contagem (56 → 58 permissões) ajustam.

---

## 6. API (prefixo `/api/admin/projects/{id}`)

| Rota | Permissão | Corpo | Resposta |
|---|---|---|---|
| `PUT /department` | `assign-department` | `{departmentId}` | Projeto (visão da equipe) |
| `POST /framing/confirm` | `review-framing` | `{round}` | Projeto |
| `POST /framing/refuse` | `review-framing` | `{round, justification}` | Projeto |

Erros novos (códigos estáveis em inglês): `framing_pending` (409), `review_incomplete` (409, com `details.pending` e `details.blockers`), `department_not_finalistic` (422), `department_without_reviewer` (422), `framing_already_decided` (409), `framing_round_stale` (409). Reaproveitados: `analysis_not_started`, `project_stage_invalid`, `project_not_submitted`.

**Visão do projeto (equipe)** ganha:

```
framing: { round, status: "pending"|"confirmed", department: {id,name},
           indicatedBy, indicatedAt, decidedBy?, decidedAt? } | ausente
framings: [ { round, department: {id,name}, decision, indicatedBy, indicatedAt,
              decidedBy?, decidedAt?, justification? } ]   // histórico, da mais antiga à mais nova
```

**Listagem (`GET /admin/projects`)**: filtros novos, combinados com o filtro de permissão como os atuais: `stage` (um ou vários separados por vírgula: `awaiting_checklist`, `in_analysis`, `awaiting_proponent`, `awaiting_opinion`), `framing` (`pending`, `confirmed` ou `none`), `department` (id) e `analyst` (id). A linha (`Summary`) ganha `department {id,name}`, `framingStatus` e `analyst {id,name}`. Valor inválido: 422 com o campo.

Filas que a Fase 5 usará: Analista → `stage=awaiting_checklist,in_analysis`; Revisor → `framing=pending` (o filtro de permissão já limita ao departamento dele); Secretário e Supervisor → sem filtro.

**Departamentos (`GET /admin/departments`)**: ganha `reviewerCount` (Revisores ativos) e `projectCount` (projetos que o referenciam).

OpenAPI na mesma entrega (existe teste que falha com rota sem documentação).

---

## 7. E-mails aos Supervisores

Gravados na caixa de saída na transação da ação (fase 2), uma mensagem por Supervisor ativo e não bloqueado, com chave de deduplicação `framing-indicated:{projeto}:{rodada}:{destinatário}` e `framing-refused:{projeto}:{rodada}:{destinatário}`.

| E-mail | Quando | Conteúdo (lista permitida, D9) |
|---|---|---|
| Documentação regular e encaminhada | Ao indicar o departamento | Código, instrumento, título, objeto, nº SIGCon, vigência, tipo de recurso e valores (total, emenda, contrapartida), municípios de execução, instituição (nome, CNPJ, tipo, natureza, UF), proponente (nome e cargo), **departamento indicado**, número da rodada e o link. Sem CPF, telefones, e-mails de contato, endereços, dados do assinante, justificativa nem anexos |
| Enquadramento recusado | Ao recusar | Código, título, instituição, departamento recusado, rodada, **justificativa (até 500 caracteres, D18)** e o link |

Um teste procura no corpo qualquer padrão de CPF, telefone ou e-mail pessoal, como nos e-mails de arquivamento.

---

## 8. Passos de trabalho (um commit por passo; lint e testes verdes antes de cada um)

1. **P4.1 Migração e consultas.** `00027_project_framing.sql`; `queries/project_framing.sql` (rodadas, atualização do projeto, contagens, trava compartilhada do departamento); `make sqlc`.
2. **P4.2 Perfis.** Ações registradas; concessões na migração e no seed; descrição do Revisor; testes de perfis e contagem.
3. **P4.3 Regras de enquadramento.** `AssignDepartment`, `ConfirmFraming`, `RefuseFraming` no serviço, com trava de linha, auditoria e a regra D22; handlers e rotas.
4. **P4.4 Bloqueios.** `framing_pending` em `RequestComplement`, `Archive` (Analista) e `ReviewItem`; mensagens de etapa para `awaiting_opinion`.
5. **P4.5 Visão e listagem.** `framing` e `framings` na visão; filtros e campos novos na listagem; `reviewerCount` e `projectCount` nos departamentos; exclusão de departamento considera projetos.
6. **P4.6 E-mails** aos Supervisores (seção 7).
7. **P4.7 OpenAPI e documentação** (`backend/README.md`, planejamentos).
8. **Banco de desenvolvimento:** aplicar a migração 00027.

P4.3 e P4.5 dependem de P4.1; P4.4 e P4.6 dependem de P4.3.

---

## 9. Testes

**Integração (`internal/app`, perfis semeados, departamentos Financeiro e Jurídico finalísticos com um Revisor cada):**

- *Indicar:* sem assumir (409); com itens pendentes (409 `review_incomplete` com contagens); com impedimento (409); para departamento não finalístico (422); sem Revisor ativo, inclusive com o único Revisor bloqueado (422); departamento inexistente; sucesso grava `department_id`, `framing_status=pending`, abre a rodada 1 com o nome e envia o e-mail aos Supervisores; indicar de novo com rodada pendente (409); permissões (só o Analista; 403 para Supervisor, Secretário, Revisor e Proponente; 403 em rascunho).
- *Visibilidade:* depois de indicar, o Revisor de Financeiro lista e abre o projeto; o de Jurídico e o sem departamento não (404/403); o Analista segue vendo.
- *Confirmar:* leva a `awaiting_opinion`, mantém o departamento, não envia e-mail; o Revisor do outro departamento não consegue (403); `round` errado (409 `framing_round_stale`).
- *Recusar:* justificativa obrigatória (422); limpa `department_id` e `framing_status`; a conferência continua toda aprovada; o projeto some da fila do Revisor; o histórico registra a rodada com a justificativa; e-mail aos Supervisores com a justificativa truncada; sem e-mail ao Analista.
- *Ciclo sem limite:* recusar e indicar de novo 5 vezes seguidas, alternando departamentos; rodadas numeradas 1 a 5 em ordem; só a última pode estar `pending`.
- *Concorrência:* dois Revisores confirmando ao mesmo tempo (um vence, o outro 409); um confirma e outro recusa (idem); um recusa e o outro confirma logo depois (o segundo recebe 409 `framing_already_decided`, não 403 — D22); duas indicações simultâneas (uma vence); desistência/negação concorrendo com a decisão do Revisor (quem travar primeiro vence; o outro vê `project_archived`).
- *Bloqueios com pendência:* devolver (409), arquivar pelo Analista (409), aprovar e desfazer (409); desistência e negação funcionam; arquivar durante a pendência avisa o Revisor do departamento.
- *Depois de confirmado:* devolver, aprovar/desfazer e indicar dão 409 `project_stage_invalid` com a mensagem certa; desistir e negar funcionam.
- *Listagem:* cada filtro e as combinações; valor inválido (422); o Revisor só enxerga o que é do departamento mesmo sem filtro; `Summary` com `department`, `framingStatus` e `analyst`.
- *Departamentos:* `reviewerCount` e `projectCount`; excluir com projeto vinculado (409 `department_in_use`); renomear não altera o nome das rodadas antigas; desmarcar `is_finalistic` não afeta rodada aberta.
- *E-mails:* destinatários (só Supervisores ativos), deduplicação, ausência de dados pessoais, falha do SMTP não derruba a ação.
- *Auditoria:* uma entrada por ação, sem texto livre (a justificativa fica na rodada).
- *Migração:* o teste de migração existente cobre subir/descer; teste com dados para o `Down` falhar com projeto em `awaiting_opinion`.

**Já existentes que mudam:** contagem do seed (58), tabela de `TestQuemPodeDecidirEmProjetos`, o teste que usa `setProjectDepartment` por SQL (continua válido), e os que checam a lista de condições/ações.

---

## 10. Riscos e cuidados

| Risco | Cuidado |
|---|---|
| Projeto esperando um departamento que não decide | D19 (limitação conhecida); o Secretário pode negar; fila filtrável por `framing=pending` e idade para a Fase 5 destacar |
| Revisor único do departamento bloqueado depois da indicação | A indicação checa na hora; depois disso o projeto espera. Mesma saída de D19 |
| Departamento excluído ou desmarcado durante uma rodada | Trava compartilhada na indicação; exclusão barrada por projetos; `CHECK` no banco; desmarcar só vale para novas |
| Justificativa com dado pessoal indo por e-mail | D18: texto de servidor, 500 caracteres, só para Supervisores; reavaliar se virar problema |
| 403 confuso para o segundo Revisor | D22 |
| Estado inconsistente se a transação falhar no meio | Tudo (rodada, projeto, e-mail, auditoria) numa transação |
| Etapa nova quebrar telas e rótulos | `awaiting_opinion` entra no frontend na Fase 5; até lá o campo `stage` já chega e os rótulos desconhecidos mostram o código |
| Muitas consultas na listagem (departamento, analista) | Junções simples por chave; índices existentes em `department_id`; `analyst_id` ganha índice se a medição pedir |

---

## 11. Perguntas

1. **D18:** o e-mail de recusa ao Supervisor leva a justificativa do Revisor (até 500 caracteres)? Recomendo que sim.
2. **D19:** o Analista pode **retirar** uma indicação pendente (por exemplo, ao perceber que escolheu o departamento errado)? Recomendo **não nesta fase**, seguindo o BPMN; hoje só a recusa do Revisor desfaz a indicação.

---

## 12. Definição de pronto

- [ ] Migração sobe e desce; `sqlc` sem diferença; banco de desenvolvimento migrado.
- [ ] Rotas, erros, filtros e campos novos no OpenAPI.
- [ ] Todos os testes da seção 9 verdes; suíte completa, `make lint` e `make cover-authz` ≥ 80%.
- [ ] Seed, migração e testes de perfis alinhados (58 permissões confirmadas ao rodar).
- [ ] Fluxo do BPMN executável por API de ponta a ponta: envio → assumir → aprovar → indicar → recusar → indicar de novo → confirmar.
- [ ] Documentação atualizada.
