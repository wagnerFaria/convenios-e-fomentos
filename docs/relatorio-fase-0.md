# Relatório da Fase 0 — Preparação

Data: 30/10/2026 · Branch: `feat/fluxo-analise-enquadramento` · Plano: [planejamento-fases-0-e-1.md](planejamento-fases-0-e-1.md)

## Resumo

1. **O achado de segurança se confirmou.** Os perfis **Analista** e **Supervisor Administrador** listam e leem **rascunhos** de outros usuários, inclusive baixam os arquivos anexados. A Fase 1 corrige isso.
2. A linha de base do backend está **toda verde**. No frontend há falhas **anteriores** a este trabalho (1 erro de lint, 12 avisos e 5 arquivos fora do formato).
3. Os seis spikes (S1 a S6) estão respondidos. Nenhum muda o desenho da Fase 1; dois ajustam o roteiro de verificação.
4. `golangci-lint` v2.15.0 foi instalado em `~/go/bin` (compilado com Go 1.27, a versão que o módulo exige). O `make lint` roda e acusa **1 problema anterior a este trabalho** (seção 1).

---

## 1. T0.1 — Linha de base (antes de qualquer mudança)

| Verificação | Resultado |
|---|---|
| `go test -race -count=1 ./...` | **Verde** (exit 0). O pacote `internal/app` leva ~206 s |
| `make cover-authz` | **91,8 %** (mínimo exigido: 80 %) |
| `sqlc generate` | Sem diferenças no código gerado |
| `go vet ./internal/app/` | Verde (com o teste novo) |
| `make lint` (`golangci-lint` v2.15.0) | **1 problema anterior**: `internal/project/declarations.go:172`, retorno de `fmt.Fprintf` não verificado (`errcheck`). O teste novo não gera nenhum |
| Frontend `typecheck` | **Verde** |
| Frontend `lint` (oxlint) | **1 erro e 12 avisos, anteriores a este trabalho** |
| Frontend `format:check` (oxfmt) | **5 arquivos fora do formato, anteriores a este trabalho** |

As verificações do frontend foram executadas **dentro do contêiner `web`** (`docker compose -f docker-compose.dev.yml exec web bun run ...`), porque `frontend/node_modules` no host é um ponto de montagem vazio, de propriedade do root; as dependências vivem num volume do contêiner.

**Falhas pré-existentes do frontend** (para as próximas fases não serem culpadas por elas):

- Erro: `react(set-state-in-effect)` em `src/components/ui/date-input.tsx:80`.
- Avisos: `src/features/documents/components/document-files.tsx:107`, `src/components/ui/calendar.tsx` (4), `src/components/ui/date-input.tsx:45`, `src/components/kbar/render-result.tsx:18`, `src/components/data-table/data-table-slider-filter.tsx` (2), entre outros.
- Fora do formato: `data-table-date-filter.tsx`, `data-table-pagination.tsx`, `forms/fields/date-picker-field.tsx`, `forms/fields/text-field.tsx`, `ui/breadcrumb.tsx`.

**Consequência para a Fase 1:** o critério de pronto do frontend passa a ser "**nenhum erro, aviso ou diferença de formato novos**" nos arquivos que mudarmos, e não "zero".

---

## 2. T0.2 — Teste de visibilidade da equipe

Arquivo: [backend/internal/app/project_staff_visibility_test.go](../backend/internal/app/project_staff_visibility_test.go).

**Cenário:** a Maria tem um rascunho completo (com um arquivo anexado) e um projeto enviado. Dois administradores com os perfis **semeados** (`analyst` e `administrative-supervisor`). O teste inclui controles: o titular alcança todas as rotas dos dois projetos, e a equipe vê o projeto enviado com o mesmo resultado do titular.

**Resultado: FALHA, como previsto.** Para os dois perfis:

| Rota | Esperado | Obtido |
|---|---|---|
| `GET /api/admin/projects` (lista) | sem o rascunho | **lista o rascunho** |
| `GET /projects/{id}` (detalhe) | 404 | **200** |
| `GET /projects/{id}/documents` (checklist) | 404 | **200** |
| `GET /projects/{id}/documents/files/{fileId}/download` | 404 | **200** |
| `GET /projects/{id}/declarations/{code}` | 404 | **200** |

Os controles passaram, então a falha é do comportamento do sistema e não do teste. (Uma primeira versão do teste falhou por usar o ajudante `do`, que lê JSON, na rota do recibo, que devolve PDF; corrigido com `raw`, que lê só o código HTTP.)

**Gravidade.** O projeto em rascunho contém dados pessoais (contatos, endereço, dados de quem assina as declarações) e os arquivos anexados de pessoas que ainda nem enviaram o projeto. Qualquer Analista ou Supervisor vê tudo isso (o CPF aparece mascarado para a equipe, mas o restante não). Os perfis vêm das migrations 00022 e 00023, então a exposição existe em qualquer ambiente que as tenha aplicado. Vale tratar como falha de privacidade (LGPD) e priorizar a correção da Fase 1.

**Estado do teste:** fica **vermelho** na branch de propósito. Não deve entrar na `main` antes da correção; ele passa a verde na Fase 1 (passos P1.1 e P1.4).

---

## 3. T0.3 — Spikes

| # | Pergunta | Resposta |
|---|---|---|
| S1 | O filtro por `status` e `department_id` funciona na listagem SQL e na checagem em memória? | **Sim.** O `Schema` do projeto vem das tags de `project.Ref` (`id`, `status`, `department_id`, `created_by_id`), então os quatro campos são filtráveis. `Filter.ToSQL` converte `And`, `Or`, `Eq` e `In` em SQL parametrizado, e `Filter.Match` (usado em `AllowsRecord`) normaliza ponteiros e UUIDs; valor nulo nunca bate. Falta só o teste unitário das condições novas (P1.1) |
| S2 | O cache de permissões enxerga mudanças feitas por migration? | **Não sozinho.** O cache (`authz/cache.go`) não tem prazo de validade; só é limpo por `invalidate`. O `PGBroadcaster` escuta o canal `authz_invalidate`, e uma lista vazia (`[]`) invalida todos os perfis. **Decisão:** a migration 00024 termina com `pg_notify('authz_invalidate', '[]')` (no `Up` e no `Down`), confirmada por teste com duas instâncias do motor (P1.4, item 8) |
| S3 | A tela de departamentos funciona para o Analista (ler, editar, excluir; sem criar)? | **Sim, sem mudança estrutural.** A página usa `hasPermission(departmentActions.read)`; em `departments-list.tsx`, editar e excluir já dependem de `can(update)` e `can(delete)`; em `department-dialog.tsx`, o botão de criar já some sem `can(create)`. O menu continua exigindo `departments.read`. Só é preciso acrescentar a opção "Área finalística" |
| S4 | Que testes dependem dos perfis que mudam? | Mapeados e confirmados na linha de base (ver abaixo) |
| S5 | Como as migrations são testadas? | Já existe **`TestMigracoesEConsultas`** (`internal/database/database_test.go`): sobe tudo, repete o `Up`, desfaz **a última migração** e sobe de novo. (Correção posterior: eu havia escrito "desce tudo"; `Migrator.Down` desfaz uma única migração. Cada migração nova desloca qual `Down` o teste exercita, então ele passou a descer até a 00024 sair.) **Atenção:** ele chama `q.UpsertDepartment(ctx, "Financeiro")`; se a assinatura mudar no P1.2, esse teste precisa ser ajustado. Um teste **com dados** (subir até a 00023, popular, subir a 00024) continua desejável |
| S6 | CPFs válidos para os usuários de desenvolvimento | Gerados com o mesmo algoritmo do pacote `cpf` (conferido contra os 4 CPFs que o seed já usa): analista `55566677720`, supervisor `66677788830`, secretário `77788899941`, revisor do Jurídico `88899900078`. Falta conferir que nenhum já é usado em testes (feito no P1.5) |

**S4, dependências de teste (confirmado):**
- `document_test.go:107` (contagens do seed: hoje `[7, 47, 2, 5, 1, 11, 2, 2]`) e `:134` (permissões do Revisor).
- `document_test.go:329` (`TestRevisorDoDepartamento`) e `document_files_test.go:185`: usam `revisor@dev.local` com permissões de documentos.
- `document_test.go:166`: percorre os usuários de desenvolvimento.
- `database_test.go`: `UpsertDepartment` (novo, ver S5).
- `project_lifecycle_test.go:24` e `:404`: perfis ad hoc para analista e secretário; não mudam.

---

## 4. T0.4 — Documentação alinhada

Em [planejamento-cadastro-projeto.md](planejamento-cadastro-projeto.md):
- Cabeçalho: o BPMN e o plano de fluxo são a fonte da verdade do fluxo da equipe.
- §3.2 e §9: papéis atualizados (Analista, Supervisor Administrador, Revisor do Departamento, Secretário de Estado).
- §5: transições novas (assumir análise, indicar departamento, confirmar e recusar enquadramento) e a etapa "Aguardando parecer".
- §11: fase 8 aponta para os novos planos.

O `README.md` não foi alterado: já tinha mudanças suas não commitadas, e o cabeçalho do planejamento cobre a referência.

---

## 5. Ajustes ao plano da Fase 1 decorrentes da Fase 0

1. **Lint do backend:** resolvido. A ferramenta foi instalada. Para o critério "`make lint` verde" valer, o primeiro commit da Fase 1 corrige o problema de `declarations.go:172` (um `_, _ =` no `Fprintf` sobre um `hash.Hash`, que nunca devolve erro), isolado dos demais.
2. **Critério do frontend:** "nenhum problema novo" (seção 1), não "zero".
3. **`UpsertDepartment`:** ajustar `database_test.go` no P1.2.
4. **Teste de visibilidade:** já existe e está vermelho. Os passos P1.1 e P1.4 o deixam verde; o P1.7 o amplia para a matriz completa.
5. **Ambiente de desenvolvimento ligado:** a API roda com `air` (recompila ao salvar), então mudanças em código de produção da Fase 1 reiniciam a API do contêiner. Antes de aplicar a migration 00024 nesse banco, confirmar como o `docker compose` executa as migrações (a verificar no início da Fase 1).

## 6. Critério de saída da Fase 0

- [x] Linha de base registrada.
- [x] Teste de visibilidade escrito, com resultado conhecido (falha confirmada).
- [x] S1 a S6 respondidos.
- [x] Documentos coerentes entre si e com o BPMN.
- [x] `golangci-lint` instalado e linha de base do lint registrada.
