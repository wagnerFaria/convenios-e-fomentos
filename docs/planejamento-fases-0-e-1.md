# Planejamento detalhado — Fases 0 e 1

Parte de [planejamento-fluxo-analise-enquadramento.md](planejamento-fluxo-analise-enquadramento.md) (seções 4.1, 4.4, 5 e 11) e do [BPMN](bpmn/convenios-e-fomento.bpmn). Aqui estão as tarefas, os arquivos, os testes e os critérios de pronto das duas primeiras fases. Não contém código.

| Fase | Objetivo | Resultado |
|---|---|---|
| **0 — Preparação** | Provar o estado atual, medir a linha de base e fechar as dúvidas técnicas que mudam o desenho da fase 1 | Relatório curto + teste de visibilidade + documentos alinhados |
| **1 — Perfis e visibilidade** | Os cinco perfis de equipe existirem com as permissões certas, a equipe nunca ver rascunho, e o departamento ter o marcador finalístico | `make test` verde com a matriz de acesso por perfil e rota |

**Fora de escopo nestas fases:** e-mails, conferência, enquadramento, novas ações (`start-analysis`, `review-item`, `assign-department`, `review-framing`), novas etapas e telas da equipe. Os perfis ficam com as ações que já existem; as novas entram nas fases 3 e 4.

---

## 1. Correção feita durante este planejamento

O plano geral previa uma condição de status somada a `same-department` no perfil Revisor. Lendo o motor, as condições de uma mesma permissão se combinam por **OU** (`permissionFilter` em `authz/engine.go`), e as permissões de perfis diferentes também. As duas juntas dariam "mesmo departamento *ou* enviado", o oposto do desejado.

**Solução adotada:** duas condições compostas, cada uma já com o E dentro (seção 4.4 do plano geral, atualizada):

- `admin::project-submitted`: status em `submitted` ou `archived`.
- `admin::project-submitted-same-department`: o mesmo status **e** departamento do projeto igual ao do usuário; sem departamento no usuário, nega.

Consequência para os testes: os testes atuais criam a equipe com perfis ad hoc (`p.staff`, `p.analyst()` em `project_lifecycle_test.go`), **sem nenhuma condição**. Eles nunca detectariam a exposição de rascunhos. O teste novo precisa usar os perfis **semeados** (`analyst`, `administrative-supervisor` etc.).

---

## 2. Fase 0 — Preparação

Ambiente: `docker compose` ligado (os testes de integração sobem Postgres via testcontainers).

### T0.1 Linha de base
Rodar e anotar o resultado **antes** de mudar qualquer coisa, para qualquer falha nova ser atribuível:
`make test`, `make lint`, `make cover-authz`, e no frontend `bun run typecheck`, `lint`, `format:check`. Registrar também o número de perfis e permissões do seed (hoje o teste espera `[7, 47, 2, 5, 1, 11, 2, 2]`).
**Pronto quando:** tudo verde, ou as falhas existentes listadas.

### T0.2 Teste de visibilidade da equipe (o achado principal)
Arquivo novo: `backend/internal/app/project_staff_visibility_test.go`, usando `newProjEnv`.

Cenário: a Maria tem um **rascunho** e um projeto **enviado** (`p.submitted(...)`). Criar administradores com os perfis semeados via `p.newAdmin(p.boss, email, "<código>")`: `analyst`, `administrative-supervisor` e `department-reviewer` (os dois primeiros e o terceiro já existem pelas migrations 00022 e 00023).

Verificar, para cada perfil e para o id do rascunho:
- `GET /api/admin/projects` não contém o rascunho;
- `GET /projects/{id}` devolve 404;
- idem para `/documents` (checklist), `/documents/files/{fileId}/download`, `/receipt`, `/declarations/{code}` e `/members`.

E, para o projeto enviado, as mesmas rotas devolvem 200 (controle, para o teste não passar só porque a equipe não vê nada).

**Resultado esperado hoje:** falha nos primeiros itens. Registrar a saída. Se **passar**, o achado é falso e o plano perde a condição de status; se falhar, o teste fica na branch e só entra na `main` junto com a correção da fase 1.
**Pronto quando:** o resultado está registrado no relatório da fase 0 (T0.7).

### T0.3 Spikes técnicos (cada um com resposta escrita)

| # | Pergunta | Como verificar | Por que importa |
|---|---|---|---|
| S1 | O `Schema` do projeto permite filtrar `status` e `department_id` na listagem SQL (`Schema.Column`)? | Ler `project.Schema` (model.go) e `authz/schema.go`; teste unitário de um filtro `And(In(status,...), Eq(department_id,...))` virando SQL | As condições novas dependem disso |
| S2 | O cache de permissões enxerga mudanças feitas por migration SQL? | **Respondido na leitura do código:** `authz/cache.go` guarda as permissões por perfil **sem prazo de validade**; só `invalidate` as descarta. `PGBroadcaster` assina o canal `authz_invalidate`, e um aviso com lista vazia (`[]`) invalida todos os perfis. Falta só confirmar com um teste (ver P1.4, item 8) | Sem invalidação, instâncias em execução continuariam com as permissões antigas até reiniciar |
| S3 | A tela de departamentos funciona para quem tem só leitura, edição e exclusão (sem criação)? | Ver `app/dashboard/administracao/departamentos/page.tsx`, `departments-list.tsx` e `department-dialog.tsx`: quais botões usam `can(...)` e se a rota tem guarda própria | O Analista passa a ter acesso intencional à tela; o menu (`departments.read`) não muda, mas "Novo departamento" deve ficar escondido para ele |
| S4 | Que testes dependem dos perfis que vão mudar? | Já mapeado (abaixo); confirmar rodando | Evitar surpresa na fase 1 |
| S5 | Como as migrations são testadas (subir/descer)? | Ver `internal/database` e testes que usam `migrate:down` | Definir o teste de migration da 00024 |
| S6 | CPFs válidos para os usuários de desenvolvimento novos | Gerar com o pacote `cpf` do projeto e validar com `cpf.Valid` | Os usuários de dev usam CPF real-formatado; inventar à mão falha na validação |

**Mapeamento do S4 (já feito):**
- `document_test.go:107` (contagens do seed) e `:134` (permissões do Revisor): vão mudar.
- `document_test.go:329` (`TestRevisorDoDepartamento`) e `document_files_test.go:185`: usam `revisor@dev.local` com permissões de **documentos**, que saem do Revisor.
- `document_test.go:166`: percorre os usuários de dev; o número de administradores (hoje 5) sobe.
- `project_test.go:96` e `project_hardening_test.go:85`: tentam rotas de equipe com um proponente alheio; não mudam.
- `project_lifecycle_test.go:24` e `:404`: usam perfis ad hoc para analista e secretário; **não mudam**, mas servem de modelo.

### T0.4 Alinhar a documentação
- `planejamento-cadastro-projeto.md`: renomeações de papel já feitas. Falta incluir a etapa `awaiting_opinion` no §5, o papel Secretário de Estado no §3.2/§9 (já existe no §3.2; conferir §9), e as novas fases no §11.
- Declarar o BPMN como fonte da verdade do fluxo (uma linha no §5 e no README).
**Pronto quando:** nenhum documento contradiz o plano geral.

### T0.5 Combinados de trabalho
- Branch: `feat/fluxo-analise-enquadramento`.
- Commits pequenos por passo (cada passo da fase 1 abaixo é um commit).
- Os arquivos novos (`docs/bpmn/`, os dois planejamentos) precisam entrar no controle de versão; hoje estão fora. Isso só acontece quando você pedir.

### T0.6 Revisão de risco antes da fase 1
Confirmar com os resultados dos spikes: (a) S2 já está decidido: a migration termina com o aviso de invalidação (P1.4, item 8); (b) S3 só confirma o comportamento da tela para o Analista.

### T0.7 Relatório da fase 0
Uma página (no próprio PR ou em `docs/`) com: linha de base, resultado do T0.2, respostas S1–S6, decisões tomadas a partir delas. A fase 1 só começa depois.

**Critério de saída da fase 0:** linha de base registrada; teste de visibilidade escrito e com resultado conhecido; S1–S6 respondidos; documentos coerentes.

---

## 3. Fase 1 — Perfis e visibilidade

Ordem pensada para cada passo ser verificável sozinho. Depois de cada um: `make test`, `make lint`; onde houver SQL, `make sqlc`.

### P1.1 Condições de visibilidade (sem mexer em banco)
- Onde: `backend/internal/project/members.go` (junto de `RegisterConditions`), com constantes `ConditionSubmitted` e `ConditionSubmittedSameDepartment` ao lado de `ConditionMember`/`ConditionEditor`.
- Comportamento: ver seção 1. `Category: "default"` para a primeira e `"custom"` para a segunda, como no restante. Os nomes aparecem sozinhos na tela de perfis (o registro de condições alimenta a UI).
- O registro precisa do departamento do usuário: `Subject.DepartmentID` já existe (usado por `admin::same-department`).
- Testes unitários: o filtro gerado em cada caso (status válido/inválido, usuário sem departamento nega, departamento diferente).
**Pronto quando:** testes unitários verdes; as condições aparecem no endpoint de condições.

### P1.2 Marcador `is_finalistic` — banco e consultas
- Migration `00024` (parte 1): `ALTER TABLE departments ADD COLUMN is_finalistic boolean NOT NULL DEFAULT false`.
- `backend/internal/database/queries/departments.sql`: incluir a coluna em `ListDepartments`, `ListDepartmentsWithCounts`, `GetDepartmentWithCounts`, `CreateDepartment`, `UpdateDepartment`; `UpsertDepartment` mantém `ON CONFLICT ... SET name = EXCLUDED.name` (não toca no marcador) e ganha o parâmetro só para o primeiro INSERT. Nova consulta de lista com filtro `finalistic`.
- `make sqlc` e commitar o código gerado.

### P1.3 Marcador — serviço, API e OpenAPI
- `internal/department/service.go`: `View` ganha `IsFinalistic`; `Create` e `Update` passam a receber um objeto de entrada (hoje recebem só o nome). Na criação o padrão é `false`; no `PUT`, campo **omitido = inalterado** (para um cliente antigo não desmarcar sem querer). A auditoria já grava antes/depois e passa a incluir o campo.
- `internal/department/handler.go`: corpo com `isFinalistic` opcional; `GET /admin/departments?finalistic=true`.
- `backend/api/openapi.yaml` (linhas ~485–555): campo novo nos schemas e o parâmetro de consulta; há teste que falha se algo ficar sem documentação.
- Regra anotada para a fase 4: desmarcar um departamento não afeta rodadas de enquadramento já abertas, só novas indicações.
- Testes (`department_test.go`): criar com e sem marcador; `PUT` sem o campo mantém; filtro; auditoria registra a mudança.

### P1.4 Migration 00024 — perfis e permissões (parte 2)
No mesmo arquivo da parte 1, idempotente como as 00022/00023 (`NOT EXISTS` / `ON CONFLICT DO NOTHING`).

1. **Perfil novo** `state-secretary` ("Secretário de Estado", `is_system = true`).
2. **Permissões do Secretário:** `admin::projects.read` e `admin::projects.deny`, ambas com a condição `admin::project-submitted`; leituras de entidades, municípios, parlamentares e parâmetros de regra (iguais às dos outros perfis de equipe).
3. **Analista:** `admin::settings.departments.read`, `.update` e `.delete` (decisão do cliente: ele lê, lista, altera e exclui departamentos; **criar não foi pedido e fica de fora**); as três ações de projeto que já tem passam a ter a condição `admin::project-submitted`. Excluir continua barrado pela regra existente `department_in_use`.
4. **Supervisor Administrador:** remover `request-complement`, `archive` e `deny`; manter `read` com a condição.
5. **Revisor do Departamento:** remover as quatro permissões `admin::content.document.*`; conceder `admin::projects.read` com `admin::project-submitted-same-department` e as leituras de catálogo e parâmetros (paridade com os demais). `review-framing` entra na fase 4.
6. **Descrições** dos perfis de sistema atualizadas, mas **só onde a descrição ainda é a original** (`WHERE description = '<texto antigo>'`), para não apagar edição feita por administradores.
7. **Mudança de condição das permissões existentes:** só onde a lista ainda é `[]` (`WHERE conditions = '[]'::jsonb`), preservando personalização.
8. **Invalidação de cache:** a migration termina com `SELECT pg_notify('authz_invalidate', '[]')` (no `Up` e no `Down`). O aviso é entregue quando a transação da migration confirma e invalida todos os perfis nas instâncias que usam `PGBroadcaster`. Com instância única (`LocalBroadcaster`) a migration roda antes de a API subir, então o cache nasce vazio; se for aplicada com a API ligada, reiniciar. Teste: com duas instâncias de `authz.Engine` sobre o mesmo banco, aplicar a mudança e conferir que a decisão da segunda muda.
9. **`Down`:** remove a coluna, devolve ao Supervisor as três ações e ao Revisor as quatro de documentos, volta as condições para `[]` e remove o perfil Secretário só se ninguém o usa (o mesmo cuidado da 00022).

Decisão de segurança: a condição `admin::project-submitted` vai em **todas** as ações de projeto da equipe (não só `read`), para uma tentativa de devolver ou arquivar um rascunho por id também ser barrada pela permissão (hoje a regra de estado do serviço já devolve 409; a condição passa a barrar antes, com 403).

**Pronto quando:** migration sobe e desce em banco limpo e em banco com dados.

### P1.5 Seed
Arquivo: `backend/internal/seed/seed.go`.
- Constante de perfil `RoleSecretary = "state-secretary"` e especificação com `system: true`, no padrão de `RoleSupervisor`.
- Especificações de Supervisor, Analista e Revisor espelhando a migration (a migration é a fonte da verdade para perfis que já existem, porque o seed só acrescenta o que falta e nunca remove; ver `ensureAdminRole`). Manter os dois sincronizados e testar a igualdade.
- `Departments` deixa de ser lista de nomes e passa a ter o marcador; `Financeiro` e `Jurídico` saem marcados. O marcador é aplicado **só na criação** (P1.2), para um novo seed não desfazer a escolha de um administrador.
- Usuários de desenvolvimento novos (`DevUsers`): `analista@dev.local`, `supervisor@dev.local`, `secretario@dev.local` e `revisor.juridico@dev.local` (departamento Jurídico); `revisor@dev.local` (Financeiro) continua. CPFs válidos gerados no S6.

### P1.6 Testes que mudam por causa dos perfis
- `document_test.go`: contagens do seed. **Confirmado ao rodar:** 8 perfis, 54 permissões (Secretário +6, Supervisor −3, Revisor +1, Analista +3: as duas de departamento a mais que o previsto) e 9 administradores. A primeira estimativa (52) esquecia `update` e `delete` de departamentos do Analista.
- Permissões do Revisor (`:134`): nova expectativa (leituras de projeto e catálogo).
- `TestRevisorDoDepartamento` e `document_files_test.go:185`: reescrever com um **perfil de teste** (`mkRole` + `grant`) que concede ações de documento com `same-department`, para a condição continuar coberta sem depender do Revisor real. Mover o teste do Revisor para as rotas de projeto (P1.7).
- Segunda execução do seed continua sem criar nada (idempotência).

### P1.7 Matriz de acesso por perfil e rota
Arquivo: `backend/internal/app/project_staff_roles_test.go` (a partir do teste do T0.2). Perfis **semeados**; para o Revisor, o departamento do projeto é definido por SQL direto nos testes (ainda não existe rota que o grave; ela vem na fase 4).

| Perfil | Lista | `GET` enviado | `GET` rascunho | Devolver / arquivar / negar |
|---|---|---|---|---|
| Proponente alheio | vazia | 404 | 404 | 403 |
| Analista | só enviados e arquivados | 200 | 404 | devolver e arquivar 200 (enviado), 403 (rascunho); negar 403 |
| Supervisor | só enviados e arquivados | 200 | 404 | 403 em todos |
| Secretário | só enviados e arquivados | 200 | 404 | negar 200 (enviado), 403 (rascunho); demais 403 |
| Revisor, departamento igual | só os do seu departamento | 200 | 404 | 403 em todos |
| Revisor, outro departamento | vazia | 404 | 404 | 403 |
| Revisor, sem departamento | **403** (a condição o nega por completo e o guard da rota responde antes de listar) | 403 | 403 | 403 |
| Super Admin | todos | 200 | 200 | conforme regra de estado |

As mesmas linhas valem para `documents`, `download`, `receipt`, `declarations` e `members` (leitura). Acrescentar:
- Projeto arquivado: a equipe vê (somente leitura).
- Pessoa com os dois perfis (Proponente + Analista) ainda vê o próprio rascunho.
- Departamentos: Analista lista, lê, edita (inclusive o marcador `isFinalistic`) e exclui (200; excluir departamento em uso devolve 409 `department_in_use`), mas não cria (403); Supervisor, Secretário e Revisor continuam 403; Proponente continua sem acesso. Toda alteração feita pelo Analista fica na auditoria com o autor.
- Permissões personalizadas sem as condições: documentar (não testar) que reabrem a exposição.

### P1.8 Frontend mínimo
- `features/departments`: `types.ts` (`isFinalistic`), `service.ts`, `queries.ts`; `department-dialog.tsx` com a opção "Área finalística" (seguir `frontend/docs/forms.md`: `useAppForm` + Zod, campo em `@/components/forms/fields`); `departments-list.tsx` com a coluna/indicador.
- `config/nav-config.ts`: **sem mudança**. O item "Departamentos" já exige `admin::settings.departments.read`, e o Analista passa a ter essa permissão, então o menu aparece para ele de propósito. Em `departments-list.tsx` e `department-dialog.tsx`, garantir com `can(...)` que o botão "Novo departamento" fica escondido para quem não tem `create`, e que editar e excluir aparecem para quem tem `update` e `delete`. A resposta do S3 confirma se a rota precisa de guarda própria.
- `bun run typecheck`, `lint`, `format:check`.
- Não há tela nova de equipe aqui; isso é a fase 5.

### P1.9 Documentação
- `backend/README.md` (linha ~99): perfis e usuários de desenvolvimento.
- `README.md`: se descreve perfis.
- `openapi.yaml`: já tratado no P1.3.
- Em `planejamento-cadastro-projeto.md` §9: tabela de permissões coerente com a matriz.
- Aviso em linguagem simples para administradores: **perfil personalizado precisa incluir as condições de visibilidade**.

### P1.10 Passo de implantação
Depois de aplicar a migration em produção, **marcar os departamentos finalísticos reais** na tela de departamentos antes de liberar o fluxo da fase 4. Registrar isso na descrição do PR.

---

## 4. Ordem, commits e dependências

```
T0.1 ─┬─ T0.2 ─┐
      ├─ T0.3 ─┼─ T0.6 ─ T0.7 ──► P1.1 ─► P1.2 ─► P1.3 ─┐
      └─ T0.4  ┘                                        ├─► P1.4 ─► P1.5 ─► P1.6 ─► P1.7 ─► P1.8 ─► P1.9
                                                        │
                                      (P1.1 também alimenta P1.4)
```

Um commit por passo, na ordem. P1.8 (frontend) pode andar em paralelo a P1.5–P1.7 depois do P1.3, porque só depende do contrato da API.

**Divisão em pull requests (sugestão):**
1. Fase 0: teste de visibilidade + relatório + documentos.
2. Fase 1A: P1.1 + P1.4 (parte de perfis) + P1.5 + P1.6 + P1.7 — fecha a exposição de rascunhos e entrega o Secretário.
3. Fase 1B: P1.2 + P1.3 + P1.8 — marcador de departamento.

A 1A não depende da 1B, exceto pelas permissões de departamento do Analista (`read`, `update`, `delete`), que não precisam do marcador para funcionar. Se for útil entregar a exposição de rascunhos o quanto antes, a 1A pode sair primeiro.

---

## 5. Riscos específicos destas fases

| Risco | Cuidado |
|---|---|
| Migration altera permissões e instâncias já em execução continuam com a versão antiga no cache | S2 define o procedimento; testar com duas instâncias se o cache tiver TTL longo |
| Retirar ações do Supervisor sobrescreve personalização de administradores | Só remover as três ações e só do perfil de sistema; comunicar antes de aplicar |
| Revisor sem departamento fica sem acesso a qualquer projeto (403, sem lista vazia) | Teste explícito na matriz (`TestMatrizDeLeituraDaEquipe`); aviso na tela de usuários fica para a fase 5 |
| O seed passa a divergir da migration | Teste que compara as permissões do banco após migration com as do seed |
| Condições novas mal escritas dão OU onde deveria ser E | Teste unitário de cada filtro (P1.1) e a matriz (P1.7) com perfis semeados |
| O Analista altera e exclui departamentos, e o marcador `is_finalistic` decide para onde os projetos podem ser indicados | Decisão do cliente. Auditoria já registra antes/depois; teste que o Analista não cria; avisar que desmarcar um departamento só afeta indicações novas |
| Marcador de departamento é sobrescrito pelo seed | Aplicado só no INSERT (P1.2) e coberto por teste de segunda execução |
| Reescrever os testes de documentos esconde uma regressão | O perfil de teste concede exatamente as ações antigas do Revisor, e os asserts originais se mantêm |

**Reversão:** a migration tem `Down` completo; o código novo da fase 1 não é usado por nada da fase 0; reverter o PR e o `Down` devolve o estado anterior.

---

## 6. Definição de pronto

**Fase 0**
- [ ] Linha de base registrada.
- [ ] Teste de visibilidade escrito e resultado anotado.
- [ ] S1–S6 respondidos.
- [ ] Documentos coerentes entre si e com o BPMN.

**Fase 1**
- [ ] Equipe não lê nem lista rascunho em nenhuma rota; Super Admin continua vendo.
- [ ] Perfis: Secretário criado; Supervisor somente leitura; Revisor sem permissões de documentos e com leitura de projeto restrita ao departamento; Analista lê, edita e exclui departamentos (não cria); a migration invalida o cache de permissões.
- [ ] Marcador `is_finalistic` gravável e filtrável; seed e telas atualizados.
- [ ] Migration sobe e desce; seed idempotente e igual à migration.
- [ ] `make test`, `make lint`, `make cover-authz` (≥ 80%), `make sqlc` sem diferenças.
- [ ] `bun run typecheck`, `lint`, `format:check`.
- [ ] OpenAPI cobre as rotas (teste existente).
- [ ] README e planejamentos atualizados; passo de implantação descrito no PR.
